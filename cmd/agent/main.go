package main

import (
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/storage"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	pollInterval   = time.Second * 2
	reportInterval = time.Second * 10
	endpoint       = "http://localhost:8080/"
)

func main() {
	var m storage.MemStorage
	ch := make(chan map[string]interface{})
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)
	client := &http.Client{}

	collectAndSendMetrics(ch, pollTicker, reportTicker, client, &m)
}

func collectAndSendMetrics(
	ch chan map[string]interface{},
	pollTicker *time.Ticker,
	reportTicker *time.Ticker,
	client *http.Client,
	m *storage.MemStorage,
) {
	var (
		pollCount int64 = 1
		builder   strings.Builder
	)
	for {
		select {
		case <-pollTicker.C:
			go m.Collect(ch, pollCount)
			m.Metrics = <-ch
			pollCount += 1
		case <-reportTicker.C:
			for key, value := range m.Metrics {
				request := createRequestToMetricsUpdate(key, value, builder, m)
				response, err := client.Do(request)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				_, err = io.ReadAll(response.Body)
				response.Body.Close()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
			pollCount = 1
		}
	}
}

func createRequestToMetricsUpdate(key string, value interface{}, builder strings.Builder, m *storage.MemStorage) *http.Request {
	mType := strings.TrimPrefix(reflect.TypeOf(m.Metrics[key]).String(), "storage.")

	builder.Reset()
	fmt.Fprintf(&builder, "%v/%v/%v",
		mType, key, value)

	request, err := http.NewRequest(http.MethodPost, endpoint+"update/"+builder.String(), nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	request.Header.Add("Content-Type", "text/plain")
	return request
}
