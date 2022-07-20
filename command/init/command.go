package init

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/draganm/kartusche/command/init/skeleton"
	"github.com/draganm/kartusche/config"
	"github.com/urfave/cli/v2"
)

var dirs = []string{
	".kartusche",
	"handler",
	"tests",
	"tests/feature",
	"tests/support",
	"static",
}

var Command = &cli.Command{
	Name: "init",
	Action: func(c *cli.Context) (err error) {

		name := c.Args().First()

		if name == "" {
			return errors.New("name of the Kartusche must be provided")
		}

		_, err = os.Stat(name)
		switch {
		case os.IsNotExist(err):
			// happy case - dir does not exist
		case err != nil:
			return err
		default:
			return fmt.Errorf("dir %s already exists", name)
		}

		err = os.Mkdir(name, 0700)
		if err != nil {
			return fmt.Errorf("while creating dir %s: %w", name, err)
		}

		toDo := []string{"."}
		for len(toDo) > 0 {
			current := toDo[0]
			toDo = toDo[1:]
			ents, err := skeleton.Content.ReadDir(current)
			if err != nil {
				return fmt.Errorf("while reading skeleton dir %s: %w", current, err)
			}
			for _, e := range ents {
				fullPath := path.Join(current, e.Name())
				pth := filepath.Join(name, filepath.FromSlash(fullPath))
				fmt.Println(">", pth)
				if e.IsDir() {
					err = os.Mkdir(pth, 0700)
					if err != nil {
						return fmt.Errorf("while creating dir %s: %w", name, err)
					}
					toDo = append(toDo, fullPath)
					continue
				}
				inf, err := skeleton.Content.Open(fullPath)
				if err != nil {
					return fmt.Errorf("while opening skeleton file %s: %w", fullPath, err)
				}

				ouf, err := os.OpenFile(pth, os.O_WRONLY|os.O_CREATE, 0700)
				if err != nil {
					return fmt.Errorf("while opening output file %s: %w", pth, err)
				}

				_, err = io.Copy(ouf, inf)
				if err != nil {
					return fmt.Errorf("while writing %s: %w", pth, err)
				}

				ouf.Close()

				inf.Close()
				// fmt.Println(path.Join(current, e.Name()))
			}

		}

		cfg := &config.Config{
			Name:          name,
			DefaultRemote: "origin",
			Remotes: map[string]string{
				"origin": "https://kartusche.netice9.xyz",
			},
		}

		err = cfg.Write(name)
		if err != nil {
			return fmt.Errorf("while writing config: %w", err)
		}

		return nil
	},
}
