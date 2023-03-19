package models

import (
	"encoding/json"
	"os"
)

type (
	Metrics struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}

	Consumer struct {
		file    *os.File
		decoder json.Decoder
	}
)
