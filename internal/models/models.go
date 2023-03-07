package models

type (
	Gauge   float64
	Counter int64
	Metric  struct {
		Name  string
		Type  string
		Value interface{}
	}
	Metrics struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}
)
