package router

import (
	"compress/gzip"
	"github.com/Nikolay961996/metsys/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"sync"
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

func WithLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}

type compressedWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *compressedWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

func WithDecompressionRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz := gzipReaderPool.Get().(*gzip.Reader)
			err := gz.Reset(r.Body)
			if err != nil {
				models.Log.Error("error creating gzip reader", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			defer func() {
				gz.Close()
				gzipReaderPool.Put(gz)
			}()
			r.Body = gz
		}
		h.ServeHTTP(w, r)
	})
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		gz, _ := gzip.NewWriterLevel(io.Discard, gzip.BestCompression)
		return gz
	},
}

func WithCompressionResponse(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzipWriterPool.Get().(*gzip.Writer)
			gz.Reset(w)
			w.Header().Add("Content-Encoding", "gzip")
			defer func() {
				gz.Close()
				gzipWriterPool.Put(gz)
			}()
			w = &compressedWriter{w, gz}
		}
		h.ServeHTTP(w, r)
	}
}
