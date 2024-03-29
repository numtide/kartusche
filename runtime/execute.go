package runtime

import (
	"context"
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
	"github.com/draganm/kartusche/runtime/cronjobs"
	"github.com/draganm/kartusche/runtime/dbwrapper"
	"github.com/draganm/kartusche/runtime/jobs"
	"github.com/draganm/kartusche/runtime/jslib"
	"github.com/draganm/kartusche/runtime/stdlib"
	"github.com/draganm/kartusche/runtime/template"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/robfig/cron/v3"
)

const maxJobHistorySize = 100

type Runtime interface {
	http.Handler
	Shutdown() error
	Update(func(tx bolted.SugaredWriteTx) error) error

	// Write won't reload any handlers, please don't
	// update any code, static content and templates
	// with this!
	Write(func(tx bolted.SugaredWriteTx) error) error
	Read(func(tx bolted.SugaredReadTx) error) error
	GetDBStats() (*DBStats, error)
}

type runtime struct {
	db     bolted.Database
	r      *mux.Router
	mu     *sync.Mutex
	cron   *cron.Cron
	logger logr.Logger
	ctx    context.Context
	cancel func()
}

func (r *runtime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	rt := r.r
	r.mu.Unlock()

	rt.ServeHTTP(w, req)
}

func (r *runtime) GetDBStats() (*DBStats, error) {
	dbs, err := r.db.Stats()
	if err != nil {
		return nil, fmt.Errorf("while getting dbstats: %w", err)
	}

	return &DBStats{
		FreePageN:     dbs.FreePageN,
		PendingPageN:  dbs.PendingPageN,
		FreeAlloc:     dbs.FreeAlloc,
		FreelistInuse: dbs.FreelistInuse,
		TxN:           dbs.TxN,
		OpenTxN:       dbs.OpenTxN,
		TxStats: TxStats{
			PageCount:     dbs.TxStats.PageCount,
			PageAlloc:     dbs.TxStats.PageAlloc,
			CursorCount:   dbs.TxStats.CursorCount,
			NodeCount:     dbs.TxStats.NodeCount,
			NodeDeref:     dbs.TxStats.NodeDeref,
			Rebalance:     dbs.TxStats.Rebalance,
			RebalanceTime: dbs.TxStats.RebalanceTime,
			Split:         dbs.TxStats.Split,
			Spill:         dbs.TxStats.Spill,
			SpillTime:     dbs.TxStats.SpillTime,
			Write:         dbs.TxStats.Write,
			WriteTime:     dbs.TxStats.WriteTime,
		},
	}, nil
}

func (r *runtime) Shutdown() error {
	ctx := r.cron.Stop()
	<-ctx.Done()
	return r.db.Close()
}

func (r *runtime) Read(fn func(tx bolted.SugaredReadTx) error) error {
	return bolted.SugaredRead(r.db, fn)
}

func (r *runtime) Write(fn func(tx bolted.SugaredWriteTx) error) error {
	return bolted.SugaredWrite(r.db, fn)
}

func (r *runtime) Update(fn func(tx bolted.SugaredWriteTx) error) error {
	var rt *mux.Router

	stCtx := r.cron.Stop()
	<-stCtx.Done()

	var cron *cron.Cron
	err := bolted.SugaredWrite(r.db, func(tx bolted.SugaredWriteTx) error {
		err := fn(tx)
		if err != nil {
			return err
		}

		err = runInit(tx, r.db, r.logger)
		if err != nil {
			return fmt.Errorf("while running init.js: %w", err)
		}

		jslib, err := jslib.Load(tx)
		if err != nil {
			return fmt.Errorf("while loading libs: %w", err)
		}

		rt, err = initializeRouter(tx, jslib, r.db, r.logger)
		if err != nil {
			return fmt.Errorf("while initializing router: %w", err)
		}

		cron, err = cronjobs.CreateCron(tx, jslib, r.db, r.logger)
		if err != nil {
			return fmt.Errorf("while initializing cron: %w", err)
		}
		return err
	})

	if err != nil {
		r.cron.Start()
		return err
	}

	r.mu.Lock()
	r.r = rt
	r.cron = cron
	r.mu.Unlock()
	r.cron.Start()

	return nil
}

