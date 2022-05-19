package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/draganm/kartusche/runtime"
)

type kartusche struct {
	Hosts  []string `json:"host"`
	Prefix string   `json:"prefix"`
	name   string

	runtime runtime.Runtime
	path    string
}

func (k *kartusche) start() error {
	rt, err := runtime.Open(k.path, k.Prefix)
	if err != nil {
		return fmt.Errorf("while starting: %w", err)
	}
	k.runtime = rt
	return nil
}

type server struct {
	db            bolted.Database
	mu            *sync.Mutex
	kartusches    map[string]*kartusche
	kartuschesDir string
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

func open(path string) (*server, error) {
	err := createIfNotExisting(path, 0700)
	if err != nil {
		return nil, err
	}

	kartuschesDir := filepath.Join(path, "kartusches")
	err = createIfNotExisting(kartuschesDir, 0700)
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
		hasKartuscheMap = tx.Exists(kartuschePath)
		return nil
	})

	if !hasKartuscheMap {
		err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
			tx.CreateMap(kartuschePath)
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
		mu:            new(sync.Mutex),
	}

	err = s.start()
	if err != nil {
		return nil, fmt.Errorf("while starting kartusches: %w", err)
	}

	return s, nil

}

var kartuschePath = dbpath.ToPath("kartusche")

func (s *server) start() error {

	kartusches := []*kartusche{}

	err := bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) error {
		for it := tx.Iterator(kartuschePath); !it.IsDone(); it.Next() {
			k := kartusche{}
			err := json.Unmarshal(it.GetValue(), &k)
			if err != nil {
				return err
			}
			k.name = it.GetKey()
			k.path = filepath.Join(s.kartuschesDir, k.name)

			kartusches = append(kartusches, &k)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("while reading kartusche data: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, k := range kartusches {
		err = k.start()
		if err != nil {
			return fmt.Errorf("while starting kartusche %s: %w", k.name, err)
		}
		s.kartusches[k.name] = k
	}

	return nil

}
