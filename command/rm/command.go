package rm

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/draganm/kartusche/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "rm",
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

		name := c.Args().First()
		if name == "" {
			return errors.New("name of the Kartusche must be provided")
		}

		cfg, err := config.Current()
		if err != nil {
			return fmt.Errorf("while getting current config: %w", err)
		}

		serverBaseURL := cfg.GetServerBaseURL(c.String("kartusche-server-base-url"))

		baseUrl, err := url.Parse(serverBaseURL)
		if err != nil {
			return fmt.Errorf("while parsing server base url: %w", err)
		}

		baseUrl.Path = path.Join(baseUrl.Path, "kartusches", name)

		req, err := http.NewRequest("DELETE", baseUrl.String(), nil)
		if err != nil {
			return fmt.Errorf("while creating request: %w", err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("while performing DELETE request: %w", err)
		}

		defer res.Body.Close()

		if res.StatusCode != 204 {
			return fmt.Errorf("unexpected status %s", res.Status)
		}

		return nil

	},
}
