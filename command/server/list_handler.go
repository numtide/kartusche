package server

import (
	"encoding/json"
	"net/http"
)

type KartuscheListEntry struct {
	Name   string   `json:"name"`
	Hosts  []string `json:"hosts"`
	Prefix string   `json:"prefix"`
}

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		handleHttpError(w, err)
	}()

	kl := []KartuscheListEntry{}

	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.kartusches {
		kl = append(kl, KartuscheListEntry{
			Name:   k,
			Hosts:  v.Hosts,
			Prefix: v.Prefix,
		})
	}

	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(kl)

}
