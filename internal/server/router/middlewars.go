// Package router consist middlewars
package router

import (
	"compress/gzip"
	"fmt"
	"google.golang.org/grpc/codes"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Nikolay961996/metsys/models"
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

func checkTrustedSubnet(xRealIP string, trustedSubnet string) codes.Code {
	if trustedSubnet == "" {
		return codes.OK
	}

	if xRealIP == "" {
		models.Log.Error("X-Real-IP header is missing")
		return codes.PermissionDenied // Forbidden
	}

	ip := net.ParseIP(xRealIP)
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		models.Log.Error("Invalid trusted subnet configuration")
		return codes.Internal
	}

	if !subnet.Contains(ip) {
		return codes.PermissionDenied
	}

	return codes.OK
}

func WithTrustedSubnetValidation(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xRealIP := r.Header.Get("X-Real-IP")
			code := checkTrustedSubnet(xRealIP, trustedSubnet)

			switch code {
			case codes.PermissionDenied:
				http.Error(w, "Forbidden: IP not in trusted subnet", http.StatusForbidden)
			case codes.Internal:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			case codes.OK:
				next.ServeHTTP(w, r)
			default:
				models.Log.Warn(fmt.Sprintf("Not expectes code: %d", code))
				next.ServeHTTP(w, r)
			}
		})
	}
}
