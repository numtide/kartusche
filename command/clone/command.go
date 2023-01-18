package clone

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/draganm/kartusche/common/auth"
	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/config"
	"github.com/urfave/cli/v2"
	"go.uber.org/multierr"
)

var Command = &cli.Command{
	Name:  "clone",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while cloning Kartusche: %w", err), 1)
			}
		}()

		ctx := context.Background()

		if c.NArg() != 1 {
			return errors.New("kartusche url must be provided")
		}

		u, err := url.Parse(c.Args().First())
		if err != nil {
			return fmt.Errorf("while parsing kartusche URL: %w", err)
		}

		parts := strings.Split(u.Path, "/")
		if len(parts) != 3 {
			return fmt.Errorf("path of the kartusche must be kartusches/<name>")
		}

		if parts[1] != "kartusches" {
			return fmt.Errorf("path of the kartusche must be kartusches/<name>")
		}

		kartuscheName := parts[2]

		baseURL := &url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
		}

		_, err = os.Stat(kartuscheName)
		if !os.IsNotExist(err) {
			return fmt.Errorf("directory %q already exists", kartuscheName)
		}

		dumpWriter := func(r io.Reader) error {
			fmt.Printf("cloning kartusche %s\n", u.String())

			tr := tar.NewReader(r)
			for {

				h, err := tr.Next()
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return fmt.Errorf("while reading tar header: %w", err)
				}

				filePath := filepath.Join(kartuscheName, filepath.FromSlash(path.Clean(h.Name)))

				if h.Typeflag == tar.TypeDir {
					err = os.MkdirAll(filePath, 0700)
					if err != nil {
						return err
					}
					continue
				}

				err = func() (err error) {
					f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0700)
					if err != nil {
						return err
					}

					defer func() {
						ce := f.Close()
						err = multierr.Combine(err, ce)
					}()

					_, err = io.Copy(f, tr)
					if err != nil {
						return err
					}
					return nil

				}()
				if err != nil {
					return err
				}

			}

		}

		tkn, err := auth.GetTokenForServer(baseURL.String())
		if err != nil {
			return fmt.Errorf("could not get token for server: %w", err)
		}

		err = client.CallAPI(ctx, baseURL.String(), tkn, "GET", u.Path, nil, nil, dumpWriter, 200)
		if err != nil {
			return err
		}

		cfg := &config.Config{
			Name:          kartuscheName,
			DefaultRemote: "origin",
			Remotes: map[string]string{
				"origin": baseURL.String(),
			},
		}

		err = cfg.Write(kartuscheName)
		if err != nil {
			return fmt.Errorf("while writing kartusche config: %w", err)
		}

		return nil

	},
}
