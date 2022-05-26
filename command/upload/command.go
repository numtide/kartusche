package upload

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/draganm/kartusche/config"
	"github.com/draganm/kartusche/manifest"
	"github.com/draganm/kartusche/runtime"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "upload",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "kartusche-server-base-url",
			EnvVars: []string{"KARTUSCHE_SERVER_BASE_URL"},
		},
		&cli.StringFlag{
			Name: "name",
		},
		&cli.StringSliceFlag{
			Name:    "hostname",
			EnvVars: []string{"HOSTNAMES"},
		},
		&cli.StringFlag{
			Name:    "prefix",
			Value:   "/",
			EnvVars: []string{"PREFIX"},
		},
	},
	Action: func(c *cli.Context) (err error) {
		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while uploading Kartusche: %w", err), 1)
			}
		}()

		dir := c.Args().First()

		if dir == "" {
			dir = "."
		}

		km, err := manifest.Load(dir)
		if err != nil {
			return fmt.Errorf("while loading manifest")
		}

		name := c.String("name")

		if name == "" {
			name = km.Name
		}

		if name == "" {
			absPath, err := filepath.Abs(".")
			if err != nil {
				return fmt.Errorf("while getting absolute path of the current dir")
			}
			name = filepath.Base(absPath)
		}

		td, err := os.MkdirTemp("", "")
		if err != nil {
			return fmt.Errorf("while creating temp dir: %w", err)
		}

		defer os.Remove(td)

		kartuscheFileName := filepath.Join(td, "kartusche")

		err = runtime.InitializeNew(kartuscheFileName, dir)
		if err != nil {
			return fmt.Errorf("while initializing Kartusche: %w", err)
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

		hostnames := km.Hostnames

		if c.IsSet("hostname") {
			hostnames = c.StringSlice("hostname")
		}

		q := url.Values{}
		q["hostname"] = hostnames

		prefix := c.String("prefix")

		if prefix == "" {
			prefix = km.Prefix
		}

		q.Set("prefix", prefix)

		baseUrl.RawQuery = q.Encode()

		kf, err := os.Open(kartuscheFileName)
		if err != nil {
			return err
		}

		defer kf.Close()

		req, err := http.NewRequest("PUT", baseUrl.String(), kf)
		if err != nil {
			return fmt.Errorf("while creating request: %w", err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("while performing PUT request: %w", err)
		}

		defer res.Body.Close()

		if res.StatusCode != 204 {
			return fmt.Errorf("unexpected status %s", res.Status)
		}

		return nil
	},
}
