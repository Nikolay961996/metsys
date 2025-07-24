package router

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/Nikolay961996/metsys/models"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func WithSigningCheck(keyForSigning string) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sign := r.Header.Get("HashSHA256")
			if sign == "" {
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
			r.Body = io.NopCloser(bytes.NewReader(body))

			h := hmac.New(sha256.New, []byte(keyForSigning))
			h.Write(body)
			expected := hex.EncodeToString(h.Sum(nil))
			if !hmac.Equal([]byte(expected), []byte(sign)) {
				http.Error(w, "sign not valid", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func WithSigningResponse(keyForSigning string) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := responseRecorder{
				ResponseWriter: w,
				body:           []byte{},
			}
			next.ServeHTTP(&recorder, r)

			if keyForSigning != "" && len(recorder.body) > 0 {
				h := hmac.New(sha256.New, []byte(keyForSigning))
				h.Write(recorder.body)
				signature := hex.EncodeToString(h.Sum(nil))
				w.Header().Set("HashSHA256", signature)
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	body []byte
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}
