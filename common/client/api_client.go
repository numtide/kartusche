package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

func CallAPI(baseURL, method, pth string, bodyEncoder func() (io.Reader, error), response interface{}, expectedStatus int) error {
	bu, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("while parsing server base url: %w", err)
	}

	bu.Path = path.Join(bu.Path, pth)

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

	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("while decoding response")
	}

	return nil

}
