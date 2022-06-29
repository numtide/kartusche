package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/draganm/bolted"
	"github.com/gofrs/uuid"
)

type authTokenInfo struct {
	CreatedAt time.Time `json:"created_at"`
	UserID    string    `json:"user_id"`
}

type LoginStartResponse struct {
	TokenRequestID  string `json:"token_request_id"`
	VerificationURI string `json:"verification_uri"`
}

func (s *server) loginStart(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		handleHttpError(w, err, s.log)
	}()

	requestID, err := uuid.NewV4()
	if err != nil {
		return
	}

	tokenID, err := uuid.NewV4()
	if err != nil {
		return
	}

	err = bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		requestPath := openTokenRequests.Append(requestID.String())
		tx.CreateMap(requestPath)
		tx.Put(requestPath.Append("token"), []byte(tokenID.String()))
		tx.Put(tokensPath.Append(tokenID.String()), toJSON(authTokenInfo{
			CreatedAt: time.Now(),
			UserID:    "no-auth",
		}))
		return nil
	})

	w.Header().Set("Content-Type", "application/json")

	scheme := r.Header.Get("X-Forwarded-Proto")

	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	verificationURL := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   "/auth/verify",
	}

	json.NewEncoder(w).Encode(LoginStartResponse{
		TokenRequestID:  requestID.String(),
		VerificationURI: verificationURL.String(),
	})

}

type RequestTokenParameters struct {
	TokenRequestID string `json:"token_request_id"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	Error       string `json:"error,omitempty"`
}

func (s *server) accessToken(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		handleHttpError(w, err, s.log)
	}()

	rp := &RequestTokenParameters{}
	err = json.NewDecoder(r.Body).Decode(rp)
	if err != nil {
		err = newErrorWithCode(fmt.Errorf("while decoding request: %w", err), 400)
		return
	}

	resp := &AccessTokenResponse{}

	err = bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) error {
		requestPath := openTokenRequests.Append(rp.TokenRequestID)
		if !tx.Exists(requestPath) {
			return newErrorWithCode(errors.New("request not found"), 404)
		}

		tokenPath := requestPath.Append("token")

		if !tx.Exists(tokenPath) {
			resp.Error = "authorization_pending"
			return nil
		}

		tokenBytes := tx.Get(tokenPath)
		resp.AccessToken = string(tokenBytes)

		return nil
	})

	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
