package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/draganm/kartusche/command/server/verifier"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type server struct {
	db            bolted.Database
	mu            *sync.Mutex
	kartusches    map[string]*kartusche
	kartuschesDir string
	tempDir       string
	domain        string

	router   *mux.Router
	log      *zap.SugaredLogger
	verifier verifier.AuthenticationProvider
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

func open(path string, domain string, verifier verifier.AuthenticationProvider, log *zap.SugaredLogger) (*server, error) {
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

	initialPaths := []dbpath.Path{
		kartuschesPath,
		authPath,
		openTokenRequests,
		tokensPath,
		usersPath,
	}

	pathsToCreate := []dbpath.Path{}

	err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {
		for _, p := range initialPaths {
			if !tx.Exists(p) {
				pathsToCreate = append(pathsToCreate, p)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("while checking if initial path exists: %w", err)
	}

	if len(pathsToCreate) != 0 {
		err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
			for _, p := range pathsToCreate {
				tx.CreateMap(p)
			}
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
		verifier:      verifier,
		domain:        domain,
	}

	go s.runtimeManager()

	return s, nil

}

var kartuschesPath = dbpath.ToPath("kartusches")
var authPath = dbpath.ToPath("auth")
var openTokenRequests = authPath.Append("open_token_requests")
var tokensPath = authPath.Append("tokens")
var usersPath = authPath.Append("users")
