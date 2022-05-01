package main

import (
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
	app.RunAndExitOnError()
}
