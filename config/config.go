package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerBaseURL string `yaml:"server_base_url"`
}

func (c *Config) GetServerBaseURL(cliURL string) string {
	if cliURL != "" {
		return cliURL
	}

	if c.ServerBaseURL != "" {
		return c.ServerBaseURL
	}

	return "http://localhost:3003"
}

func (c *Config) merge(nx *Config) {

	if nx == nil {
		return
	}

	if c.ServerBaseURL == "" {
		c.ServerBaseURL = nx.ServerBaseURL
	}

}

func Current() (*Config, error) {

	c := &Config{}

	cd, err := loadConfig(filepath.Join(".kartusche", "config.yaml"))
	if err != nil {
		return nil, err
	}
	c.merge(cd)

	hd, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("while getting user's home dir: %w", err)
	}

	ch, err := loadConfig(filepath.Join(hd, ".kartusche", "config.yaml"))
	if err != nil {
		return nil, err
	}

	c.merge(ch)

	return c, nil

}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	defer f.Close()

	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, fmt.Errorf("while decoding %s: %w", path, err)
	}

	return c, nil
}
