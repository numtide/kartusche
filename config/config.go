package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ErrConfigNotFound = errors.New("please make sure you're in a sub-directory of a kartusche")

type Config struct {
	Name          string            `yaml:"name"`
	DefaultRemote string            `yaml:"default_remote"`
	Remotes       map[string]string `yaml:"remotes"`
}

func (c *Config) GetServerBaseURL(remoteName string) (string, error) {

	if remoteName == "" {
		remoteName = c.DefaultRemote
	}

	if remoteName == "" {
		return "", errors.New("no remotes configured")
	}

	u := c.Remotes[remoteName]

	if u == "" {
		return "", fmt.Errorf("could not find remote url for %s", remoteName)
	}

	return u, nil
}

func Current() (*Config, error) {

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("while getting current dir: %w", err)
	}

	for {
		pth := filepath.Join(currentDir, ".kartusche", "config.yaml")
		_, err = os.Stat(pth)
		if os.IsNotExist(err) {
			parent := filepath.Dir(currentDir)
			if currentDir == parent {
				return nil, ErrConfigNotFound
			}
			currentDir = filepath.Dir(currentDir)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("while getting stat of %s: %w", pth, err)
		}

		c, err := loadConfig(filepath.Join(".kartusche", "config.yaml"))
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	return nil, ErrConfigNotFound

}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, ErrConfigNotFound
	}

	if err != nil {
		return nil, err
	}

	defer f.Close()

	c := &Config{}
	err = yaml.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, fmt.Errorf("while decoding %s: %w", path, err)
	}

	if c.Remotes == nil {
		c.Remotes = make(map[string]string)
	}

	return c, nil
}

func (c *Config) Write(dir string) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("while writing config: %w", err)
		}
	}()

	configDir := filepath.Join(dir, ".kartusche")

	_, err = os.Stat(configDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(configDir, 0700)
	}

	if err != nil {
		return err
	}

	pth := filepath.Join(configDir, "config.yaml")
	f, err := os.OpenFile(pth, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return fmt.Errorf("while opening kartusche config for write: %w", err)
	}

	defer f.Close()

	err = yaml.NewEncoder(f).Encode(c)
	if err != nil {
		return fmt.Errorf("while writing kartusche config: %w", err)
	}

	return nil
}
