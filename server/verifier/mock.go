package verifier

import (
	"net/http"
	"net/url"
)

type mockProvider struct{}

func (m *mockProvider) Verify(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	scheme := r.Header.Get("X-Forwarded-Proto")

	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	// q := r.URL.Query()

	q := url.Values{}

	q.Set("code", r.URL.Query().Get("request_id"))
	verificationURL := &url.URL{
		Scheme:   scheme,
		Host:     r.Host,
		Path:     "/auth/oauth2/callback",
		RawQuery: q.Encode(),
	}

	http.Redirect(w, r, verificationURL.String(), 302)
}

func (m *mockProvider) Callback(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
	w.Write([]byte("authentication successful"))
	return &AuthResult{
		Code:  r.URL.Query().Get("code"),
		Email: "mock@kartusche.com",
	}, nil
}

func NewMockProvider() AuthenticationProvider {
	return &mockProvider{}
}
