package runtime

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/gorilla/mux"
)

func addStaticHandlers(r *mux.Router, tx bolted.SugaredReadTx) error {
	staticPath := dbpath.ToPath("static")
	if !tx.Exists(staticPath) {
		return nil
	}

	toDo := []dbpath.Path{staticPath}

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

			fullPathWithoutStatic := []string(fullPath)[1:]
			requestPath := "/" + path.Join(fullPathWithoutStatic...)

			handler, err := staticContentHandler(fullPath, tx, requestPath)
			if err != nil {
				return fmt.Errorf("while creating static handler for %s: %w", requestPath, err)
			}
			r.Methods("GET").Path(requestPath).HandlerFunc(handler)

			if key == "index.html" {
				indexRequestPath := "/" + path.Join(fullPathWithoutStatic[:len(fullPathWithoutStatic)-1]...)
				r.Methods("GET").Path(indexRequestPath).HandlerFunc(handler)
			}

		}
	}

	return nil

}

func staticContentHandler(dbPath dbpath.Path, tx bolted.SugaredReadTx, name string) (http.HandlerFunc, error) {
	t := time.Now()

	d := tx.Get(dbPath)

	contentType := http.DetectContentType(d)

	sum := sha1.Sum(d)
	etag := fmt.Sprintf(`"%x"`, sum[:])

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", contentType)
		w.Header().Set("etag", etag)
		http.ServeContent(w, r, name, t, bytes.NewReader(d))
	}, nil
}
