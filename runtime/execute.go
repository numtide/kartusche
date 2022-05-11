package runtime

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

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

func InitializeNew(fileName string, pathPrefix string, r *tar.Reader) (err error) {
	db, err := embedded.Open(fileName, 0700)
	if err != nil {
		return fmt.Errorf("while opening database: %w", err)
	}
	defer db.Close()

	wtx, err := db.BeginWrite()
	if err != nil {
		return fmt.Errorf("while starting write transaction: %w", err)
	}

	defer func() {
		if err != nil {
			wtx.Rollback()
		} else {
			err = wtx.Finish()
		}
	}()

	initScript := ""
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("while reading header: %w", err)
		}

		cleanPath := path.Clean(h.Name)

		dp, err := dbpath.Parse(cleanPath)
		if err != nil {
			return fmt.Errorf("while parsing dbpath: %w", err)
		}

		if h.Typeflag == tar.TypeDir {
			err = wtx.CreateMap(dp)
			if err != nil {
				return err
			}
		}

		if h.Typeflag == tar.TypeReg {
			d, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("while reading entry %s: %w", cleanPath, err)
			}

			if cleanPath == "init.js" {
				initScript = string(d)
			}

			err = wtx.Put(dp, d)
			if err != nil {
				return err
			}
		}

	}

	initScriptProgram, err := goja.Compile("init.js", initScript, false)
	if err != nil {
		return fmt.Errorf("while parsing init: %w", err)
	}

	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	vm.Set("tx", wtx)
	vm.Set("uuidv4", uuid.NewV4)
	vm.Set("uuidv7", func() (string, error) {
		id, err := uuid.NewV7(uuid.NanosecondPrecision)
		if err != nil {
			return "", err
		}
		return id.String(), nil
	})

	_, err = vm.RunProgram(initScriptProgram)
	if err != nil {
		return fmt.Errorf("while running init script: %w", err)
	}

	return nil
}

func Open(fileName string, pathPrefix string) (Runtime, error) {
	db, err := embedded.Open(fileName, 0700)
	if err != nil {
		return nil, fmt.Errorf("while opening database: %w", err)
	}

	r := mux.NewRouter()

	err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {
		apiPath := dbpath.ToPath("api")
		if !tx.Exists(apiPath) {
			return nil
		}
		toDo := []dbpath.Path{apiPath}

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
						vm.Set("vars", vars)
						vm.Set("r", r)
						vm.Set("w", w)
						vm.Set("read", reader(db))
						vm.Set("write", writer(db))
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
