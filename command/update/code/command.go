package code

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/common/serverurl"
	"github.com/draganm/kartusche/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "code",
	Flags: []cli.Flag{},
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

		cfg, err := config.Current()
		if err != nil {
			return fmt.Errorf("while getting current config: %w", err)
		}

		serverBaseURL, err := serverurl.BaseServerURL(c.Args().First())
		if err != nil {
			return err
		}

		tf, err := os.CreateTemp("", "")
		if err != nil {
			return fmt.Errorf("while creating temp file: %w", err)
		}

		defer tf.Close()
		defer os.Remove(tf.Name())

		tw := tar.NewWriter(tf)

		pathsToLoad := map[string]string{
			"static":    "static",
			"cronjobs":  "cronjobs",
			"handler":   "handler",
			"lib":       "lib",
			"tests":     "tests",
			"templates": "templates",
			"init.js":   "init.js",
		}

		for p, pth := range pathsToLoad {

			if !filepath.IsAbs(pth) {
				pth = filepath.Join(dir, pth)
			}

			_, err = os.Stat(pth)
			if os.IsNotExist(err) {
				continue
			}

			if err != nil {
				return err
			}

			absDir, err := filepath.Abs(pth)

			if err != nil {
				return fmt.Errorf("while getting absolute dir of %s: %w", dir, err)
			}

			absDirParts := strings.Split(absDir, string(os.PathSeparator))

			err = filepath.Walk(absDir, func(file string, fi os.FileInfo, err error) error {

				if err != nil {
					return err
				}

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
				pathParts = append([]string{p}, pathParts[len(absDirParts):]...)

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
		}

		_, err = tf.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("while seeking tar file to beginning: %w", err)
		}

		err = client.CallAPI(serverBaseURL, "PATCH", path.Join("kartusches", cfg.Name, "code"), nil, func() (io.Reader, error) { return tf, nil }, nil, 204)
		if err != nil {
			return err
		}

		fmt.Println("code updated")
		return nil

	},
}
