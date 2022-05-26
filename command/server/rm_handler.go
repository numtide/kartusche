package server

import (
	"errors"
	"net/http"

	"github.com/draganm/bolted"
	"github.com/gorilla/mux"
)

func (s *server) rm(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		handleHttpError(w, err, s.log)
	}()

	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		err = newErrorWithCode(errors.New("name not provided"), 400)
		return
	}

	err = bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		toDeletePath := kartuschesPath.Append(name)
		if !tx.Exists(toDeletePath) {
			return newErrorWithCode(errors.New("not found"), 404)
		}
		tx.Delete(toDeletePath)
		return nil
	})

	if err != nil {
		return
	}

	w.WriteHeader(204)

}
