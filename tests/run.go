package tests

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	rt "runtime"
	"strings"

	_ "embed"

	"github.com/cucumber/godog"
	"github.com/dop251/goja"
	"github.com/draganm/kartusche/runtime"
	"github.com/draganm/kartusche/tests/wsclient"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

//go:embed expect.js
var expectSource string

func nArgFuncType(n int) reflect.Type {
	switch n {
	case 0:
		return reflect.TypeOf(func() error { return nil })
	case 1:
		return reflect.TypeOf(func(x string) error { return nil })
	case 2:
		return reflect.TypeOf(func(x, z string) error { return nil })
	default:
		panic(fmt.Errorf("%d number of args not supported", n))
	}
}

func Run(dir string) error {
	programs := []*goja.Program{}
	ps, err := goja.Compile(`step.js`, `
		function step(matcher, fn) {
			__step(matcher, fn.length, fn)
		}
		world = {}
	`, false)

	if err != nil {
		return fmt.Errorf("while compiling step.js: %w", err)
	}

	exp, err := goja.Compile("expect.js", expectSource, false)

	if err != nil {
		return fmt.Errorf("while compiling expect.js: %w", err)
	}

	programs = append(programs, ps, exp)

	td, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("while creating temp file: %w", err)
	}

	defer os.RemoveAll(td)

	err = filepath.WalkDir("tests/support", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			ext := filepath.Ext(d.Name())
			if ext == ".js" {

				d, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("while reading %s: %w", path, err)
				}
				p, err := goja.Compile(path, string(d), false)
				if err != nil {
					return fmt.Errorf("while compiling %s: %w", path, err)
				}
				programs = append(programs, p)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("while compiling support files: %w", err)
	}

	masterKartuscheFile := filepath.Join(td, "kartusche")

	err = runtime.InitializeNew(masterKartuscheFile, ".")
	if err != nil {
		return fmt.Errorf("while initializing Kartusche: %w", err)
	}

	status := godog.TestSuite{
		Name: "kartusche",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {

			kartuscheFile, err := os.CreateTemp(td, "kartusche-*")
			if err != nil {
				panic(fmt.Errorf("while creating temp kartusche file: %w", err))
			}

			mkf, err := os.Open(masterKartuscheFile)
			if err != nil {
				panic(fmt.Errorf("while opening master kartusche file: %w", err))
			}

			defer mkf.Close()

			_, err = io.Copy(kartuscheFile, mkf)
			if err != nil {
				panic(fmt.Errorf("while copying from master kartusche: %w", err))
			}

			err = kartuscheFile.Close()
			if err != nil {
				panic(fmt.Errorf("while closing kartusche file: %w", err))
			}

			kartusche, err := runtime.Open(kartuscheFile.Name(), zap.NewNop().Sugar())
			if err != nil {
				panic(fmt.Errorf("while opening Kartusche: %w", err))
			}

			server := httptest.NewServer(kartusche)

			sc.After(func(ctx context.Context, sc *godog.Scenario, scenarioError error) (context.Context, error) {
				server.CloseClientConnections()
				server.Close()
				err := kartusche.Shutdown()
				if err != nil {
					return ctx, fmt.Errorf("while shutting down Kartusche: %w", err)
				}

				return ctx, nil
			})

			vm := goja.New()
			// TODO record open responses and close them after the scenario

			vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
			vm.Set("println", fmt.Println)
			vm.Set("apiCall", func(method, path, body string, headers map[string]string) (*http.Response, error) {
				su, err := url.Parse(server.URL)
				if err != nil {
					return nil, fmt.Errorf("while parsing server url: %w", err)
				}
				su.Path = path
				req, err := http.NewRequest(method, su.String(), strings.NewReader(body))
				if err != nil {
					return nil, fmt.Errorf("while creating new request: %w", err)
				}

				for k, v := range headers {
					req.Header.Set(k, v)
				}

				res, err := server.Client().Do(req)
				if err != nil {
					return nil, err
				}

				return res, nil
			})

			vm.Set("connectWebsocket", func(path string) (*wsclient.WSClient, error) {
				u, err := url.Parse(server.URL)
				if err != nil {
					return nil, err
				}

				u.Scheme = "ws"
				u.Path = path
				conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					return nil, err
				}

				return &wsclient.WSClient{Conn: conn}, nil

			})

			vm.Set("readToString", func(r io.Reader) (string, error) {
				d, err := io.ReadAll(r)
				if err != nil {
					return "", fmt.Errorf("while reading stream: %w", err)
				}

				return string(d), nil
			})
			vm.Set("__step", func(matcher string, argCount int, fn func(args ...interface{}) error) {

				matchingFnValue := reflect.MakeFunc(nArgFuncType(argCount), func(args []reflect.Value) (results []reflect.Value) {
					defer func() {
						p := recover()
						if p != nil {
							var ok bool
							err, ok = p.(error)
							if ok {
								results = []reflect.Value{reflect.ValueOf(&err).Elem()}
								return
							}
							panic(p)
						}
					}()
					argsValues := []interface{}{}
					for _, v := range args {
						argsValues = append(argsValues, v.Interface())
					}
					err := fn(argsValues...)
					return []reflect.Value{reflect.ValueOf(&err).Elem()}
				})
				sc.Step(matcher, matchingFnValue.Interface())
			})
			for _, p := range programs {
				_, err := vm.RunProgram(p)
				if err != nil {
					panic(fmt.Errorf("while executing js: %w", err))
				}
			}

		},
		Options: &godog.Options{
			Format:              "kartusche",
			Paths:               []string{"tests/features"},
			NoColors:            true,
			Strict:              true,
			StopOnFailure:       true,
			ShowStepDefinitions: false,
			Concurrency:         rt.NumCPU(),
		},
	}.Run()

	if status != 0 {
		return fmt.Errorf("tests failed")
	}

	return nil

}
