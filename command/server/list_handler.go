package server

import (
	"encoding/json"
	"net/http"
)

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		handleHttpError(w, err, s.log)
	}()

	kl := []string{}

	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.kartusches {
		kl = append(kl, k)
	}

	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(kl)

}
