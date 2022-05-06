package runtime_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/draganm/kartusche/runtime"
	"github.com/nsf/jsondiff"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {

			w, err := newWorld()
			if err != nil {
				panic(err)
			}

			ctx.Step(`^a service with the "([^"]*)" handler "([^"]*)"$`, w.aServiceWithTheHandler)
			ctx.Step(`^I get the path "([^"]*)" from the Kartusche$`, w.iGetThePathFromTheKartusche)
			ctx.Step(`^I should get status code (\d+)$`, w.iShouldGetStatusCode)
			ctx.Step(`^I post following data to the "([^"]*)" path of the Kartusche$`, w.iPostFollowingDataToThePathOfTheKartusche)
			ctx.Step(`^the result body should match JSON$`, w.theResultBodyShouldMatchJSON)
			// Add step definitions here.
		},
		Options: &godog.Options{
			Format:              "pretty",
			Paths:               []string{"features"},
			TestingT:            t, // Testing instance that will run subtests.
			NoColors:            true,
			Strict:              true,
			StopOnFailure:       true,
			ShowStepDefinitions: false,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func newWorld() (*world, error) {
	tf, err := os.CreateTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("while creating temp file: %w", err)
	}

	err = tf.Close()
	if err != nil {
		return nil, fmt.Errorf("while closing temp file: %w", err)
	}

	err = os.Remove(tf.Name())
	if err != nil {
		return nil, fmt.Errorf("while removing temp file: %w", err)
	}

	return &world{dbFileName: tf.Name()}, nil
}

type world struct {
	dbFileName   string
	httpResponse *httptest.ResponseRecorder
}

func (w *world) addHandler(method, path, source string) error {
	db, err := embedded.Open(w.dbFileName, 0700)
	if err != nil {
		return fmt.Errorf("while opening the database: %w", err)
	}

	defer db.Close()

	err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
		handersPath := dbpath.ToPath("__handlers__")
		if !tx.Exists(handersPath) {
			tx.CreateMap(handersPath)
		}
		handlerPath := handersPath.Append(fmt.Sprintf("%s:%s", method, path))
		tx.CreateMap(handlerPath)
		meta := runtime.HandlerMeta{
			Method: method,
			Path:   path,
		}
		md, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("while marshalling meta data: %w", err)
		}

		tx.Put(handlerPath.Append("meta_data"), md)
		tx.Put(handlerPath.Append("source"), []byte(source))
		return nil
	})

	if err != nil {
		return fmt.Errorf("while adding handler: %w", err)
	}

	return nil
}

// type aServiceWithTheHandler struct{}

func (w *world) aServiceWithTheHandler(method, path string, src *godog.DocString) error {
	return w.addHandler(method, path, src.Content)
}

func (w *world) iGetThePathFromTheKartusche(path string) error {
	rt, err := runtime.New(w.dbFileName, "")
	if err != nil {
		return fmt.Errorf("while starting runtime: %w", err)
	}

	defer rt.Shutdown()

	mw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("while creating http request: %w", err)
	}

	rt.ServeHTTP(mw, req)

	w.httpResponse = mw
	return nil
}

func (w *world) iShouldGetStatusCode(expectedStatusCode int) error {
	sc := w.httpResponse.Result().StatusCode
	if sc != expectedStatusCode {
		return fmt.Errorf("unexpected status: %s, expected: %d", w.httpResponse.Result().Status, expectedStatusCode)
	}
	return nil
}

func (w *world) iPostFollowingDataToThePathOfTheKartusche(path string, dataDoc *godog.DocString) error {
	rt, err := runtime.New(w.dbFileName, "")
	if err != nil {
		return fmt.Errorf("while starting runtime: %w", err)
	}

	defer rt.Shutdown()

	mw := httptest.NewRecorder()

	req, err := http.NewRequest("POST", path, strings.NewReader(dataDoc.Content))
	if err != nil {
		return fmt.Errorf("while creating http request: %w", err)
	}

	rt.ServeHTTP(mw, req)

	w.httpResponse = mw
	return nil
}

func (w *world) theResultBodyShouldMatchJSON(expected *godog.DocString) error {
	rb, err := io.ReadAll(w.httpResponse.Result().Body)
	if err != nil {
		return fmt.Errorf("while reading response body: %w", err)
	}

	opts := jsondiff.DefaultJSONOptions()
	diffStatus, diff := jsondiff.Compare(rb, []byte(expected.Content), &opts)

	if diffStatus == jsondiff.FullMatch {
		return nil
	}

	return errors.New(diff)
}
