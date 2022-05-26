package ls

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/draganm/kartusche/command/server"
	"github.com/draganm/kartusche/config"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "ls",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "kartusche-server-base-url",
			Value:   "http://localhost:3003",
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

		baseUrl, err := url.Parse(serverBaseURL)
		if err != nil {
			return fmt.Errorf("while parsing server base url: %w", err)
		}

		baseUrl.Path = path.Join(baseUrl.Path, "kartusches")

		req, err := http.NewRequest("GET", baseUrl.String(), nil)
		if err != nil {
			return fmt.Errorf("while creating request: %w", err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("while performing GET request: %w", err)
		}

		defer res.Body.Close()

		if res.StatusCode != 200 {
			return fmt.Errorf("unexpected status %s", res.Status)
		}

		kl := []server.KartuscheListEntry{}

		err = json.NewDecoder(res.Body).Decode(&kl)
		if err != nil {
			return fmt.Errorf("while decoding response")
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
