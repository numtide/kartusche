package token

import (
	"fmt"

	"github.com/draganm/kartusche/common/auth"
	"github.com/draganm/kartusche/common/serverurl"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "token",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while getting token: %w", err), 1)
			}
		}()

		serverBaseURL, err := serverurl.BaseServerURL(c.Args().First())
		if err != nil {
			return err
		}

		tkn, err := auth.GetTokenForServer(serverBaseURL)

		if err != nil {
			return err
		}

		fmt.Println(tkn + "xx")
		return nil

	},
}
