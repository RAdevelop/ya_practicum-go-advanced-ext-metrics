package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/sign"
)

// SingCheck - проверяет, есть ли заголовок "HashSHA256" со значением, если есть, то проверяет подпись
func SingCheck(serverContext *server.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		receivedHash := r.Header.Get("HashSHA256")
		if receivedHash == "" {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			serverContext.Logger.Error("cannot read body", "error", err)
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		if !sign.SHA256Verify(body, serverContext.Config.SignKey(), receivedHash) {
			http.Error(w, "cannot verify sign", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		next.ServeHTTP(w, r)
	})
}
