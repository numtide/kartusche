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
		&cli.StringFlag{
			Name:    "work-dir",
			Value:   "work",
			EnvVars: []string{"WORK_DIR"},
		},
	},
	Action: func(c *cli.Context) (err error) {
		defer func() {
			if err != nil {
				err = cli.Exit(fmt.Errorf("while running server: %w", err), 1)
			}
		}()

		ks, err := open(c.String("work-dir"))
		if err != nil {
			return fmt.Errorf("while starting kartusche server: %w", err)
		}

		r := mux.NewRouter()

		controllerHost := c.String("controller-hostname")

		r.Host(controllerHost).Methods("PUT").Path("/kartusches/{name}").HandlerFunc(ks.upload)
		r.Host(controllerHost).Methods("GET").Path("/kartusches").HandlerFunc(ks.list)

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
