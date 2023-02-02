package tokens

import (
	"fmt"
	"sort"

	"github.com/draganm/kartusche/common/auth"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "tokens",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while getting token: %w", err), 1)
			}
		}()

		tokensMap, err := auth.GetAllTokens()
		if err != nil {
			return fmt.Errorf("could not get server tokens: %w", err)
		}

		servers := []string{}

		for s := range tokensMap {
			servers = append(servers, s)
		}

		sort.Strings(servers)

		for _, s := range servers {
			fmt.Println(s, tokensMap[s])
		}

		return nil

	},
}
