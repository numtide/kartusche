package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/draganm/bolted"
	"github.com/gorilla/mux"
)

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
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

	tf, err := os.CreateTemp(s.tempDir, "")
	if err != nil {
		return
	}

	_, err = io.Copy(tf, r.Body)

	if err != nil {
		return
	}

	defer func() {
		tf.Close()
		os.Remove(tf.Name())
	}()

	err = tf.Close()
	if err != nil {
		return
	}

	kartuscheFilePath := filepath.Join(s.kartuschesDir, name)

	err = os.Rename(tf.Name(), kartuscheFilePath)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			os.Remove(kartuscheFilePath)
		}
	}()

	k := &kartusche{}

	kb, err := json.Marshal(k)
	if err != nil {
		return
	}

	err = bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		newKartuschePath := kartuschesPath.Append(name)
		if tx.Exists(newKartuschePath) {
			return newErrorWithCode(errors.New("already exists"), 419)
		}

		tx.Put(newKartuschePath, kb)
		return nil
	})

	if err != nil {
		return
	}

	w.WriteHeader(204)

}
