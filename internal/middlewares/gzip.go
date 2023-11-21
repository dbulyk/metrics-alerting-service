package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

// GzipMiddleware compresses HTTP response using gzip.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gzWriter := gzip.NewWriter(w)

		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				log.Error().Err(err).Msg("error closing gzip.Writer")
			}
		}(gzWriter)

		gzResponse := gzipWriter{Writer: gzWriter, ResponseWriter: w}
		next.ServeHTTP(gzResponse, r)
	})
}

// Write writes compressed data to the gzip.Writer.
func (gzResponse gzipWriter) Write(b []byte) (int, error) {
	return gzResponse.Writer.Write(b)
}

// Header returns the header of the gzip.ResponseWriter.
func (gzResponse gzipWriter) Header() http.Header {
	return gzResponse.ResponseWriter.Header()
}

// WriteHeader writes the header to the gzip.ResponseWriter.
func (gzResponse gzipWriter) WriteHeader(statusCode int) {
	gzResponse.ResponseWriter.WriteHeader(statusCode)
}
