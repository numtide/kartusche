package testrig

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/draganm/kartusche/common/client"
	"github.com/draganm/kartusche/runtime"
	"github.com/draganm/kartusche/server"
	"github.com/draganm/kartusche/server/verifier"
	"github.com/go-logr/logr"
)

type TestServerInstance struct {
	s            *server.Server
	apiURL       string
	kartuscheURL string
	adminToken   string
}

func NewTestServerInstance(ctx context.Context, log logr.Logger) (*TestServerInstance, error) {
	td, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("could not open test server instance dir: %w", err)
	}

	s, err := server.Open(td, "127.0.0.1.nip.io", verifier.NewMockProvider(), log)
	if err != nil {
		return nil, fmt.Errorf("could not start test server instance: %w", err)
	}

	ks := httptest.NewServer(s)
	as := httptest.NewServer(s.ServerRouter)

	tctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	loginStartResponse := &server.LoginStartResponse{}
	err = client.CallAPI(tctx, as.URL, "", "POST", "auth/login", nil, nil, client.JSONDecoder(loginStartResponse), 200)
	if err != nil {
		return nil, fmt.Errorf("while starting login process: %w", err)
	}

	ur, err := url.Parse(loginStartResponse.VerificationURI)
	if err != nil {
		return nil, fmt.Errorf("while parsing verification URI(%s): %w", loginStartResponse.VerificationURI, err)
	}

	q := url.Values{}
	q.Set("request_id", loginStartResponse.TokenRequestID)
	ur.RawQuery = q.Encode()

	res, err := http.Get(ur.String())
	if err != nil {
		return nil, fmt.Errorf("could not execute get for the verification URI: %w", err)
	}

	// return nil, fmt.Errorf("%s %s", ur.String(), res.Status)

	res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to finish auth flow, unexpected status %s", res.Status)
	}

	var token string

	for tctx.Err() == nil {

		var tr server.AccessTokenResponse
		err = client.CallAPI(
			tctx,
			as.URL,
			"",
			"POST", "auth/access_token",
			nil,
			client.JSONEncoder(
				server.RequestTokenParameters{
					TokenRequestID: loginStartResponse.TokenRequestID,
				},
			),
			client.JSONDecoder(&tr),
			200,
		)

		if err != nil {
			return nil, fmt.Errorf("while fetching token: %w", err)
		}

		if tr.Error == "authorization_pending" {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		if tr.Error != "" {
			return nil, fmt.Errorf("unexpected token error: %s", tr.Error)
		}

		token = tr.AccessToken

		break

	}

	if tctx.Err() != nil {
		return nil, tctx.Err()
	}

	go func() {
		<-ctx.Done()
		ks.Close()
		as.Close()
		err = s.Close()
		if err != nil {
			log.Error(err, "could not shut down server")
		}
		os.RemoveAll(td)
	}()

	return &TestServerInstance{
		s:            s,
		kartuscheURL: ks.URL,
		apiURL:       as.URL,
		adminToken:   token,
	}, nil
}

func (s *TestServerInstance) CreateKartusche(ctx context.Context, name string) error {

	td, err := os.MkdirTemp("", "")

	if err != nil {
		return fmt.Errorf("could not create temp dir: %w", err)
	}

	defer func() {
		os.RemoveAll(td)
	}()

	fileName := filepath.Join(td, "kartusche")

	err = runtime.InitializeEmpty(fileName)
	if err != nil {
		return fmt.Errorf("could not initialize kartusche: %w", err)
	}

	kf, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("could not open kartusche: %w", err)
	}

	defer kf.Close()

	err = client.CallAPI(ctx, s.apiURL, s.adminToken, "PUT", path.Join("kartusches", name), nil, func() (io.Reader, error) { return kf, nil }, nil, 204)
	if err != nil {
		return fmt.Errorf("could not create kartusche: %w", err)
	}

	return nil
}

func (s *TestServerInstance) DumpServer(ctx context.Context) ([]byte, error) {

	var dump []byte

	err := client.CallAPI(ctx, s.apiURL, s.adminToken, "GET", "dump", nil, nil, func(r io.Reader) error {
		var err error
		dump, err = io.ReadAll(r)
		if err != nil {
			return err
		}
		return nil
	}, 200)
	if err != nil {
		return nil, fmt.Errorf("could not get dump: %w", err)
	}

	return dump, nil
}

func (s *TestServerInstance) DeleteKartusche(ctx context.Context, name string) error {
	err := client.CallAPI(ctx, s.apiURL, s.adminToken, "DELETE", path.Join("kartusches", name), nil, nil, nil, 204)
	if err != nil {
		return fmt.Errorf("could not delete kartusche %s: %w", name, err)
	}

	return nil
}
