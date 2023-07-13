package metric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	mem := NewRepository()
	r := NewRouter(mem)
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
	mem := NewRepository()
	r, _ := NewRouter(mem)
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/counter/testCounter2/10", nil)
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/value/counter/testCounter2", nil)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "10", body)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/testCounter2/15", nil)
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body = testRequest(t, ts, "GET", "/value/counter/testCounter2", nil)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "25", body)

	statusCode, _ = testRequest(t, ts, "POST", "/update/unknown/testCounter2/15", nil)
	assert.Equal(t, http.StatusNotImplemented, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/unknown/testCounter2/15", nil)
	assert.Equal(t, http.StatusNotImplemented, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/15/", nil)
	assert.Equal(t, http.StatusNotFound, statusCode)

	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/testCounter2/invalid", nil)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func TestUpdateWithJSON(t *testing.T) {
	mem := NewRepository()
	r, _ := NewRouter(mem)
	ts := httptest.NewServer(r)
	defer ts.Close()

	delta := int64(20)
	hash := utils.Hash(fmt.Sprintf("%s:counter:%d", "testCounter1", &delta), "test")
	jsonData, err := json.Marshal(Metric{
		ID:    "testCounter1",
		MType: "counter",
		Delta: &delta,
		Value: nil,
		Hash:  hash,
	})
	assert.NoError(t, err)

	statusCode, _ := testRequest(t, ts, "POST", "/update/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)

	jsonData, err = json.Marshal(Metric{
		ID:    "testCounter1",
		MType: "counter",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, body := testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)

	var m Metric
	err = json.Unmarshal([]byte(body), &m)
	assert.NoError(t, err)
	assert.Equal(t, int64(20), *m.Delta)

	delta = int64(15)
	jsonData, err = json.Marshal(Metric{
		ID:    "testCounter1",
		MType: "counter",
		Delta: &delta,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, _ = testRequest(t, ts, "POST", "/update/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)

	jsonData, err = json.Marshal(Metric{
		ID:    "",
		MType: "",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, _ = testRequest(t, ts, "POST", "/update/", jsonData)
	assert.Equal(t, http.StatusNotImplemented, statusCode)

	jsonData, err = json.Marshal(Metric{
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
	mem := NewRepository()
	r, _ := NewRouter(mem)
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

func TestGetWithJSON(t *testing.T) {
	mem := NewRepository()
	r, _ := NewRouter(mem)
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/gauge/testGauge/123.15", nil)
	assert.Equal(t, http.StatusOK, statusCode)

	jsonData, err := json.Marshal(Metric{
		ID:    "testGauge",
		MType: "gauge",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, body := testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)

	var m Metric
	err = json.Unmarshal([]byte(body), &m)
	assert.NoError(t, err)
	assert.Equal(t, float64(123.15), *m.Value)

	jsonData, err = json.Marshal(Metric{
		ID:    "unknown",
		MType: "gauge",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, _ = testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusNotFound, statusCode)

	jsonData, err = json.Marshal(Metric{
		ID:    "",
		MType: "",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, _ = testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusNotFound, statusCode)
}

func TestUpdateWithTextIncorrect(t *testing.T) {
	mem := NewRepository()

	router := chi.NewRouter()
	router.Post("/{type}/{name}/{value}", UpdateWithText)

	testCases := []struct {
		name       string
		mType      string
		mName      string
		mValue     string
		statusCode int
	}{
		{"missing type", "", "testGauge", "1.05", http.StatusNotFound},
		{"missing name", "gauge", "", "1.05", http.StatusNotFound},
		{"missing value", "gauge", "testGauge", "", http.StatusNotFound},
		{"invalid type", "invalid", "testGauge", "1.05", http.StatusNotImplemented},
		{"invalid gauge value", "gauge", "testGauge", "invalid", http.StatusBadRequest},
		{"invalid counter value", "counter", "testCounter", "invalid", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(router)
			defer ts.Close()

			url := fmt.Sprintf("%s/%s/%s/%s", ts.URL, tc.mType, tc.mName, tc.mValue)

			req, err := http.NewRequest("POST", url, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.statusCode, resp.StatusCode)

			if tc.statusCode == http.StatusOK {
				metric, err := mem.GetMetric(tc.mName, tc.mType)
				require.NoError(t, err)

				if tc.mType == "gauge" {
					value, err := strconv.ParseFloat(tc.mValue, 64)
					require.NoError(t, err)
					assert.Equal(t, value, *metric.Value)
				} else if tc.mType == "counter" {
					value, err := strconv.ParseInt(tc.mValue, 0, 64)
					require.NoError(t, err)
					assert.Equal(t, value, *metric.Delta)
				}
			}
		})
	}
}
