package server

import (
	"net/http"
	"time"

	"github.com/draganm/bolted"
	"github.com/gofrs/uuid"
)

func (s *server) authVerify(w http.ResponseWriter, r *http.Request) {
	s.verifier.Verify(w, r)
}

func (s *server) authOauth2Callback(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		handleHttpError(w, err, s.log)
	}()

	res, err := s.verifier.Callback(w, r)
	if err != nil {
		return
	}

	tokenID, err := uuid.NewV4()
	if err != nil {
		return
	}
	err = bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		requestPath := openTokenRequests.Append(res.Code)
		tx.Put(requestPath.Append("token"), []byte(tokenID.String()))
		tx.Put(tokensPath.Append(tokenID.String()), toJSON(authTokenInfo{
			CreatedAt: time.Now(),
			UserID:    "no-auth",
		}))
		return nil
	})
}
