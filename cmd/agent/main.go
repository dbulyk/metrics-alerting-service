package main

import (
	"bytes"
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
		mName,
		mType string
		mValue   interface{}
		endpoint = "http://127.0.0.1:8080/"

		m         storage.MemStorage
		pollCount int64 = 1
	)

	ch := make(chan storage.MemStorage)
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	client := &http.Client{}
	for {
		select {
		case <-pollTicker.C:
			go m.Collect(ch, pollCount)
			m = <-ch
			pollCount += 1
		case <-reportTicker.C:
			val := reflect.ValueOf(m)
			for fieldIndex := 0; fieldIndex < val.NumField(); fieldIndex++ {
				mName, mType, mValue = m.GetNameTypeAndValue(val, fieldIndex)

				builder := strings.Builder{}
				fmt.Fprintf(&builder, "%v/%v/%v",
					strings.TrimPrefix(mType, "storage."), mName, mValue)

				request, err := http.NewRequest(http.MethodPost, endpoint+"update/",
					bytes.NewBufferString(builder.String()))
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
