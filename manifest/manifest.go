package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type Static struct {
	Dir string `yaml:"dir"`
	// StaticBuildCommand *StaticBuildCommand `yaml:"build_command"`
}

// type StaticBuildCommand struct {
// 	Cmd  string            `yaml:"cmd"`
// 	Args []string          `yaml:"args"`
// 	Env  map[string]string `yaml:"env"`
// }

type KartuscheManifest struct {
	Name   string  `yaml:"name"`
	Static *Static `yaml:"static"`
	dir    string
}

func (k *KartuscheManifest) StaticDir() (string, error) {
	if k.Static != nil && k.Static.Dir != "" {
		if filepath.IsAbs(k.Static.Dir) {
			return k.Static.Dir, nil
		}

		return filepath.Join(k.dir, k.Static.Dir), nil

	}
	return filepath.Join(k.dir, "static"), nil
}

func Load(path string) (*KartuscheManifest, error) {
	kf, err := os.Open(filepath.Join(path, "kartusche.yaml"))
	if err != nil {
		return nil, fmt.Errorf("while opening kartusche.yaml: %w", err)
	}

	defer kf.Close()

	dec := yaml.NewDecoder(kf)
	km := &KartuscheManifest{}
	err = dec.Decode(km)
	if err != nil {
		return nil, fmt.Errorf("while parsing kartusche.yaml: %w", err)
	}

	return km, nil
}
