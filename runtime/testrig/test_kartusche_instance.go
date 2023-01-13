package testrig

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/kartusche/runtime"
	"github.com/go-logr/logr"
)

type TestKartuscheInstance interface {
	GetURL() string
	GetRuntime() runtime.Runtime
	AddContent(pth, content string) error
}

type testKartuscheInstance struct {
	url     string
	runtime runtime.Runtime
}

func (tk *testKartuscheInstance) GetURL() string {
	return tk.url
}

func (tk *testKartuscheInstance) GetRuntime() runtime.Runtime {
	return tk.runtime
}

func (tk *testKartuscheInstance) AddContent(pth, content string) error {

	parts := strings.Split(pth, "/")
	parents := []dbpath.Path{
		dbpath.ToPath(parts[0]),
	}

	for i, p := range parts[1:] {
		parents = append(parents, parents[i].Append(p))
	}

	return tk.runtime.Update(func(tx bolted.SugaredWriteTx) error {
		for _, p := range parents[:len(parents)-1] {
			if !tx.Exists(p) {
				tx.CreateMap(p)
			}
		}
		fullPath := dbpath.ToPath(parts...)

		tx.Put(fullPath, []byte(content))
		return nil

	})
}

func StartTestKartuscheInstance(ctx context.Context) (TestKartuscheInstance, error) {

	td, err := os.MkdirTemp("", "")

	if err != nil {
		return nil, fmt.Errorf("could not create temp dir: %w", err)
	}

	dbpath := filepath.Join(td, "db")

	err = runtime.InitializeEmpty(dbpath)
	if err != nil {
		return nil, fmt.Errorf("could not initialize empty kartusche: %w", err)
	}

	rt, err := runtime.Open(dbpath, logr.Discard())
	if err != nil {
		return nil, fmt.Errorf("could not open new runtime: %w", err)
	}

	server := httptest.NewServer(rt)

	go func() {
		<-ctx.Done()
		server.Close()
		rt.Shutdown()
		os.RemoveAll(td)
	}()

	tki := &testKartuscheInstance{
		url:     server.URL,
		runtime: rt,
	}

	return tki, nil

}
