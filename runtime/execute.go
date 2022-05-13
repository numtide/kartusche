package runtime

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	_ "embed"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/draganm/kartusche/runtime/jslib"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
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

// type HandlerMeta struct {
// 	Method string `json:"method"`
// 	Path   string `json:"path"`
// }

type Runtime interface {
	http.Handler
	Shutdown() error
}

type runtime struct {
	db bolted.Database
	*mux.Router
}

func (r *runtime) Shutdown() error {
	return r.db.Close()
}

func Open(fileName string, pathPrefix string) (Runtime, error) {
	db, err := embedded.Open(fileName, 0700)
	if err != nil {
		return nil, fmt.Errorf("while opening database: %w", err)
	}

	r := mux.NewRouter()

	err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {

		jslib, err := jslib.Load(tx)
		if err != nil {
			return err
		}

		err = addStaticHandlers(r, tx, db)
		if err != nil {
			return fmt.Errorf("while adding static handlers: %w", err)
		}

		handlersPath := dbpath.ToPath("handler")
		if !tx.Exists(handlersPath) {
			return nil
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
						return fmt.Errorf("while compiling %s: %w", current.Append(key).String(), err)
					}

					prefixRoute := r.PathPrefix(pathPrefix)

					prefixRoute.Methods(method).Path("/" + path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						vars := mux.Vars(r)
						vm := goja.New()
						vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
						vm.Set("require", jslib.Require(vm))
						vm.Set("vars", vars)
						vm.Set("r", r)
						vm.Set("w", w)
						vm.Set("read", reader(db))
						vm.Set("write", writer(db))
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

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("while starting runtime: %w", err)
	}

	return &runtime{db: db, Router: r}, nil

}
