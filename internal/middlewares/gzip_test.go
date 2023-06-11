package middlewares

import (
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	r := chi.NewRouter()

	r.Use(GzipMiddleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	reader, err := gzip.NewReader(resp.Body)
	assert.NoError(t, err)
	defer reader.Close()

	body, err := io.ReadAll(reader)
	assert.NoError(t, err)

	assert.Equal(t, "Hello, World!", string(body))
}
