package main

import (
	"github.com/draganm/kartusche/command/ls"
	"github.com/draganm/kartusche/command/server"
	"github.com/draganm/kartusche/command/test"
	"github.com/draganm/kartusche/command/upload"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Action: func(ctx *cli.Context) error {
			return nil
		},
		Commands: []*cli.Command{
			test.Command,
			server.Command,
			upload.Command,
			ls.Command,
		},
	}
	app.RunAndExitOnError()

}
