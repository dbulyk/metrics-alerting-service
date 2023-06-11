package middlewares

import (
	"compress/gzip"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	rw http.ResponseWriter
	w  *gzip.Writer
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

		gzResponse := gzipResponseWriter{w: gzWriter, rw: w}
		next.ServeHTTP(gzResponse, r)
	})
}

func (gzResponse gzipResponseWriter) Write(b []byte) (int, error) {
	return gzResponse.w.Write(b)
}

func (gzResponse gzipResponseWriter) Header() http.Header {
	return gzResponse.rw.Header()
}

func (gzResponse gzipResponseWriter) WriteHeader(statusCode int) {
	gzResponse.rw.WriteHeader(statusCode)
}
