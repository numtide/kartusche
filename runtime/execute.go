package runtime

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"strings"
	"sync"

	_ "embed"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/draganm/kartusche/runtime/dbwrapper"
	"github.com/draganm/kartusche/runtime/httprequest"
	"github.com/draganm/kartusche/runtime/jslib"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

//go:embed stdlib.js
var stdlibSource string

var compiledStdlib *goja.Program

func init() {
	var err error
	compiledStdlib, err = goja.Compile("stdlib.js", stdlibSource, false)
	if err != nil {
		panic(fmt.Errorf("while compiling stdlib: %w", err))
	}
}

type Runtime interface {
	http.Handler
	Shutdown() error
	Update(func(tx bolted.SugaredWriteTx) error) error
}

type runtime struct {
	db bolted.Database
	r  *mux.Router
	mu *sync.Mutex
}

func (r *runtime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	rt := r.r
	r.mu.Unlock()

	rt.ServeHTTP(w, req)
}

func (r *runtime) Shutdown() error {
	return r.db.Close()
}

func (r *runtime) Update(fn func(tx bolted.SugaredWriteTx) error) error {
	var rt *mux.Router
	err := bolted.SugaredWrite(r.db, func(tx bolted.SugaredWriteTx) error {
		err := fn(tx)
		if err != nil {
			return err
		}
		rt, err = initializeRouter(tx, dbwrapper.New(r.db))
		return err
	})

	if err != nil {
		return err
	}

	r.mu.Lock()
	r.r = rt
	r.mu.Unlock()
	return nil
}

func initializeRouter(tx bolted.SugaredReadTx, dbw *dbwrapper.DB) (*mux.Router, error) {
	r := mux.NewRouter()

	jslib, err := jslib.Load(tx)
	if err != nil {
		return nil, err
	}

	err = addStaticHandlers(r, tx)
	if err != nil {
		return nil, fmt.Errorf("while adding static handlers: %w", err)
	}

	handlersPath := dbpath.ToPath("handler")
	if !tx.Exists(handlersPath) {
		return r, nil
	}
	toDo := []dbpath.Path{handlersPath}

	for len(toDo) > 0 {
		current := toDo[0]
		toDo = toDo[1:]
		for it := tx.Iterator(current); !it.IsDone(); it.Next() {
			key := it.GetKey()
			fullPath := current.Append(key)
			if tx.IsMap(fullPath) {
				toDo = append(toDo, fullPath)
				continue
			}

			if strings.HasSuffix(key, ".js") {
				path := path.Join([]string(current[1:])...)
				method := strings.TrimSuffix(key, ".js")

				src := it.GetValue()

				program, err := goja.Compile(current.Append(key).String(), string(src), false)
				if err != nil {
					return nil, fmt.Errorf("while compiling %s: %w", current.Append(key).String(), err)
				}

				r.Methods(method).Path("/" + path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					vars := mux.Vars(r)
					vm := goja.New()
					vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
					vm.Set("require", jslib.Require(vm))
					vm.Set("vars", vars)
					vm.Set("r", r)
					vm.Set("w", w)
					vm.Set("println", fmt.Println)
					vm.Set("read", dbw.Read)
					vm.Set("write", dbw.Write)
					vm.Set("http_do", httprequest.Request)
					vm.Set("watch", func(path []string, fn func(interface{}) (bool, error)) selectable {
						os, _ := dbw.Watch(path, fn)
						return os
					})
					_, err = vm.RunProgram(compiledStdlib)
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
					vm.Set("uuidv4", func() (string, error) {
						id, err := uuid.NewV4()
						if err != nil {
							return "", err
						}
						return id.String(), nil

					})
					vm.Set("uuidv7", func() (string, error) {
						id, err := uuid.NewV7(uuid.NanosecondPrecision)
						if err != nil {
							return "", err
						}
						return id.String(), nil
					})
					vm.Set("requestBody", func() (string, error) {
						d, err := io.ReadAll(r.Body)
						if err != nil {
							return "", fmt.Errorf("while reading request body: %w", err)
						}

						return string(d), nil
					})

					vm.Set("select", func(selectables ...selectable) (err error) {

						// reflect.SelectCase
						cases := make([]reflect.SelectCase, len(selectables))
						for i, s := range selectables {
							cases[i] = reflect.SelectCase{
								Dir:  reflect.SelectRecv,
								Chan: s.SelectChan(),
							}
						}
						for {
							chosen, val, ok := reflect.Select(cases)
							if !ok {
								// TODO - return something else?
								return nil
							}
							done, err := selectables[chosen].Fn()(val)
							if err != nil {
								continue
							}
							if done {
								return nil
							}
						}

					})

					vm.Set("upgradeToWebsocket", func(handler func(interface{}) (bool, error)) (selectable, error) {
						upgrader := websocket.Upgrader{
							ReadBufferSize:  1024,
							WriteBufferSize: 1024,
						}

						conn, err := upgrader.Upgrade(w, r, nil)
						if err != nil {
							return nil, err
						}

						ch := make(chan interface{}, 1)

						go func() {
							defer conn.Close()
							defer close(ch)

							for {
								var v interface{}
								err = conn.ReadJSON(&v)
								if err != nil {
									return
								}
								ch <- v
							}

						}()

						vm.Set("wsSend", func(msg interface{}) error {
							return conn.WriteJSON(msg)
						})

						return &defaultSelectable{ch: ch, fn: handler}, nil

					})

					ctx := r.Context()

					go func() {
						<-ctx.Done()
						e := ctx.Err()
						if e != nil {
							vm.Interrupt(ctx.Err())
						}
					}()

					_, err := vm.RunProgram(program)
					if err != nil {
						fmt.Println(err)
						http.Error(w, err.Error(), 500)
						return
					}
				})

			}
		}

	}

	return r, nil

}

func Open(fileName string) (Runtime, error) {
	db, err := embedded.Open(fileName, 0700)
	if err != nil {
		return nil, fmt.Errorf("while opening database: %w", err)
	}

	var r *mux.Router

	err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {
		r, err = initializeRouter(tx, dbwrapper.New(db))
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("while starting runtime: %w", err)
	}

	return &runtime{db: db, r: r, mu: new(sync.Mutex)}, nil

}
