// Package rest is a driving (inbound) adapter: it exposes the core's
// KeyValueStore port over HTTP.
//
// It depends only on the inbound port, never on a concrete service or on any
// storage engine. Translating HTTP <-> domain is its single responsibility;
// all business rules live in the core.
package rest

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/luckybaraka/hazinadb/internal/core/domain"
	"github.com/luckybaraka/hazinadb/internal/core/port"
)

// Server adapts HTTP requests to the KeyValueStore port.
type Server struct {
	store port.KeyValueStore
	http  *http.Server
}

// NewServer builds an HTTP server bound to addr and backed by the given store.
func NewServer(addr string, store port.KeyValueStore) *Server {
	s := &Server{store: store}

	// Method-based routing (Go 1.22+): the mux matches verb + path pattern and
	// extracts {key} for us, replacing manual method switches and path slicing.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /kv/{key}", s.handleGet)
	mux.HandleFunc("PUT /kv/{key}", s.handlePut)
	mux.HandleFunc("DELETE /kv/{key}", s.handleDelete)

	s.http = &http.Server{Addr: addr, Handler: mux}
	return s
}

// ListenAndServe starts serving and blocks until the server is shut down.
func (s *Server) ListenAndServe() error {
	return s.http.ListenAndServe()
}

// Shutdown gracefully drains in-flight requests before stopping.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// handleGet serves GET /kv/{key}.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := domain.Key(r.PathValue("key"))

	value, err := s.store.Get(r.Context(), key)
	if err != nil {
		s.writeError(w, err)
		return
	}
	w.Write(value)
}

// handlePut serves PUT /kv/{key}; the request body is the value.
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}

	key := domain.Key(r.PathValue("key"))
	if err := s.store.Put(r.Context(), key, value); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDelete serves DELETE /kv/{key}.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := domain.Key(r.PathValue("key"))

	if err := s.store.Delete(r.Context(), key); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeError maps domain errors to HTTP status codes in one place, so every
// handler reports failures consistently.
func (s *Server) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrKeyNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrEmptyKey):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
