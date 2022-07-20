package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/gorilla/mux"
)

var handlerPath = dbpath.ToPath("handler")

type HandlerInfo struct {
	Verb   string `json:"method"`
	Path   string `json:"path"`
	Source string `json:"source"`
}

func (s *Server) infoHandlers(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		handleHttpError(w, err, s.log)
	}()

	v := mux.Vars(r)
	name := v["name"]

	if name == "" {
		err = newErrorWithCode(errors.New("kartusche not found"), 400)
		return
	}

	s.mu.Lock()
	k, found := s.kartusches[name]
	s.mu.Unlock()

	if !found {
		err = newErrorWithCode(errors.New("kartusche not found"), 400)
		return
	}

	handlers := []HandlerInfo{}

	err = k.runtime.Read(func(tx bolted.SugaredReadTx) error {

		if !tx.Exists(handlerPath) {
			return nil
		}

		toDo := []dbpath.Path{handlerPath}

		for len(toDo) > 0 {
			head := toDo[0]
			toDo = toDo[1:]
			if tx.IsMap(head) {
				for it := tx.Iterator(head); !it.IsDone(); it.Next() {
					pth := head.Append(it.GetKey())
					toDo = append(toDo, pth)
				}
				continue
			}

			verb := strings.TrimSuffix(head[len(head)-1], ".js")
			handlers = append(handlers, HandlerInfo{
				Verb:   verb,
				Path:   "/" + head[1:len(head)-1].String(),
				Source: string(tx.Get(head)),
			})
		}
		return nil
	})

	if err != nil {
		return
	}

	sort.Slice(handlers, func(i, j int) bool {
		a, b := handlers[i], handlers[j]
		if a.Path == b.Path {
			return a.Verb < b.Verb
		}
		return a.Path < b.Path
	})

	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(handlers)

}
