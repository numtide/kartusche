package develop

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/kartusche/common/paths"
	"github.com/draganm/kartusche/common/util/path"
	"github.com/draganm/kartusche/runtime"
	"github.com/draganm/kartusche/tests"
	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var Command = &cli.Command{
	Name: "develop",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "addr",
			EnvVars: []string{"ADDR"},
			Value:   ":5001",
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

		dir := "."

		_, err = os.Stat(".kartusche")
		if os.IsNotExist(err) {
			err = os.Mkdir(".kartusche", 0700)
		}

		if err != nil {
			return err
		}

		_, err = os.Stat(".kartusche/development")
		if os.IsNotExist(err) {
			err = runtime.InitializeNew(".kartusche/development", dir)
		}

		if err != nil {
			return fmt.Errorf("while initializing runtime: %w", err)
		}

		l, err := net.Listen("tcp", c.String("addr"))
		if err != nil {
			return fmt.Errorf("while creating listener: %w", err)
		}

		rt, err := runtime.Open(".kartusche/development", dl.Sugar())
		if err != nil {
			return fmt.Errorf("while starting runtime: %w", err)
		}

		s := &http.Server{
			Handler: rt,
		}

		w, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("while creating fs notify watcher")
		}

		go func() {
			names := make(chan string, 20)
			names <- "."
			done := make(chan error)
			go watch(dir, w, names, done)
			for range names {
				// debounce
			inner:
				for {
					select {
					case <-names:
					case <-time.NewTimer(100 * time.Millisecond).C:
						break inner
					}
				}

				err := updateRuntimeCode(rt, dir)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to update runtime: %w", err))
					continue
				}
				fmt.Println("updated runtime")
				err = tests.Run(dir)
				if err != nil {
					fmt.Println(err)
				}
			}
		}()

		return s.Serve(l)

	},
}

func updateRuntimeCode(rt runtime.Runtime, dir string) error {

	return rt.Update(func(tx bolted.SugaredWriteTx) error {
		for _, p := range paths.WellKnown {
			pth := dbpath.ToPath(p)
			if tx.Exists(pth) {
				tx.Delete(pth)
			}

			localFile := filepath.Join(dir, p)
			err := loadFromPath(localFile, tx, pth)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func loadFromPath(dir string, wtx bolted.SugaredWriteTx, prefix dbpath.Path) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("while getting abs path: %w", err)
	}

	_, err = os.Stat(absDir)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	absDirPaths := strings.Split(absDir, string(os.PathSeparator))

	err = filepath.Walk(absDir, func(file string, fi os.FileInfo, err error) error {

		if err != nil {
			return fmt.Errorf("while walking dir: %w", err)
		}

		pathParts := path.FilePathToDBPath(file)

		dbPathParts := pathParts[len(absDirPaths):]

		dbp := prefix.Append(dbPathParts...)
		if len(dbp) == 0 {
			return nil
		}

		if fi.IsDir() {
			wtx.CreateMap(dbp)
			return nil
		}

		if fi.Mode().IsRegular() {
			d, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("while reading %s: %w", file, err)
			}

			wtx.Put(dbp, d)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil

}
