package models

// Metric is a struct for metrics.
type Metric struct {
	ID    string   `json:"id" example:"metric_name"`
	MType string   `json:"type" example:"counter"`
	Delta *int64   `json:"delta,omitempty" example:"1"`
	Value *float64 `json:"value,omitempty" example:"1.0"`
	Hash  string   `json:"hash,omitempty" example:"hash"`
}
