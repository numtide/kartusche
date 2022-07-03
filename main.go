package main

import (
	"github.com/draganm/kartusche/command/auth"
	"github.com/draganm/kartusche/command/clone"
	"github.com/draganm/kartusche/command/develop"
	initCmd "github.com/draganm/kartusche/command/init"
	"github.com/draganm/kartusche/command/ls"
	"github.com/draganm/kartusche/command/rm"
	"github.com/draganm/kartusche/command/server"
	"github.com/draganm/kartusche/command/test"
	"github.com/draganm/kartusche/command/update"
	"github.com/draganm/kartusche/command/upload"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			test.Command,
			server.Command,
			upload.Command,
			ls.Command,
			rm.Command,
			update.Command,
			develop.Command,
			initCmd.Command,
			auth.Command,
			clone.Command,
		},
	}
	app.RunAndExitOnError()

}
