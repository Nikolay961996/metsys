package router

import (
	"github.com/Nikolay961996/metsys/models"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type (
	responseDate struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		data *responseDate
	}
)

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.data.status = status
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.data.size += size
	return size, err
}

func WithLogger(h http.Handler) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		rd := &responseDate{
			status: 200,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			data:           rd,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		models.Log.Info("request log",
			zap.String("method", method),
			zap.String("uri", uri),
			zap.Duration("duration", duration),
			zap.Int("status", lw.data.status),
			zap.Int("size", lw.data.size),
		)
	}
	return logFn
}
