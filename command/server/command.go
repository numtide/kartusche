package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Command = &cli.Command{
	Name: "server",
	Flags: []cli.Flag{

		&cli.StringFlag{
			Name:    "controller-addr",
			Value:   ":3003",
			EnvVars: []string{"CONTROLLER_ADDR"},
		},
		&cli.StringFlag{
			Name:    "kartusches-addr",
			Value:   ":3002",
			EnvVars: []string{"KARTUSCHES_ADDR"},
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

		lc := zap.NewProductionConfig()

		lc.Sampling = nil
		lc.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		lc.DisableStacktrace = true
		lc.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

		logger, err := lc.Build()
		if err != nil {
			return
		}

		defer logger.Sync()
		log := logger.Sugar()

		ks, err := open(c.String("work-dir"), log)
		if err != nil {
			return fmt.Errorf("while starting kartusche server: %w", err)
		}

		r := mux.NewRouter()

		r.Methods("PUT").Path("/kartusches/{name}").HandlerFunc(ks.upload)
		r.Methods("GET").Path("/kartusches").HandlerFunc(ks.list)
		r.Methods("DELETE").Path("/kartusches/{name}").HandlerFunc(ks.rm)
		r.Methods("PATCH").Path("/kartusches/{name}/code").HandlerFunc(ks.updateCode)
		r.Methods("POST").Path("/auth/login").HandlerFunc(ks.loginStart)
		r.Methods("POST").Path("/auth/access_token").HandlerFunc(ks.accessToken)

		s := &http.Server{
			Handler: r,
		}
		serverAddr := c.String("controller-addr")
		log.Infof("server listening on %s", serverAddr)
		l, err := net.Listen("tcp", serverAddr)
		if err != nil {
			return fmt.Errorf("while starting listener: %w", err)
		}

		kartuschesAddr := c.String("kartusches-addr")
		kl, err := net.Listen("tcp", kartuschesAddr)
		if err != nil {
			return fmt.Errorf("while creating kartusches listener: %w", err)
		}

		khs := &http.Server{
			Handler: ks,
		}

		go func() {
			log.Infof("listening for kartusche requests on %s", kartuschesAddr)
			err := khs.Serve(kl)
			if err != nil {
				log.With("server", "kartusche", "error", err).Error("while serving kartusches")
			}
		}()

		return s.Serve(l)
	},
}
