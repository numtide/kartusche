package remote

import (
	"github.com/draganm/kartusche/command/remote/add"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "remote",
	Subcommands: []*cli.Command{
		add.Command,
	},
}
