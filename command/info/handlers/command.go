package handlers

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/draganm/kartusche/common/auth"
	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/common/serverurl"
	"github.com/draganm/kartusche/server"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "handlers",
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

		tkn, err := auth.GetTokenForServer(serverBaseURL)
		if err != nil {
			return fmt.Errorf("could not get token for server: %w", err)
		}

		handlers := []server.HandlerInfo{}
		err = client.CallAPI(context.Background(), serverBaseURL, tkn, "GET", path.Join("kartusches", name, "info", "handlers"), nil, nil, client.JSONDecoder(&handlers), 200)
		if err != nil {
			return fmt.Errorf("while starting login process: %w", err)
		}

		for _, h := range handlers {
			fmt.Printf("%s\t%s\n", h.Verb, h.Path)
		}
		return nil

	},
}
