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
)

func main() {
	var (
		mType    string
		endpoint = "http://localhost:8080/"

		m         storage.MemStorage
		pollCount int64 = 1
	)

	ch := make(chan map[string]interface{})
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	client := &http.Client{}
	for {
		select {
		case <-pollTicker.C:
			go m.Collect(ch, pollCount)
			m.Metrics = <-ch
			pollCount += 1
		case <-reportTicker.C:
			for key, value := range m.Metrics {
				mType = strings.TrimPrefix(reflect.TypeOf(m.Metrics[key]).String(), "storage.")

				builder := strings.Builder{}
				fmt.Fprintf(&builder, "%v/%v/%v",
					mType, key, value)

				request, err := http.NewRequest(http.MethodPost, endpoint+"update/"+builder.String(), nil)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				request.Header.Add("Content-Type", "text/plain")

				response, err := client.Do(request)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Println("Статус-код: ", response.Status)

				_, err = io.ReadAll(response.Body)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				response.Body.Close()
			}
			pollCount = 1
		}
	}
}
