package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/draganm/bolted"
	"github.com/gorilla/mux"
)

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	rt := s.router
	s.mu.Unlock()
	rt.ServeHTTP(w, r)
}

func (s *server) runtimeManager() (err error) {
	log := s.log.With("process", "runtimeManager")
	defer func() {
		if err != nil {
			log.With("error", err).Error("manager exited")
		}
	}()

	changesChan, cancel := s.db.Observe(kartuschesPath.ToMatcher().AppendAnyElementMatcher())
	defer cancel()

	for range changesChan {
		toAdd := []*kartusche{}
		allFound := map[string]struct{}{}
		err = bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) error {
			for it := tx.Iterator(kartuschesPath); !it.IsDone(); it.Next() {

				name := it.GetKey()
				k := &kartusche{
					name: name,
					path: filepath.Join(s.kartuschesDir, name),
				}
				allFound[name] = struct{}{}
				err = json.Unmarshal(it.GetValue(), &k)
				if err != nil {
					return fmt.Errorf("while unmarshalling Kartusche %s: %w", name, err)
				}
				s.mu.Lock()
				_, found := s.kartusches[name]
				s.mu.Unlock()
				if !found {
					toAdd = append(toAdd, k)
				}
			}
			return nil
		})

		if err != nil {
			return err
		}

		toDelete := []string{}

		s.mu.Lock()
		for existing := range s.kartusches {
			_, found := allFound[existing]
			if !found {
				toDelete = append(toDelete, existing)
			}
		}
		s.mu.Unlock()

		for _, k := range toAdd {
			err = k.start()
			s.mu.Lock()
			s.kartusches[k.name] = k
			s.mu.Unlock()
		}

		for _, deleteName := range toDelete {
			s.mu.Lock()
			k := s.kartusches[deleteName]
			delete(s.kartusches, deleteName)
			s.mu.Unlock()
			if k != nil {
				err = k.delete()
				if err != nil {
					log.With("error", err, "kartusche", deleteName).Error("while deleting Kartusche")
				}
			} else {
				log.With("kartusche", deleteName).Warn("trying to delete not existing Kartusche")
			}

		}
		s.updateRouter()
	}

	return nil

}

func (s *server) updateRouter() {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := mux.NewRouter()
	for _, k := range s.kartusches {
		if k.runtime != nil {
			// TODO allow for any host
			r.Host(fmt.Sprintf("%s.%s", k.name, s.domain)).Handler(k.runtime)
		}
	}
	s.router = r
}
