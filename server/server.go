package server

import (
	"net/http"
	"sync"

	"github.com/draganm/kartusche/runtime"
	"github.com/gorilla/mux"
)

type Server struct {
	http.Handler
	dataDir    string
	mu         *sync.Mutex
	kartusches map[string]runtime.Runtime
}

func Start(dataDir string) (*Server, error) {
	r := mux.NewRouter()
	s := &Server{
		Handler: r,
		dataDir: dataDir,
		mu:      &sync.Mutex{},
	}

	r.Path("__kartusche__/{name}").Methods("POST").HandlerFunc(s.create)

	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	s.mu.Lock()
	_, exists := s.kartusches[name]
	s.mu.Unlock()

	if exists {
		http.Error(w, "already exists", http.StatusConflict)
	}

}

func (s *Server) Stop() error {
	return nil
}
