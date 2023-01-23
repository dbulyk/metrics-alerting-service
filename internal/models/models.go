package models

type (
	Gauge   float64
	Counter int64
	Metric  struct {
		Name  string
		Type  string
		Value interface{}
	}
)
