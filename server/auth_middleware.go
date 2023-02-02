package server

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/draganm/bolted"
)

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.HasPrefix(r.URL.Path, "/auth") {
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		before, after, found := strings.Cut(auth, " ")
		if !found {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		if strings.ToLower(before) != "bearer" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		var tkn string

		switch strings.ToLower(after) {
		case "bearer":
			tkn = after
		case "basic":
			decoded, err := base64.StdEncoding.DecodeString(after)
			if err != nil {
				http.Error(w, "not authorized", http.StatusUnauthorized)
				return
			}
			_, password, found := strings.Cut(string(decoded), ":")
			if !found {
				http.Error(w, "not authorized", http.StatusUnauthorized)
				return
			}
			tkn = password
		}

		tokenValid := false

		err := bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) error {
			tokenValid = tx.Exists(tokensPath.Append(tkn))
			return nil
		})

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if !tokenValid {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)

	})
}
