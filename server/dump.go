package server

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"net/http"
	"sort"

	"github.com/draganm/bolted"
)

func (s Server) dumpHandler(w http.ResponseWriter, r *http.Request) {
	gzw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		err = fmt.Errorf("could not create gzip writer: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.log.Error(err, "could not backup")
		return
	}

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	s.mu.Lock()

	type dumpWriter struct {
		name    string
		applier func(func(tx bolted.SugaredReadTx) error) error
	}

	dumpWriters := []dumpWriter{
		dumpWriter{
			"__server",
			func(fn func(bolted.SugaredReadTx) error) error {
				return bolted.SugaredRead(s.db, fn)
			},
		},
	}

	for name, k := range s.kartusches {
		dumpWriters = append(dumpWriters, dumpWriter{name: name, applier: k.runtime.Read})
	}

	s.mu.Unlock()

	sort.Slice(dumpWriters, func(i, j int) bool {
		return dumpWriters[i].name < dumpWriters[j].name
	})

	w.Header().Set("Trailer", "X-KartuscheDumpComplete")

	for _, dw := range dumpWriters {
		err := dw.applier(func(tx bolted.SugaredReadTx) error {
			err := tw.WriteHeader(&tar.Header{
				Name:     dw.name,
				Typeflag: tar.TypeReg,
				Size:     tx.FileSize(),
			})

			if err != nil {
				return fmt.Errorf("could not write dump header for %s: %w", dw.name, err)
			}

			tx.Dump(tw)
			return nil
		})
		if err != nil {
			s.log.Error(err, "could not write dump")
			return
		}
	}
	err = tw.Close()
	if err != nil {
		s.log.Error(err, "could not close dump tar writer")
	}

	w.Header().Set("X-KartuscheDumpComplete", "true")

}