func runInit(tx bolted.SugaredWriteTx, db bolted.Database, logger logr.Logger) (err error) {
	initPath := dbpath.ToPath("init.js")
	ex := tx.Exists(initPath)

	if ex {
		initScript := tx.Get(initPath)

		initScriptProgram, err := goja.Compile("init.js", string(initScript), false)
		if err != nil {
			return fmt.Errorf("while parsing init: %w", err)
		}

		vm := goja.New()
		lib, err := jslib.Load(tx)
		if err != nil {
			return fmt.Errorf("while loading jslib: %w", err)
		}

		stdlib.SetStandardLibMethods(vm, lib, db, dbpath.ToPath(), logger)
		vm.Set("tx", &dbwrapper.WriteTxWrapper{WriteTx: tx.GetRawWriteTX(), VM: vm})
		vm.GlobalObject().Delete("read")
		vm.GlobalObject().Delete("write")

		_, err = vm.RunProgram(initScriptProgram)
		if err != nil {
			return fmt.Errorf("while running init script: %w", err)
		}

	}

	return nil

}

func initializeRouter(tx bolted.SugaredReadTx, jslib *jslib.Libs, db bolted.Database, logger logr.Logger) (*mux.Router, error) {
	r := mux.NewRouter()
	r.StrictSlash(false)

	err := addStaticHandlers(r, tx)
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

				handlerFunc := func(w http.ResponseWriter, r *http.Request) {
					vars := mux.Vars(r)
					vm := goja.New()
					stdlib.SetStandardLibMethods(vm, jslib, db, current, logger)
					dbw := dbwrapper.New(db, vm, logger)

					vm.Set("vars", vars)
					vm.Set("r", r)
					vm.Set("w", w)
					vm.Set("render_template", template.RenderTemplate(db, current, w))
					vm.Set("watch", func(path []string, fn func(interface{}) (bool, error)) selectable {
						os, _ := dbw.Watch(path, fn)
						return os
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
							done, err := selectables[chosen].Fn()(val.Interface())
							if err != nil {
								logger.Error(err, "while running selectable")
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

						vm.Set("wsSendJson", func(msg interface{}) error {
							return conn.WriteJSON(msg)
						})

						vm.Set("wsSendHtml", func(msg string) error {
							return conn.WriteMessage(websocket.TextMessage, []byte(msg))
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
				}

				r.Methods(method).Path("/" + path).HandlerFunc(handlerFunc)
				r.Methods(method).Path("/" + path + "/").HandlerFunc(handlerFunc)
			}
		}

	}

	return r, nil

}

func Open(fileName string, logger logr.Logger) (Runtime, error) {
	db, err := embedded.Open(fileName, 0700, embedded.Options{})
	if err != nil {
		return nil, fmt.Errorf("while opening database: %w", err)
	}

	var r *mux.Router

	var cron *cron.Cron
	ctx, cancel := context.WithCancel(context.Background())

	err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {

		jslib, err := jslib.Load(tx)
		if err != nil {
			return fmt.Errorf("while loading libs: %w", err)
		}

		r, err = initializeRouter(tx, jslib, db, logger)
		if err != nil {
			return fmt.Errorf("while initializing router: %w", err)
		}

		cron, err = cronjobs.CreateCron(tx, jslib, db, logger)
		if err != nil {
			return fmt.Errorf("while initializing cron: %w", err)
		}

		go jobs.JobScheduler(ctx, db, maxJobHistorySize, jslib, logger)

		return err
	})

	if err != nil {
		cancel()
		return nil, fmt.Errorf("while starting runtime: %w", err)
	}

	cron.Start()

	return &runtime{
		db:     db,
		r:      r,
		mu:     new(sync.Mutex),
		logger: logger,
		cron:   cron,
		ctx:    ctx,
		cancel: cancel,
	}, nil

}
