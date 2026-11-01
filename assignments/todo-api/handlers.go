package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	maxTitleRunes = 100
	maxBodyBytes  = 1 << 20 // 1 MiB
)

// Handlers serves the todo endpoints on top of a Store.
type Handlers struct {
	store Store
}

// NewHandlers returns handlers backed by the given store.
func NewHandlers(store Store) *Handlers {
	return &Handlers{store: store}
}

type createRequest struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Create handles POST /todos.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createRequest
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	// A second Decode must hit EOF, otherwise there is trailing data after the object.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "body must contain a single JSON object")
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if utf8.RuneCountInString(title) > maxTitleRunes {
		writeError(w, http.StatusBadRequest, "title must be at most 100 characters")
		return
	}

	todo, err := h.store.Create(r.Context(), title, req.Done)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, todo)
}

// Get handles GET /todos/{id}.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	todo, err := h.store.Get(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

// List handles GET /todos.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.List(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

// Delete handles DELETE /todos/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.store.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseID reads the {id} path value. It writes a 400 and returns false if it is not an integer.
func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be an integer")
		return 0, false
	}
	return id, true
}

// writeStoreError maps store errors to HTTP. Unknown errors are logged, never exposed.
func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	slog.Error("store error", "err", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}
