package model

import (
	"net/http"
	"sync"
)

type Agent struct {
	client *http.Client
	mu     sync.Mutex
}
