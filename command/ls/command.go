package ls

import (
	"fmt"

	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/common/serverurl"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "ls",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while listing Kartusches: %w", err), 1)
			}
		}()

		serverBaseURL, err := serverurl.BaseServerURL(c.Args().First())
		if err != nil {
			return err
		}

		kl := []string{}
		err = client.CallAPI(serverBaseURL, "GET", "kartusches", nil, nil, client.JSONDecoder(&kl), 200)
		if err != nil {
			return err
		}

		for _, k := range kl {
			fmt.Println(k)
		}

		return nil

	},
}
