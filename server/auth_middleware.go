package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/draganm/bolted"
)

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log := s.log

		requireAuthentication := func(reason string) {
			log.Info("auth failed", "reason", reason)
			w.Header().Set("WWW-Authenticate", `Basic realm="Kartusche"`)
			http.Error(w, "not authorized", http.StatusUnauthorized)
		}

		if strings.HasPrefix(r.URL.Path, "/auth") {
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			requireAuthentication("no authorization header")
			return
		}

		before, after, found := strings.Cut(auth, " ")
		if !found {
			requireAuthentication("authorization header malformed")
			return
		}

		var tkn string

		switch strings.ToLower(before) {
		case "bearer":
			tkn = after
		case "basic":
			decoded, err := base64.StdEncoding.DecodeString(after)
			if err != nil {
				requireAuthentication("basic auth malformed")
				return
			}
			_, password, found := strings.Cut(string(decoded), ":")
			if !found {
				requireAuthentication("basic auth malformed")
				return
			}
			tkn = password
		default:
			requireAuthentication(fmt.Sprintf("unsupported auth method %s", before))
			return
		}

		tokenValid := false

		err := bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) error {
			tokenValid = tx.Exists(tokensPath.Append(tkn))
			return nil
		})

		if err != nil {
			log.Error(err, "could not check token")
			http.Error(w, err.Error(), 500)
			return
		}

		if !tokenValid {
			requireAuthentication("token invalid")
			return
		}

		next.ServeHTTP(w, r)

	})
}
