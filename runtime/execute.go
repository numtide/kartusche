package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

type HandlerMeta struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

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

func New(fileName string, pathPrefix string) (Runtime, error) {
	db, err := embedded.Open(fileName, 0700)
	if err != nil {
		return nil, fmt.Errorf("while opening database: %w", err)
	}

	r := mux.NewRouter()

	err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {
		handlersPath := dbpath.ToPath("__handlers__")
		for it := tx.Iterator(handlersPath); !it.IsDone(); it.Next() {
			handlerName := it.GetKey()
			handlerPath := handlersPath.Append(handlerName)
			metadataPath := handlerPath.Append("meta_data")
			mdd := tx.Get(metadataPath)
			md := HandlerMeta{}
			err = json.Unmarshal(mdd, &md)
			if err != nil {
				return fmt.Errorf("while unmarshalling %s: %w", metadataPath, err)
			}

			sourcePath := handlerPath.Append("source")
			src := string(tx.Get(sourcePath))
			program, err := goja.Compile("handler.js", src, false)
			if err != nil {
				return fmt.Errorf("while compiling %s: %w", sourcePath, err)
			}

			prefixRoute := r.PathPrefix(pathPrefix)

			prefixRoute.Methods(md.Method).Path(md.Path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)
				vm := goja.New()
				vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
				vm.Set("vars", vars)
				vm.Set("r", r)
				vm.Set("w", w)
				vm.Set("read", reader(db))
				vm.Set("write", writer(db))
				vm.Set("uuidv4", uuid.NewV4)
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
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("while starting runtime: %w", err)
	}

	return &runtime{db: db, Router: r}, nil

}
