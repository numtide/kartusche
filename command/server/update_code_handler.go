package server

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/gorilla/mux"
)

func (s *server) updateCode(w http.ResponseWriter, r *http.Request) {
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

	err = rt.Update(func(tx bolted.SugaredWriteTx) error {
		// step one: delete everything apart from the data
		for it := tx.Iterator(dbpath.NilPath); !it.IsDone(); it.Next() {
			if it.GetKey() != "data" {
				tx.Delete(dbpath.ToPath(it.GetKey()))
			}
		}

		// step two: unpack the tar
		tr := tar.NewReader(r.Body)

		for {
			h, err := tr.Next()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return fmt.Errorf("while reading code tar: %w", err)
			}

			cleanPath := path.Clean(h.Name)
			dp, err := dbpath.Parse(cleanPath)
			if err != nil {
				return fmt.Errorf("while parsing dbpath: %w", err)

			}
			if h.Typeflag == tar.TypeDir {
				tx.CreateMap(dp)
			}
			if h.Typeflag == tar.TypeReg {
				d, err := io.ReadAll(tr)
				if err != nil {
					return fmt.Errorf("while reading entry %s: %w", cleanPath, err)
				}
				tx.Put(dp, d)
			}
		}

	})

	if err != nil {
		return
	}

	w.WriteHeader(204)

}
