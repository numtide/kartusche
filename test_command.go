package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	_ "embed"

	"github.com/cucumber/godog"
	"github.com/dop251/goja"
	"github.com/draganm/kartusche/runtime"
	"github.com/urfave/cli/v2"
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

var testCommand = &cli.Command{
	Name: "test",
	Action: func(ctx *cli.Context) error {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("while getting CWD: %w", err)
		}

		kartuscheName := filepath.Base(currentDir)
		programs := []*goja.Program{}
		ps, err := goja.Compile(`step.js`, `
			function step(matcher, fn) {
				__step(matcher, fn.length, fn)
			}
			world = {}
		`, false)

		if err != nil {
			panic(err)
		}

		exp, err := goja.Compile("expect.js", expectSource, false)

		if err != nil {
			panic(err)
		}

		tarBytes, err := tarDirToByteSlice("content")
		if err != nil {
			return fmt.Errorf("while tarring content: %w", err)
		}

		programs = append(programs, ps, exp)
		status := godog.TestSuite{
			Name: kartuscheName,
			TestSuiteInitializer: func(tsc *godog.TestSuiteContext) {
				err := filepath.WalkDir("tests/support", func(path string, d fs.DirEntry, err error) error {
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
					panic(err)
				}
			},
			ScenarioInitializer: func(sc *godog.ScenarioContext) {

				td, err := os.MkdirTemp("", "")
				if err != nil {
					panic(fmt.Errorf("while creating temp file: %w", err))
				}

				kartuscheFile := filepath.Join(td, "kartusche")

				err = runtime.InitializeNew(kartuscheFile, "/", tar.NewReader(bytes.NewReader(tarBytes)))
				if err != nil {
					panic(fmt.Errorf("while initializing Kartusche: %w", err))
				}

				kartusche, err := runtime.Open(kartuscheFile, "/")
				if err != nil {
					panic(fmt.Errorf("while opening Kartusche: %w", err))
				}

				sc.After(func(ctx context.Context, sc *godog.Scenario, scenarioError error) (context.Context, error) {
					err := kartusche.Shutdown()
					if err != nil {
						return ctx, fmt.Errorf("while shutting down Kartusche: %w", err)
					}

					err = os.RemoveAll(td)
					if err != nil {
						return ctx, fmt.Errorf("while removing Kartusche temp dir: %w", err)
					}

					return ctx, nil
				})

				vm := goja.New()
				vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
				vm.Set("println", fmt.Println)
				vm.Set("apiCall", func(method, path, body string, headers map[string]string) (*http.Response, error) {
					rec := httptest.NewRecorder()
					req, err := http.NewRequest(method, path, strings.NewReader(body))
					if err != nil {
						return nil, fmt.Errorf("while creating new request: %w", err)
					}

					kartusche.ServeHTTP(rec, req)

					return rec.Result(), nil
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
				Format:              "progress",
				Paths:               []string{"tests/features"},
				NoColors:            true,
				Strict:              true,
				StopOnFailure:       true,
				ShowStepDefinitions: false,
			},
		}.Run()

		return cli.Exit("godog", status)
	},
}

func tarDirToByteSlice(dir string) ([]byte, error) {
	bb := new(bytes.Buffer)
	tw := tar.NewWriter(bb)

	// walk through every file in the folder
	err := filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {

		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		pathParts := strings.Split(file, string(os.PathSeparator))
		pathParts[0] = "."

		header.Name = filepath.ToSlash(strings.Join(pathParts, string(os.PathSeparator)))

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("while creating tar: %w", err)
	}

	err = tw.Close()
	if err != nil {
		return nil, fmt.Errorf("while finishing tar: %w", err)
	}

	return bb.Bytes(), nil
}
