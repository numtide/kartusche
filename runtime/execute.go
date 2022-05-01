package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
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

	prefixRoute := r.PathPrefix(pathPrefix)

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

			prefixRoute.Path(md.Path).Methods(md.Method).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)
				vm := goja.New()
				vm.Set("vars", vars)
				vm.Set("r", r)
				vm.Set("w", w)
				_, err := vm.RunProgram(program)
				if err != nil {
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
