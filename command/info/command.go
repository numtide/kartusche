package info

import (
	"github.com/draganm/kartusche/command/info/handlers"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "info",
	Subcommands: []*cli.Command{
		handlers.Command,
	},
}
