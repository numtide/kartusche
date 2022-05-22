package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/draganm/kartusche/runtime"
	"github.com/gorilla/mux"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type kartusche struct {
	Hosts   []string `json:"host"`
	Prefix  string   `json:"prefix"`
	Error   string   `json:"error,omitempty"`
	name    string
	runtime runtime.Runtime
	path    string
}

func (k *kartusche) start() error {
	rt, err := runtime.Open(k.path)
	if err != nil {
		k.Error = err.Error()
		return fmt.Errorf("while starting: %w", err)
	}
	k.runtime = rt
	return nil
}

func (k *kartusche) delete() error {
	var err error
	if k.runtime != nil {
		err = multierr.Append(err, k.runtime.Shutdown())
	}

	err = multierr.Append(err, os.Remove(k.path))
	return err
}

type server struct {
	db            bolted.Database
	mu            *sync.Mutex
	kartusches    map[string]*kartusche
	kartuschesDir string
	tempDir       string

	router *mux.Router
	log    *zap.SugaredLogger
}

func createIfNotExisting(dir string, perm os.FileMode) error {
	s, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, perm)
		if err != nil {
			return fmt.Errorf("while creating %s: %w", dir, err)
		}
		return nil
	}
	if !s.IsDir() {
		return fmt.Errorf("%s is not a dir", dir)
	}
	return nil
}

func open(path string, log *zap.SugaredLogger) (*server, error) {
	err := createIfNotExisting(path, 0700)
	if err != nil {
		return nil, err
	}

	kartuschesDir := filepath.Join(path, "kartusches")
	err = createIfNotExisting(kartuschesDir, 0700)
	if err != nil {
		return nil, err
	}

	tempDir := filepath.Join(path, "tmp")
	err = createIfNotExisting(tempDir, 0700)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(path, "state")
	db, err := embedded.Open(dbPath, 0700)
	if err != nil {
		return nil, fmt.Errorf("while opening state db: %w", err)
	}

	hasKartuscheMap := false

	err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {
		hasKartuscheMap = tx.Exists(kartuschesPath)
		return nil
	})

	if !hasKartuscheMap {
		err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
			tx.CreateMap(kartuschesPath)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("while initializing db: %w", err)
		}
	}

	s := &server{
		db:            db,
		kartusches:    map[string]*kartusche{},
		kartuschesDir: kartuschesDir,
		tempDir:       tempDir,
		mu:            new(sync.Mutex),
		router:        mux.NewRouter(),
		log:           log,
	}

	go s.runtimeManager()

	return s, nil

}

var kartuschesPath = dbpath.ToPath("kartusches")
