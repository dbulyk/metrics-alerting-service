package storage

import (
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
)

type Storage interface {
	Collect(chan MemStorage, int64)
	GetNameTypeAndValue(reflect.Value, int) (string, string, reflect.Value)
	SetMetric()
}

type (
	gauge      float64
	counter    int64
	MemStorage struct {
		Alloc,
		BuckHashSys,
		Frees,
		GcCPUFraction,
		GcSys,
		HeapAlloc,
		HeapIdle,
		HeapInuse,
		HeapObjects,
		HeapReleased,
		HeapSys,
		LastGC,
		Lookups,
		MCacheInuse,
		MCacheSys,
		MSpanInuse,
		MSpanSys,
		Mallocs,
		NextGC,
		NumForcedGC,
		NumGC,
		OtherSys,
		PauseTotalNs,
		StackInuse,
		StackSys,
		Sys,
		TotalAlloc,
		RandomValue gauge
		PollCount counter
	}
)

func (m *MemStorage) Collect(ch chan MemStorage, count int64) {
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

	ch <- MemStorage{
		Alloc:         gauge(rtm.Alloc),
		BuckHashSys:   gauge(rtm.BuckHashSys),
		Frees:         gauge(rtm.Frees),
		GcCPUFraction: gauge(rtm.GCCPUFraction),
		GcSys:         gauge(rtm.GCSys),
		HeapAlloc:     gauge(rtm.HeapAlloc),
		HeapIdle:      gauge(rtm.HeapIdle),
		HeapInuse:     gauge(rtm.HeapInuse),
		HeapObjects:   gauge(rtm.HeapObjects),
		HeapReleased:  gauge(rtm.HeapReleased),
		HeapSys:       gauge(rtm.HeapSys),
		LastGC:        gauge(rtm.LastGC),
		Lookups:       gauge(rtm.Lookups),
		MCacheInuse:   gauge(rtm.MCacheInuse),
		MCacheSys:     gauge(rtm.MCacheSys),
		MSpanInuse:    gauge(rtm.MSpanInuse),
		MSpanSys:      gauge(rtm.MSpanSys),
		Mallocs:       gauge(rtm.Mallocs),
		NextGC:        gauge(rtm.NextGC),
		NumForcedGC:   gauge(rtm.NumForcedGC),
		NumGC:         gauge(rtm.NumGC),
		OtherSys:      gauge(rtm.OtherSys),
		PauseTotalNs:  gauge(rtm.PauseTotalNs),
		StackInuse:    gauge(rtm.StackInuse),
		StackSys:      gauge(rtm.StackSys),
		Sys:           gauge(rtm.Sys),
		TotalAlloc:    gauge(rtm.TotalAlloc),
		PollCount:     counter(count),
		RandomValue:   gauge(rand.Float64()),
	}
}

func (m *MemStorage) GetNameTypeAndValue(val reflect.Value, fieldIndex int) (string, string, reflect.Value) {
	field := val.Field(fieldIndex)
	return val.Type().Field(fieldIndex).Name, field.Type().String(), field
}

func (m *MemStorage) SetMetric(fType string, fName string, fValue float64) int {
	a := reflect.ValueOf(m).Elem()
	b := a.FieldByName(fName)

	switch fType {
	case "counter":
		b.Set(reflect.ValueOf(counter(b.Int()) + counter(fValue)))
	case "gauge":
		b.Set(reflect.ValueOf(gauge(fValue)))
	default:
		return http.StatusNotImplemented
	}
	return http.StatusOK
}
