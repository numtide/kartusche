package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("auth token not found")

func GetTokenForServer(server string) (string, error) {
	hd, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("could not get user's home dir: %w", err)
	}

	configPath := filepath.Join(hd, ".kartusche", "auth.yaml")
	d, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("while getting auth token for %s: %w", server, ErrNotFound)
	}

	tokens := map[string]string{}
	err = yaml.Unmarshal(d, tokens)

	if err != nil {
		return "", fmt.Errorf("could not unmarshal %s: %w", configPath, err)
	}

	t, found := tokens[server]

	if !found {
		return "", ErrNotFound
	}

	return t, nil

}

func StoreTokenForServer(server, token string) error {
	hd, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("could not get user's home dir: %w", err)
	}

	configPath := filepath.Join(hd, ".kartusche", "auth.yaml")
	d, err := os.ReadFile(configPath)

	tokens := map[string]string{}

	fileNotFound := os.IsNotExist(err)

	if err != nil && !fileNotFound {
		return fmt.Errorf("while reading %s: %w", configPath, err)
	}

	if err == nil {
		err = yaml.Unmarshal(d, tokens)

		if err != nil {
			return fmt.Errorf("could not unmarshal %s: %w", configPath, err)
		}
	}

	if fileNotFound {
		err = os.MkdirAll(filepath.Dir(configPath), 0700)
		if err != nil {
			return fmt.Errorf("while creating dir %s: %w", filepath.Dir(configPath), err)
		}

	}

	tokens[server] = token

	nc, err := yaml.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("while marshalling new auth config: %w", err)
	}

	err = os.WriteFile(configPath, nc, 0700)
	if err != nil {
		return fmt.Errorf("while writing %s: %w", configPath, err)
	}

	return nil

}
