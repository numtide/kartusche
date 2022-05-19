package main

import (
	"github.com/draganm/kartusche/command/test"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Action: func(ctx *cli.Context) error {
			return nil
		},
		Commands: []*cli.Command{
			test.Command,
		},
	}
	app.RunAndExitOnError()

}
