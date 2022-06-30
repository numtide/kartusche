package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/draganm/kartusche/common/auth"
)

func CallAPI(baseURL, method, pth string, q url.Values, bodyEncoder func() (io.Reader, error), response interface{}, expectedStatus int) error {

	var tkn string
	if !(strings.HasPrefix(pth, "/auth/") || strings.HasPrefix(pth, "auth/")) {
		var err error
		tkn, err = auth.GetTokenForServer(baseURL)
		if err != nil {
			return err
		}
	}

	bu, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("while parsing server base url: %w", err)
	}

	bu.Path = path.Join(bu.Path, pth)
	if q != nil {
		bu.RawQuery = q.Encode()
	}

	var body io.Reader

	if bodyEncoder != nil {
		body, err = bodyEncoder()
		if err != nil {
			return fmt.Errorf("while encoding body: %w", err)
		}
	}

	req, err := http.NewRequest(method, bu.String(), body)
	if err != nil {
		return fmt.Errorf("while creating request: %w", err)
	}

	if tkn != "" {
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", tkn))
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("while performing %s %s request: %w", method, pth, err)
	}

	defer res.Body.Close()

	if res.StatusCode != expectedStatus {
		bod, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("while reading response body: %w", err)
		}
		return fmt.Errorf("unexpected status %s: %s", res.Status, string(bod))
	}

	if response == nil {
		return nil
	}

	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("while decoding response: %w", err)
	}

	return nil

}
