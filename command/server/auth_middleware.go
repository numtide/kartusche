package server

import (
	"net/http"
	"strings"

	"github.com/draganm/bolted"
)

func (s *server) authMiddleware(next http.Handler) http.Handler {
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

		parts := strings.Split(auth, " ")
		if len(parts) != 2 {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		if strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		tkn := parts[1]

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

		// TODO: add user info to the request context

		next.ServeHTTP(w, r)

	})
}
