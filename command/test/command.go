package test

import (
	"fmt"

	"github.com/draganm/kartusche/tests"
	_ "github.com/draganm/kartusche/tests"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "test",
	Action: func(ctx *cli.Context) (err error) {
		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while uploading Kartusche: %w", err), 1)
			}
		}()
		return tests.Run(".")
	},
}
