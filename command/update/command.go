package update

import (
	"github.com/draganm/kartusche/command/update/code"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "update",
	Subcommands: []*cli.Command{
		code.Command,
	},
}
