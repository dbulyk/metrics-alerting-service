package main

import (
	"fmt"
	"github.com/dbulyk/metrics-alerting-service/internal/metric"
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
		mName,
		mType string
		mValue   interface{}
		endpoint = "http://127.0.0.1:8080/"

		m         metric.Metric
		pollCount int64 = 1
	)

	ch := make(chan metric.Metric)
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	client := &http.Client{}
	for {
		select {
		case <-pollTicker.C:
			go metric.Collect(ch, pollCount)
			m = <-ch
			pollCount += 1

		case <-reportTicker.C:
			val := reflect.ValueOf(m)
			for fieldIndex := 0; fieldIndex < val.NumField(); fieldIndex++ {
				w := strings.Builder{}
				mName, mType, mValue = metric.GetNameTypeAndValue(val, fieldIndex)

				_, err := fmt.Fprintf(&w, "%vupdate/%v/%v/%v", endpoint,
					strings.TrimPrefix(mType, "metric."), mName, mValue)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				request, err := http.NewRequest(http.MethodPost, endpoint, nil)
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

				err = response.Body.Close()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			pollCount = 1
		}
	}
}
