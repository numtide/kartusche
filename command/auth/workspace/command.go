package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/draganm/kartusche/common/auth"
	"github.com/urfave/cli/v2"
)

type WorkspaceFile struct {
	Folders []WorkspaceFolder `json:"folders"`
}

type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

var Command = &cli.Command{
	Name:  "workspace",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) (err error) {

		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while getting token: %w", err), 1)
			}
		}()

		tokensMap, err := auth.GetAllTokens()
		if err != nil {
			return fmt.Errorf("could not get server tokens: %w", err)
		}

		if c.Args().Len() != 1 {
			return errors.New("file name must be provided as first argument")
		}

		servers := []string{}

		for s := range tokensMap {
			servers = append(servers, s)
		}

		sort.Strings(servers)

		workspaceFile := &WorkspaceFile{}
		for _, s := range servers {

			token := tokensMap[s]

			u, err := url.Parse(s)
			if err != nil {
				return fmt.Errorf("could not parse URL %s: %w", s, err)
			}

			u.User = url.UserPassword("kartusche", token)
			u = u.JoinPath("dav")
			u.Scheme = strings.Replace(u.Scheme, "http", "webdav", 1)

			workspaceFile.Folders = append(workspaceFile.Folders, WorkspaceFolder{
				URI:  u.String(),
				Name: s,
			})
		}

		fd, err := json.MarshalIndent(workspaceFile, "", " ")
		if err != nil {
			return fmt.Errorf("could not marshal workspace file: %w", err)
		}

		err = os.WriteFile(c.Args().First(), fd, 0666)
		if err != nil {
			return fmt.Errorf("could not write workspace file %s: %w", c.Args().First(), err)
		}

		return nil

	},
}
