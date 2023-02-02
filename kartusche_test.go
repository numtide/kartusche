package main_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"go.uber.org/multierr"
)

func TestFeatures(t *testing.T) {
	btd, err := os.MkdirTemp("", "kartusche-binary")
	if err != nil {
		t.Fatal(fmt.Errorf("while creating binary temp dir: %w", err))
	}

	binaryPath := filepath.Join(btd, "kartusche")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	err = cmd.Run()
	if err != nil {
		t.Fatal(fmt.Errorf("while building kartusche: %w", err))
	}

	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			w, err := newWorld(binaryPath)
			if err != nil {
				panic(fmt.Errorf("while creating world: %w", err))
			}
			ctx.Step(`^the server is running$`, w.theServerIsRunning)
			ctx.Step(`^I authenticate the user using browser$`, w.iAuthenticateTheUserUsingBrowser)
			ctx.Step(`^the user config should contain token for the server$`, w.theUserConfigShouldContainTokenForTheServer)
			ctx.After(w.shutdown)
		},
		Options: &godog.Options{
			Format:   "progress",
			NoColors: true,
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.

		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

type world struct {
	dir        string
	binaryPath string
	s          *runningServer
}

func newWorld(binaryPath string) (*world, error) {
	td, err := os.MkdirTemp("", "kartusche-scenario")
	if err != nil {
		return nil, err
	}

	return &world{dir: td, binaryPath: binaryPath}, nil
}

func (w *world) cleanup() error {
	var err error
	if w.s != nil {
		err = multierr.Append(err, w.s.shutdown())
	}
	err = multierr.Append(err, os.RemoveAll(w.dir))

	return err
}

func (w *world) theServerIsRunning() error {
	rs, err := startServer()
	if err != nil {
		return fmt.Errorf("while starting server: %w", err)
	}
	w.s = rs
	return nil
}

func (w *world) iAuthenticateTheUserUsingBrowser(ctx context.Context) error {
	runningCli, err := startCLI(
		[]string{"auth", "login", w.s.serverURL},
		map[string]string{},
		w.dir,
		w.binaryPath,
	)
	if err != nil {
		return fmt.Errorf("while running auth login: %w", err)
	}

	defer runningCli.cleanup()

	bb := new(bytes.Buffer)

	scanner := bufio.NewScanner(io.TeeReader(runningCli.output, bb))
	if !scanner.Scan() {
		return fmt.Errorf("unable to read the first line from cmd: %s", scanner.Err())
	}
	line := scanner.Text()

	au := strings.TrimPrefix(line, "Please complete authentication flow by visiting ")
	res, err := http.Get(au)
	if err != nil {
		return fmt.Errorf("while visiting auth url: %w", err)
	}
	res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unable to finish auth flow, unexpected code %s returned from server", res.Status)
	}

	res.Body.Close()

	err = runningCli.waitToFinish()
	if err != nil {
		return err
	}

	return nil
}

func (w *world) theUserConfigShouldContainTokenForTheServer() error {
	out, _, err := runCLI(
		[]string{"auth", "tokens"},
		map[string]string{
			"KARTUSCHE_SERVER_BASE_URL": w.s.serverURL,
		},
		w.dir,
		w.binaryPath,
	)

	if err != nil {
		return err
	}

	if out == "" {
		return fmt.Errorf("could not find new token")
	}

	return nil
}

func (w *world) shutdown(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	if w.s != nil {
		err = multierr.Append(err, w.s.shutdown())
	}
	err = multierr.Append(err, os.RemoveAll(w.dir))

	return ctx, err

}
