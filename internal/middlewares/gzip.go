package middlewares

import (
	"compress/gzip"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

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
				log.Error().Err(err).Msg("ошибка закрытия gzip.Writer")
			}
		}(gzWriter)

		gzResponse := gzipWriter{Writer: gzWriter, ResponseWriter: w}
		next.ServeHTTP(gzResponse, r)
	})
}

func (gzResponse gzipWriter) Write(b []byte) (int, error) {
	return gzResponse.Writer.Write(b)
}

func (gzResponse gzipWriter) Header() http.Header {
	return gzResponse.ResponseWriter.Header()
}

func (gzResponse gzipWriter) WriteHeader(statusCode int) {
	gzResponse.ResponseWriter.WriteHeader(statusCode)
}
