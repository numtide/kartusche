package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name: "server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "controller-hostname",
			Value:   "localhost",
			EnvVars: []string{"CONTROLLER_HOSTNAME"},
		},
		&cli.StringFlag{
			Name:    "addr",
			Value:   ":3003",
			EnvVars: []string{"ADDR"},
		},
	},
	Action: func(c *cli.Context) (err error) {
		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while running server: %w", err), 1)
			}
		}()
		r := mux.NewRouter()
		s := &http.Server{
			Handler: r,
		}
		l, err := net.Listen("tcp", c.String("addr"))
		if err != nil {
			return fmt.Errorf("while starting listener: %w", err)
		}
		return s.Serve(l)
	},
}
