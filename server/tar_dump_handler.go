package server

import (
	"archive/tar"
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/kartusche/common/util/path"
	"github.com/gorilla/mux"
)

func (s *Server) tarDump(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("content-type", "application/x-tar")

	err = rt.Read(func(tx bolted.SugaredReadTx) error {

		tw := tar.NewWriter(w)
		toDo := []dbpath.Path{dbpath.NilPath}

		for len(toDo) > 0 {
			current := toDo[0]
			toDo = toDo[1:]
			for it := tx.Iterator(current); !it.IsDone(); it.Next() {
				sp := current.Append(it.GetKey())
				h := &tar.Header{
					Name:    filepath.ToSlash(path.DBPathToFilePath(sp)),
					ModTime: time.Now(),
					Mode:    0700,
				}
				isMap := tx.IsMap(sp)
				if isMap {
					toDo = append(toDo, sp)
					h.Typeflag = tar.TypeDir
					h.Name += "/"
					err := tw.WriteHeader(h)
					if err != nil {
						return err
					}
					continue
				} else {
					d := tx.Get(sp)
					h.Typeflag = tar.TypeReg
					h.Size = int64(len(d))
					err := tw.WriteHeader(h)
					if err != nil {
						return err
					}
					_, err = tw.Write(d)
					if err != nil {
						return err
					}

				}

			}

		}
		err := tw.Close()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.log.Error(err, "while dumping tar")
	}

}
