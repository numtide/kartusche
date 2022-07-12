package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/draganm/kartusche/server"
	"github.com/draganm/kartusche/server/verifier"
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
		&cli.StringFlag{
			Name:    "auth-provider",
			Value:   "mock",
			EnvVars: []string{"AUTH_PROVIDER"},
		},
		&cli.StringFlag{
			Name:    "oauth2-github-client-id",
			EnvVars: []string{"OAUTH2_GITHUB_CLIENT_ID"},
		},
		&cli.StringFlag{
			Name:    "oauth2-github-client-secret",
			EnvVars: []string{"OAUTH2_GITHUB_CLIENT_SECRET"},
		},
		&cli.StringFlag{
			Name:    "oauth2-github-organization",
			EnvVars: []string{"OAUTH2_GITHUB_ORGANIZATION"},
		},
		&cli.StringFlag{
			Name:    "kartusche-domain",
			EnvVars: []string{"KARTUSCHE_DOMAIN"},
			Value:   "127.0.0.1.nip.io",
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

		vf := verifier.NewMockProvider()

		if c.String("auth-provider") == "github" {
			switch {
			case !c.IsSet("oauth2-github-client-id"):
				return fmt.Errorf("OAUTH2_GITHUB_CLIENT_ID must be set")
			case !c.IsSet("oauth2-github-client-secret"):
				return fmt.Errorf("OAUTH2_GITHUB_CLIENT_SECRET must be set")
			case !c.IsSet("oauth2-github-organization"):
				return fmt.Errorf("OAUTH2_GITHUB_ORGANIZATION")
			}
			vf = verifier.NewGithubProvider(
				c.String("oauth2-github-client-id"),
				c.String("oauth2-github-client-secret"),
				c.String("oauth2-github-organization"),
			)
		}

		ks, err := server.Open(
			c.String("work-dir"),
			c.String("kartusche-domain"),
			vf,
			log,
		)
		if err != nil {
			return fmt.Errorf("while starting kartusche server: %w", err)
		}

		s := &http.Server{
			Handler: ks.ServerRouter,
		}
		serverAddr := c.String("controller-addr")
		l, err := net.Listen("tcp", serverAddr)
		if err != nil {
			return fmt.Errorf("while starting listener: %w", err)
		}
		log.With("addr", l.Addr().String()).Info("server started")

		kartuschesAddr := c.String("kartusches-addr")
		kl, err := net.Listen("tcp", kartuschesAddr)
		if err != nil {
			return fmt.Errorf("while creating kartusches listener: %w", err)
		}

		khs := &http.Server{
			Handler: ks,
		}

		go func() {
			log.With("addr", kl.Addr().String()).Infof("listening for kartusche requests")
			err := khs.Serve(kl)
			if err != nil {
				log.With("server", "kartusche", "error", err).Error("while serving kartusches")
			}
		}()

		return s.Serve(l)
	},
}
