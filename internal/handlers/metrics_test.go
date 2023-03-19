package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/stores"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Chdir("../../")
	code := m.Run()
	os.Exit(code)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, jsonData []byte) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(jsonData))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestRouter(t *testing.T) {
	mem := stores.NewMemStorage()
	r, _, _ := MetricsRouter(mem)
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/gauge/testGauge/123", nil)
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/value/gauge/testGauge", nil)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "123", body)

	statusCode, body = testRequest(t, ts, "GET", "/", nil)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Contains(t, body, "testGauge")
}

func TestUpdateWithText(t *testing.T) {
	mem := stores.NewMemStorage()
	r, _, _ := MetricsRouter(mem)
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/counter/testCounter/10", nil)
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/value/counter/testCounter", nil)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "10", body)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/testCounter/15", nil)
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body = testRequest(t, ts, "GET", "/value/counter/testCounter", nil)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "25", body)

	statusCode, _ = testRequest(t, ts, "POST", "/update/unknown/testCounter/15", nil)
	assert.Equal(t, http.StatusNotImplemented, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/unknown/testCounter/15", nil)
	assert.Equal(t, http.StatusNotImplemented, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/15/", nil)
	assert.Equal(t, http.StatusNotFound, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/testCounter/invalid", nil)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func TestUpdateWithJSON(t *testing.T) {
	mem := stores.NewMemStorage()
	r, _, _ := MetricsRouter(mem)
	ts := httptest.NewServer(r)
	defer ts.Close()

	delta := int64(20)
	jsonData, err := json.Marshal(models.Metrics{
		ID:    "testCounter1",
		MType: "counter",
		Delta: &delta,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, _ := testRequest(t, ts, "POST", "/update/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)

	jsonData, err = json.Marshal(models.Metrics{
		ID:    "testCounter1",
		MType: "counter",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, body := testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)

	var m models.Metrics
	err = json.Unmarshal([]byte(body), &m)
	assert.NoError(t, err)
	assert.Equal(t, int64(20), *m.Delta)

	delta = int64(15)
	jsonData, err = json.Marshal(models.Metrics{
		ID:    "testCounter1",
		MType: "counter",
		Delta: &delta,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, _ = testRequest(t, ts, "POST", "/update/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)

	jsonData, err = json.Marshal(models.Metrics{
		ID:    "testCounter1",
		MType: "counter",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, body = testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)
	err = json.Unmarshal([]byte(body), &m)
	assert.NoError(t, err)
	assert.Equal(t, int64(35), *m.Delta)
}

func TestGetWithText(t *testing.T) {
	mem := stores.NewMemStorage()
	r, _, _ := MetricsRouter(mem)
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/gauge/testGauge/123.15", nil)
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/value/gauge/testGauge", nil)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "123.15", body)

	statusCode, _ = testRequest(t, ts, "GET", "/value/gauge/unknown", nil)
	assert.Equal(t, http.StatusNotFound, statusCode)
}
