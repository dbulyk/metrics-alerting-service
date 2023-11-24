package handlers

import "github.com/go-chi/chi/v5"

// Handler is an interface for all handlers
type Handler interface {
	Register(router *chi.Mux)
}
