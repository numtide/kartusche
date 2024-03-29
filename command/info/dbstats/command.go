package dbstats

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/common/serverurl"
	"github.com/draganm/kartusche/runtime"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var Command = &cli.Command{
	Name:  "dbstats",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while listing handlers: %w", err), 1)
			}
		}()

		serverBaseURL, err := serverurl.BaseServerURL("")
		if err != nil {
			return err
		}

		if serverBaseURL == "" {
			return errors.New("could not determine Kartusche server")
		}

		name := c.Args().First()

		if name == "" {
			return errors.New("name of kartusche must be provided")
		}

		dbs := &runtime.DBStats{}
		err = client.CallAPI(serverBaseURL, "GET", path.Join("kartusches", name, "info", "dbstats"), nil, nil, client.JSONDecoder(&dbs), 200)
		if err != nil {
			return fmt.Errorf("while starting login process: %w", err)
		}

		return yaml.NewEncoder(os.Stdout).Encode(dbs)

	},
}
