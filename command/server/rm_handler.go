package server

import (
	"errors"
	"net/http"
	"os"

	"github.com/draganm/bolted"
	"github.com/gorilla/mux"
)

func (s *server) rm(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		handleHttpError(w, err)
	}()

	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		err = newErrorWithCode(errors.New("name not provided"), 400)
		return
	}

	err = bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		toDeletePath := kartuschePath.Append(name)
		if !tx.Exists(toDeletePath) {
			return newErrorWithCode(errors.New("not found"), 404)
		}
		tx.Delete(toDeletePath)
		return nil
	})

	if err != nil {
		return
	}

	s.mu.Lock()
	k := s.kartusches[name]
	delete(s.kartusches, name)
	s.mu.Unlock()

	defer os.Remove(k.path)

	err = k.runtime.Shutdown()
	if err != nil {
		return
	}

	w.WriteHeader(204)

}
