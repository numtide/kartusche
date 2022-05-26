package code

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/draganm/kartusche/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "code",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "kartusche-server-base-url",
			Value:   "http://localhost:3003",
			EnvVars: []string{"KARTUSCHE_SERVER_BASE_URL"},
		},
		&cli.StringFlag{
			Name: "name",
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

		name := c.String("name")
		if name == "" {
			absPath, err := filepath.Abs(".")
			if err != nil {
				return fmt.Errorf("while getting absolute path of the current dir")
			}
			name = filepath.Base(absPath)
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

		baseUrl.Path = path.Join(baseUrl.Path, "kartusches", name, "code")

		q := url.Values{}

		baseUrl.RawQuery = q.Encode()

		tf, err := os.CreateTemp("", "")
		if err != nil {
			return fmt.Errorf("while creating temp file: %w", err)
		}

		defer tf.Close()
		defer os.Remove(tf.Name())

		absDir, err := filepath.Abs(dir)

		if err != nil {
			return fmt.Errorf("while getting absolute dir of %s: %w", dir, err)
		}

		absDirParts := strings.Split(absDir, string(os.PathSeparator))

		tw := tar.NewWriter(tf)

		err = filepath.Walk(absDir, func(file string, fi os.FileInfo, err error) error {

			// generate tar header
			header, err := tar.FileInfoHeader(fi, file)
			if err != nil {
				return err
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			absPath, err := filepath.Abs(file)
			if err != nil {
				return fmt.Errorf("while getting absolute path of %s: %w", file, err)
			}
			pathParts := strings.Split(absPath, string(os.PathSeparator))
			pathParts = pathParts[len(absDirParts):]

			header.Name = filepath.ToSlash(strings.Join(pathParts, string(os.PathSeparator)))

			// write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if fi.Mode().IsRegular() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				_, err = io.Copy(tw, data)
				if err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return err
		}

		_, err = tf.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("while seeking tar file to beginning: %w", err)
		}

		req, err := http.NewRequest("PATCH", baseUrl.String(), tf)
		if err != nil {
			return fmt.Errorf("while creating request: %w", err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("while performing PATCH request: %w", err)
		}

		defer res.Body.Close()

		if res.StatusCode != 204 {
			return fmt.Errorf("unexpected status %s", res.Status)
		}

		fmt.Println("code updated")
		return nil

	},
}
