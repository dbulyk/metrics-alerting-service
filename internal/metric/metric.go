package metric

import (
	"math/rand"
	"reflect"
	"runtime"
)

type (
	gauge   float64
	counter int64
	Metric  struct {
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

func Collect(ch chan Metric, count int64) {
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

	ch <- Metric{
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

func GetNameTypeAndValue(val reflect.Value, fieldIndex int) (string, string, reflect.Value) {
	field := val.Field(fieldIndex)
	return val.Type().Field(fieldIndex).Name, field.Type().String(), field
}
