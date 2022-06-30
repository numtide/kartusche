package ls

import (
	"fmt"
	"os"
	"strings"

	"github.com/draganm/kartusche/command/server"
	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/config"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "ls",
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

		cfg, err := config.Current()
		if err != nil {
			return fmt.Errorf("while getting current config: %w", err)
		}

		serverBaseURL := cfg.GetServerBaseURL(c.String("kartusche-server-base-url"))

		kl := []server.KartuscheListEntry{}
		err = client.CallAPI(serverBaseURL, "GET", "kartusches", nil, nil, &kl, 200)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Hosts", "Prefix"})
		for _, k := range kl {
			table.Append([]string{k.Name, strings.Join(k.Hosts, ", "), k.Prefix})
		}

		table.Render()
		return nil

	},
}
