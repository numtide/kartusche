package auth

import (
	"github.com/draganm/kartusche/command/auth/login"
	"github.com/draganm/kartusche/command/auth/tokens"
	"github.com/draganm/kartusche/command/auth/workspace"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "auth",
	Subcommands: []*cli.Command{
		login.Command,
		tokens.Command,
		workspace.Command,
	},
}
