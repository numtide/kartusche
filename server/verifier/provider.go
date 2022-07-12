package verifier

import "net/http"

type AuthResult struct {
	Code  string
	Email string
}

type AuthenticationProvider interface {
	Verify(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request) (*AuthResult, error)
}
