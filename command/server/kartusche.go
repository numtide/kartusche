package server

import (
	"fmt"
	"os"

	"github.com/draganm/kartusche/runtime"
	"go.uber.org/multierr"
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
