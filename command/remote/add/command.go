package add

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/draganm/kartusche/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "add",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "default",
		},
	},
	Action: func(c *cli.Context) (err error) {

		if c.Args().Len() != 2 {
			return errors.New("<remote name> and <url> must be provided")
		}

		args := c.Args().Slice()
		name := args[0]

		urlString := args[1]

		_, err = url.Parse(urlString)
		if err != nil {
			return fmt.Errorf("could not parse URL %s: %w", urlString, err)
		}

		return config.Update(".", func(cfg *config.Config) error {
			cfg.Remotes[name] = urlString
			if c.Bool("default") {
				cfg.DefaultRemote = name
			}
			return nil
		})

	},
}
