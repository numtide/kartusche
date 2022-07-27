package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) infoDBStats(w http.ResponseWriter, r *http.Request) {
	var err error

	var name = mux.Vars(r)["name"]

	defer func() {
		handleHttpError(w, err, s.log)
	}()

	s.mu.Lock()
	k, ok := s.kartusches[name]
	s.mu.Unlock()

	if !ok {
		err = newErrorWithCode(errors.New("not found"), 404)
		return
	}

	rt := k.runtime
	if rt == nil {
		err = newErrorWithCode(errors.New("kartusche is not running"), 409)
		return
	}

	w.Header().Set("content-type", "application/json")

	dbs, err := rt.GetDBStats()
	if err != nil {
		return
	}

	json.NewEncoder(w).Encode(dbs)

}
