package upload

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/common/serverurl"
	"github.com/draganm/kartusche/config"
	"github.com/draganm/kartusche/runtime"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "upload",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "kartusche-server-base-url",
			EnvVars: []string{"KARTUSCHE_SERVER_BASE_URL"},
		},
		&cli.StringFlag{
			Name: "name",
		},
		&cli.StringSliceFlag{
			Name:    "hostname",
			EnvVars: []string{"HOSTNAMES"},
		},
		&cli.StringFlag{
			Name:    "prefix",
			Value:   "/",
			EnvVars: []string{"PREFIX"},
		},
	},
	Action: func(c *cli.Context) (err error) {
		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while uploading Kartusche: %w", err), 1)
			}
		}()

		dir := c.Args().First()

		if dir == "" {
			dir = "."
		}

		td, err := os.MkdirTemp("", "")
		if err != nil {
			return fmt.Errorf("while creating temp dir: %w", err)
		}

		defer os.Remove(td)

		kartuscheFileName := filepath.Join(td, "kartusche")

		err = runtime.InitializeNew(kartuscheFileName, dir)
		if err != nil {
			return fmt.Errorf("while initializing Kartusche: %w", err)
		}

		cfg, err := config.Current()
		if err != nil {
			return err
		}

		serverBaseURL, err := serverurl.BaseServerURL(c.Args().First())
		if err != nil {
			return err
		}

		kf, err := os.Open(kartuscheFileName)
		if err != nil {
			return err
		}

		defer kf.Close()

		err = client.CallAPI(serverBaseURL, "PUT", path.Join("kartusches", cfg.Name), nil, func() (io.Reader, error) { return kf, nil }, nil, 204)
		if err != nil {
			return err
		}

		return nil
	},
}
