package serverurl

import (
	"errors"
	"os"

	"github.com/draganm/kartusche/config"
)

func BaseServerURL(remoteName string) (string, error) {

	cfg, err := config.Current()
	if err != nil && !errors.Is(err, config.ErrConfigNotFound) {
		return "", err
	}

	if err == nil {
		return cfg.GetServerBaseURL(remoteName)
	}

	serverBaseURL := os.Getenv("KARTUSCHE_SERVER_BASE_URL")

	if serverBaseURL == "" {
		return "", errors.New("could not determinate remote server url")
	}

	return serverBaseURL, nil

}
