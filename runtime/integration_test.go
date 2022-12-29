package runtime_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/draganm/kartusche/runtime/testrig"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func init() {
	logger, _ := zap.NewDevelopment()
	if true {
		opts.DefaultContext = logr.NewContext(context.Background(), zapr.NewLogger(logger))
	}
}

var opts = godog.Options{
	Output:        os.Stdout,
	StopOnFailure: true,
	Strict:        true,
	Paths:         []string{"features"},
	NoColors:      true,
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
	ti             testrig.TestKartuscheInstance
	lastStatusCode int
	lastResponse   string
}

func (s *State) get(path string) (int, string, error) {
	u, err := url.JoinPath(s.ti.GetURL(), path)
	if err != nil {
		return -1, "", fmt.Errorf("could not join path for GET request: %w", err)
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return -1, "", fmt.Errorf("could not create GET request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, "", fmt.Errorf("could not perform GET request: %w", err)
	}

	defer res.Body.Close()

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, "", fmt.Errorf("could not read response body: %w", err)
	}

	return res.StatusCode, string(d), nil

}

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

		tri, err := testrig.StartTestKartuscheInstance(ctx)
		if err != nil {
			return ctx, fmt.Errorf("could not start test rig: %w", err)
		}

		state.ti = tri

		ctx = context.WithValue(ctx, stateKey, state)

		return ctx, nil
	})

	ctx.Step(`^a kartusche with a root get handler$`, aKartuscheWithARootGetHandler)
	ctx.Step(`^the kartusche receives GET request$`, theKartuscheReceivesGETRequest)
	ctx.Step(`^the kartusche should respond with (\d+) status code$`, theKartuscheShouldRespondWithStatusCode)
}

func getState(ctx context.Context) *State {
	return ctx.Value(stateKey).(*State)
}

func aKartuscheWithARootGetHandler(ctx context.Context) error {
	s := getState(ctx)
	return s.ti.AddContent("handler/GET.js", `w.write("OK")`)
}

func theKartuscheReceivesGETRequest(ctx context.Context) error {
	s := getState(ctx)
	var err error
	s.lastStatusCode, s.lastResponse, err = s.get("/")
	if err != nil {
		return err
	}
	return nil
}

func theKartuscheShouldRespondWithStatusCode(ctx context.Context, expectedStatusCode int) error {
	s := getState(ctx)

	if s.lastStatusCode != expectedStatusCode {
		return fmt.Errorf("expected status code %d but got %d: %s", expectedStatusCode, s.lastStatusCode, s.lastResponse)
	}

	return nil
}
