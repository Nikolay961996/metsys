package router

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/Nikolay961996/metsys/internal/crypto"
	"github.com/Nikolay961996/metsys/models"
	"go.uber.org/zap"
)

func WithDecrypt(privateKey *rsa.PrivateKey) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				models.Log.Error("error read request body", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			decrypted, err := crypto.DecryptMessageWithPrivateKey(body, privateKey)
			if err != nil {
				models.Log.Error("error decrypt message", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			next.ServeHTTP(w, r)
		})
	}
}
