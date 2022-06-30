package rm

import (
	"errors"
	"fmt"
	"path"

	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "rm",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "kartusche-server-base-url",
			EnvVars: []string{"KARTUSCHE_SERVER_BASE_URL"},
		},
	},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while listing Kartusches: %w", err), 1)
			}
		}()

		name := c.Args().First()
		if name == "" {
			return errors.New("name of the Kartusche must be provided")
		}

		cfg, err := config.Current()
		if err != nil {
			return fmt.Errorf("while getting current config: %w", err)
		}

		serverBaseURL := cfg.GetServerBaseURL(c.String("kartusche-server-base-url"))

		err = client.CallAPI(serverBaseURL, "DELETE", path.Join("kartusches", name), nil, nil, nil, 204)
		if err != nil {
			return err
		}

		return nil

	},
}
