package develop

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	gopath "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/common/paths"
	"github.com/draganm/kartusche/common/serverurl"
	"github.com/draganm/kartusche/config"
	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/zapr"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var Command = &cli.Command{
	Name: "develop",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "addr",
			EnvVars: []string{"KARTUSCHE_ADDR"},
			Value:   "localhost:5001",
		},
	},
	Action: func(c *cli.Context) (err error) {
		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while running dev server: %w", err), 1)
			}
		}()

		dl, err := zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("while starting logger: %w", err)
		}

		defer dl.Sync()

		log := zapr.NewLogger(dl)

		dir := "."

		if err != nil {
			return err
		}

		serverBaseURL, err := serverurl.BaseServerURL(c.Args().First())
		if err != nil {
			return err
		}

		log.Info("syncing with server", "server", serverBaseURL)

		w, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("while creating fs notify watcher")
		}

		names := make(chan string, 20)
		names <- "."
		done := make(chan error)
		go watch(dir, w, names, done)
		for name := range names {
			// debounce

			allNames := []string{name}
		inner:
			for {
				select {
				case name = <-names:
					allNames = append(allNames, name)
				case <-time.NewTimer(100 * time.Millisecond).C:
					break inner
				}
			}

			log.Info("updating server", "changedFiles", allNames)

			err := UpdateServerCode(dir, serverBaseURL)
			if err != nil {
				log.Error(err, "failed to update server")
				continue
			}
			log.Info("updated server")
		}

		return nil

	},
}

func UpdateServerCode(dir, serverBaseURL string) error {

	cfg, err := config.Current()
	if err != nil {
		return fmt.Errorf("while getting current config: %w", err)
	}

	tf, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("while creating temp file: %w", err)
	}

	defer tf.Close()
	defer os.Remove(tf.Name())

	tw := tar.NewWriter(tf)

	for _, p := range paths.WellKnown {

		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}

		_, err = os.Stat(p)
		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return err
		}

		absDir, err := filepath.Abs(p)

		if err != nil {
			return fmt.Errorf("while getting absolute dir of %s: %w", dir, err)
		}

		absDirParts := strings.Split(absDir, string(os.PathSeparator))

		err = filepath.Walk(absDir, func(file string, fi os.FileInfo, err error) error {

			if err != nil {
				return fmt.Errorf("while walking dir: %w", err)
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

	err = client.CallAPI(serverBaseURL, "PATCH", gopath.Join("kartusches", cfg.Name, "code"), nil, func() (io.Reader, error) { return tf, nil }, nil, 204)
	if err != nil {
		return err
	}

	return nil

}
