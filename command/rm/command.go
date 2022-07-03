package rm

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/common/serverurl"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "rm",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while removing Kartusches: %w", err), 1)
			}
		}()

		firstArg := c.Args().First()

		parts := strings.Split(firstArg, "/")

		var name string
		var remote string

		switch len(parts) {
		case 1:
			name = parts[0]
		case 2:
			remote = parts[0]
			name = parts[1]
		default:
			return errors.New("either <kartusche name> or <remote name>/<kartusche name> must be provided as an argument")
		}

		serverBaseURL, err := serverurl.BaseServerURL(remote)
		if err != nil {
			return err
		}
		err = client.CallAPI(serverBaseURL, "DELETE", path.Join("kartusches", name), nil, nil, nil, 204)
		if err != nil {
			return err
		}

		return nil

	},
}
