package storage

import (
	"math/rand"
	"net/http"
	"runtime"
	"sync"
)

type Storage interface {
	Collect(ch chan map[string]interface{}, count int64)
	SetMetric(mType string, mName string, mValue float64) (status int)
}

type (
	gauge      float64
	counter    int64
	MemStorage struct {
		Metrics map[string]interface{}
	}
)

var (
	rtm runtime.MemStats
	mu  sync.Mutex
)

func (m *MemStorage) Collect(ch chan map[string]interface{}, count int64) {
	mu.Lock()
	defer mu.Unlock()
	runtime.ReadMemStats(&rtm)

	ch <- map[string]interface{}{
		"Alloc":         gauge(rtm.Alloc),
		"BuckHashSys":   gauge(rtm.BuckHashSys),
		"Frees":         gauge(rtm.Frees),
		"GcCPUFraction": gauge(rtm.GCCPUFraction),
		"GcSys":         gauge(rtm.GCSys),
		"HeapAlloc":     gauge(rtm.HeapAlloc),
		"HeapIdle":      gauge(rtm.HeapIdle),
		"HeapInuse":     gauge(rtm.HeapInuse),
		"HeapObjects":   gauge(rtm.HeapObjects),
		"HeapReleased":  gauge(rtm.HeapReleased),
		"HeapSys":       gauge(rtm.HeapSys),
		"LastGC":        gauge(rtm.LastGC),
		"Lookups":       gauge(rtm.Lookups),
		"MCacheInuse":   gauge(rtm.MCacheInuse),
		"MCacheSys":     gauge(rtm.MCacheSys),
		"MSpanInuse":    gauge(rtm.MSpanInuse),
		"MSpanSys":      gauge(rtm.MSpanSys),
		"Mallocs":       gauge(rtm.Mallocs),
		"NextGC":        gauge(rtm.NextGC),
		"NumForcedGC":   gauge(rtm.NumForcedGC),
		"NumGC":         gauge(rtm.NumGC),
		"OtherSys":      gauge(rtm.OtherSys),
		"PauseTotalNs":  gauge(rtm.PauseTotalNs),
		"StackInuse":    gauge(rtm.StackInuse),
		"StackSys":      gauge(rtm.StackSys),
		"Sys":           gauge(rtm.Sys),
		"TotalAlloc":    gauge(rtm.TotalAlloc),
		"PollCount":     counter(count),
		"RandomValue":   gauge(rand.Float64()),
	}
}

func (m *MemStorage) SetMetric(mType string, mName string, mValue float64) (status int) {
	if m.Metrics == nil {
		m.Metrics = make(map[string]interface{}, 5)
	}

	switch mType {
	case "counter":
		if val, ok := m.Metrics[mName]; !ok {
			m.Metrics[mName] = counter(mValue)
		} else {
			m.Metrics[mName] = val.(counter) + counter(mValue)
		}
	case "gauge":
		m.Metrics[mName] = gauge(mValue)
	default:
		return http.StatusNotImplemented
	}
	return http.StatusOK
}
