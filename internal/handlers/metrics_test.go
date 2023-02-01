package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestRouter(t *testing.T) {
	r := MetricsRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/gauge/testGauge/123")
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/value/gauge/testGauge")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "123", body)

	//statusCode, body = testRequest(t, ts, "GET", "/")
	//assert.Equal(t, http.StatusOK, statusCode)
	//assert.Len(t, body, 1)
}

func TestUpdate(t *testing.T) {
	r := MetricsRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/counter/testCounter/10")
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/value/counter/testCounter")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "10", body)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/testCounter/15")
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body = testRequest(t, ts, "GET", "/value/counter/testCounter")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "25", body)

	statusCode, _ = testRequest(t, ts, "POST", "/update/unknown/testCounter/15")
	assert.Equal(t, http.StatusNotImplemented, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/unknown/testCounter/15")
	assert.Equal(t, http.StatusNotImplemented, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/15/")
	assert.Equal(t, http.StatusNotFound, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/testCounter/invalid")
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func TestGet(t *testing.T) {
	r := MetricsRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/gauge/testGauge/123.15")
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/value/gauge/testGauge")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "123.15", body)

	statusCode, body = testRequest(t, ts, "GET", "/value/gauge/unknown")
	assert.Equal(t, http.StatusNotFound, statusCode)
}
