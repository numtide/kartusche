package server_test

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/cucumber/godog"
	"github.com/draganm/kartusche/server/testrig"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func init() {
	logger, _ := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	if true {
		opts.DefaultContext = logr.NewContext(context.Background(), zapr.NewLogger(logger))
	}
}

var opts = godog.Options{
	Output:        os.Stdout,
	StopOnFailure: true,
	Strict:        true,
	Format:        "progress",
	Paths:         []string{"features"},
	NoColors:      true,
	Concurrency:   runtime.NumCPU() * 2,
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	pflag.Parse()
	opts.Paths = pflag.Args()

	status := godog.TestSuite{
		Name:                "godogs",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	os.Exit(status)
}

type State struct {
	si         *testrig.TestServerInstance
	serverDump []byte
}

// func (s *State) get(path string) (int, string, error) {
// u, err := url.JoinPath(s.ti.GetURL(), path)
// if err != nil {
// 	return -1, "", fmt.Errorf("could not join path for GET request: %w", err)
// }

// req, err := http.NewRequest("GET", u, nil)
// if err != nil {
// 	return -1, "", fmt.Errorf("could not create GET request: %w", err)
// }

// res, err := http.DefaultClient.Do(req)
// if err != nil {
// 	return -1, "", fmt.Errorf("could not perform GET request: %w", err)
// }

// defer res.Body.Close()

// d, err := io.ReadAll(res.Body)
// if err != nil {
// 	return -1, "", fmt.Errorf("could not read response body: %w", err)
// }

// return res.StatusCode, string(d), nil

// }

type StateKeyType string

const stateKey = StateKeyType("")

func InitializeScenario(ctx *godog.ScenarioContext) {
	var cancel context.CancelFunc

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		ctx, cancel = context.WithCancel(ctx)

		return ctx, nil

	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		cancel()
		return ctx, nil
	})

	state := &State{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {

		ts, err := testrig.NewTestServerInstance(ctx, logr.FromContextOrDiscard(ctx))
		if err != nil {
			return ctx, fmt.Errorf("could not start test rig: %w", err)
		}

		state.si = ts

		ctx = context.WithValue(ctx, stateKey, state)

		return ctx, nil
	})

	ctx.Step(`^a server with a single kartusche$`, aServerWithASingleKartusche)
	ctx.Step(`^I backup the server$`, iBackupTheServer)
	ctx.Step(`^I delete the kartusche$`, iDeleteTheKartusche)
	ctx.Step(`^I restore the server$`, iRestoreTheServer)
	ctx.Step(`^the kartusche should existd$`, theKartuscheShouldExistd)

}

func getState(ctx context.Context) *State {
	return ctx.Value(stateKey).(*State)
}

func aServerWithASingleKartusche(ctx context.Context) error {
	s := getState(ctx)
	err := s.si.CreateKartusche(ctx, "k1")
	if err != nil {
		return err
	}
	return nil
}

func iBackupTheServer(ctx context.Context) error {
	s := getState(ctx)
	d, err := s.si.DumpServer(ctx)
	if err != nil {
		return err
	}
	s.serverDump = d
	return nil
}

func iDeleteTheKartusche(ctx context.Context) error {
	s := getState(ctx)
	err := s.si.DeleteKartusche(ctx, "k1")
	if err != nil {
		return err
	}
	return nil
}

func iRestoreTheServer() error {
	return godog.ErrPending
}

func theKartuscheShouldExistd() error {
	return godog.ErrPending
}
