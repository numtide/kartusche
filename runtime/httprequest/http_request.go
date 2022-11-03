package httprequest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
)

type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       interface{}
}

type Options struct {
	// Headers map[string]string
	// Query   map[string]string
	Body interface{}
	Json bool
}

func Request(method, url string, options Options) (*Response, error) {
	bb := new(bytes.Buffer)
	if options.Json {
		err := json.NewEncoder(bb).Encode(options.Body)
		if err != nil {
			return nil, fmt.Errorf("while serializing JSON body: %w", err)
		}
	}

	// if options.Query

	req, err := http.NewRequest(method, url, bb)
	if err != nil {
		return nil, fmt.Errorf("while creating new request: %w", err)
	}

	if options.Json {
		req.Header.Set("content-type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("while performing request: %w", err)
	}

	defer res.Body.Close()

	hct, err := hasContentType(res, "application/json")
	if err != nil {
		return nil, err
	}

	r := &Response{}

	if hct {
		err = json.NewDecoder(res.Body).Decode(&r.Body)
		if err != nil {
			return nil, fmt.Errorf("while decoding JSON response: %w", err)
		}
	}

	headerCopy := http.Header{}
	for k, v := range res.Header {
		headerCopy[k] = v
	}

	r.Headers = headerCopy
	r.Status = res.Status
	r.StatusCode = res.StatusCode

	return r, nil

}

func hasContentType(r *http.Response, mt string) (bool, error) {
	ct := r.Header.Get("Content-type")

	if ct == "" {
		return false, nil
	}

	cmt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false, fmt.Errorf("while checking for %s content type: %w", mt, err)
	}

	return cmt == mt, nil
}
