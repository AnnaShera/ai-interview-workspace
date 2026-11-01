package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

// newMux registers the todo routes. Kept separate from main so tests can serve it with httptest.
func newMux(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /todos", h.Create)
	mux.HandleFunc("GET /todos", h.List)
	mux.HandleFunc("GET /todos/{id}", h.Get)
	mux.HandleFunc("DELETE /todos/{id}", h.Delete)
	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           newMux(NewHandlers(NewMemoryStore())),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
