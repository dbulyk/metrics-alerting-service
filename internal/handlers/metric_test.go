package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/cmd/server/config"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/services"

	"github.com/dbulyk/metrics-alerting-service/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	err := os.Chdir("../../")
	if err != nil {
		log.Panic().Timestamp().Err(err).Msg("ошибка смены директории")
		return
	}
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

	err = resp.Body.Close()
	if err != nil {
		log.Panic().Timestamp().Err(err).Msg("ошибка закрытия тела ответа")
	}

	return resp.StatusCode, string(respBody)
}

func TestHandler_GetWithText(t *testing.T) {
	mem := services.NewFileRepository()

	r := chi.NewRouter()
	h := NewRouter(r, &mem)
	h.Register(r)

	ts := httptest.NewServer(r)
	defer ts.Close()

	hash := utils.Hash("testGauge:gauge:123.150000", config.GetKey())
	statusCode, _ := testRequest(t, ts, "POST", "/update/gauge/testGauge/123.15/"+hash, nil)
	assert.Equal(t, http.StatusOK, statusCode)

	hash = utils.Hash("testCounter:counter:123", config.GetKey())
	statusCode, _ = testRequest(t, ts, "POST", "/update/counter/testCounter/123/"+hash, nil)
	assert.Equal(t, http.StatusOK, statusCode)

	testCases := []struct {
		name       string
		mType      string
		mName      string
		statusCode int
	}{
		{
			name:       "testGetGauge",
			mType:      "gauge",
			mName:      "testGauge",
			statusCode: http.StatusOK,
		},
		{
			name:       "testGetCounter",
			mType:      "counter",
			mName:      "testCounter",
			statusCode: http.StatusOK,
		},
		{
			name:       "testGetUnknown",
			mType:      "unknown",
			mName:      "testUnknown",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "testGetWithEmptyName",
			mType:      "",
			mName:      "test",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		statusCode, _ = testRequest(t, ts, "GET", "/value/"+tc.mType+"/"+tc.mName, nil)
		assert.Equal(t, tc.statusCode, statusCode)
	}
}

func TestHandler_GetWithJSON(t *testing.T) {
	mem := services.NewFileRepository()

	r := chi.NewRouter()
	h := NewRouter(r, &mem)
	h.Register(r)

	ts := httptest.NewServer(r)
	defer ts.Close()

	value := 123.15

	m := models.Metric{
		ID:    "testGauge",
		MType: services.Gauge,
		Delta: nil,
		Value: &value,
		Hash:  utils.Hash(fmt.Sprintf("testGauge:gauge:%f", value), config.GetKey()),
	}

	b, err := json.Marshal(m)
	assert.NoError(t, err)

	statusCode, _ := testRequest(t, ts, "POST", "/update/", b)
	assert.Equal(t, http.StatusOK, statusCode)

	jsonData, err := json.Marshal(models.Metric{
		ID:    "testGauge",
		MType: "gauge",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, body := testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusOK, statusCode)

	err = json.Unmarshal([]byte(body), &m)
	assert.NoError(t, err)
	assert.Equal(t, 123.15, *m.Value)

	jsonData, err = json.Marshal(models.Metric{
		ID:    "unknown",
		MType: "gauge",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, _ = testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusNotFound, statusCode)

	jsonData, err = json.Marshal(models.Metric{
		ID:    "",
		MType: "",
		Delta: nil,
		Value: nil,
	})
	assert.NoError(t, err)

	statusCode, _ = testRequest(t, ts, "POST", "/value/", jsonData)
	assert.Equal(t, http.StatusNotFound, statusCode)
}

func TestHandler_UpdateWithText(t *testing.T) {
	mem := services.NewFileRepository()

	r := chi.NewRouter()
	h := NewRouter(r, &mem)
	h.Register(r)

	testCases := []struct {
		name       string
		mType      string
		mName      string
		mValue     string
		mHash      string
		statusCode int
	}{
		{"missing type", "", "testGauge", "1.05", "123", http.StatusNotFound},
		{"missing name", "gauge", "", "1.05", "123", http.StatusNotFound},
		{"missing value", "gauge", "testGauge", "", "123", http.StatusNotFound},
		{"invalid type", "invalid", "testGauge", "1.05", "123", http.StatusNotImplemented},
		{"invalid gauge value", "gauge", "testGauge", "invalid", "123", http.StatusBadRequest},
		{"invalid counter value", "counter", "testCounter", "invalid", "123", http.StatusBadRequest},
		{"correct metric", "gauge", "testGauge", "1.05", utils.Hash("testGauge:gauge:1.050000", config.GetKey()), http.StatusOK},
		{"correct metric", "counter", "testCounter", "123", utils.Hash("testCounter:counter:123", config.GetKey()), http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(r)
			defer ts.Close()

			url := fmt.Sprintf("%s/update/%s/%s/%s/%s", ts.URL, tc.mType, tc.mName, tc.mValue, tc.mHash)

			req, err := http.NewRequest("POST", url, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.statusCode, resp.StatusCode)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if tc.statusCode == http.StatusOK {
				metric, err := mem.Get(ctx, tc.mName, tc.mType)
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

func TestHandler_UpdateWithJSON(t *testing.T) {
	mem := services.NewFileRepository()

	r := chi.NewRouter()
	h := NewRouter(r, &mem)
	h.Register(r)

	delta := int64(12)
	value := 123.15

	testCases := []struct {
		name       string
		body       models.Metric
		statusCode int
	}{
		{
			"correct metric counter",
			models.Metric{
				ID:    "testCounter",
				MType: services.Counter,
				Delta: &delta,
				Value: nil,
				Hash:  utils.Hash(fmt.Sprintf("testCounter:counter:%d", delta), config.GetKey()),
			},
			http.StatusOK,
		},
		{
			"correct metric gauge",
			models.Metric{
				ID:    "testGauge",
				MType: services.Gauge,
				Delta: nil,
				Value: &value,
				Hash:  utils.Hash(fmt.Sprintf("testGauge:gauge:%f", value), config.GetKey()),
			},
			http.StatusOK,
		},
		{
			"incorrect metric with nil value",
			models.Metric{
				ID:    "testIncorrect",
				MType: "incorrect",
				Delta: nil,
				Value: nil,
			},
			http.StatusNotImplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(r)
			defer ts.Close()

			b, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", ts.URL+"/update/", bytes.NewBuffer(b))
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.statusCode, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				var resMetric models.Metric
				err = json.NewDecoder(resp.Body).Decode(&resMetric)
				require.NoError(t, err)

				assert.Equal(t, tc.body, resMetric)
			}
		})
	}
}

func TestHandler_GetAll(t *testing.T) {
	mem := services.NewFileRepository()

	r := chi.NewRouter()
	h := NewRouter(r, &mem)
	h.Register(r)

	ts := httptest.NewServer(r)
	defer ts.Close()

	hash := utils.Hash("testGauge:gauge:123.150000", config.GetKey())
	statusCode, _ := testRequest(t, ts, "POST", "/update/gauge/testGauge/123.15/"+hash, nil)
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/", nil)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.NotEmpty(t, body)
}
