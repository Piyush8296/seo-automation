package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Server wires together storage, the hub, the manager, and the HTTP mux.
type Server struct {
	h    *Handlers
	addr string
}

// New creates a Server ready to serve.
// baseDir is where audit reports are stored; defaults to ~/.seo-reports when empty.
func New(baseDir string) (*Server, error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home dir: %w", err)
		}
		baseDir = filepath.Join(home, ".seo-reports")
	}

	storage, err := NewStorage(baseDir)
	if err != nil {
		return nil, fmt.Errorf("init storage at %q: %w", baseDir, err)
	}

	hub := NewHub()
	manager := NewManager(storage, hub)
	handlers := newHandlers(manager, storage, hub)

	return &Server{h: handlers}, nil
}

// Handler builds and returns the HTTP mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Root collection: list + create
	mux.HandleFunc("/api/audits", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.h.listAudits(w, r)
		case http.MethodPost:
			s.h.startAudit(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Sub-resources: /api/audits/diff, /api/audits/{id}[/action]
	mux.HandleFunc("/api/audits/", func(w http.ResponseWriter, r *http.Request) {
		// strip prefix, split on first /
		rest := strings.TrimPrefix(r.URL.Path, "/api/audits/")
		parts := strings.SplitN(rest, "/", 2)
		first := parts[0]
		second := ""
		if len(parts) > 1 {
			second = parts[1]
		}

		if first == "" {
			http.NotFound(w, r)
			return
		}

		switch {
		// GET /api/audits/diff?a=X&b=Y
		case first == "diff" && second == "" && r.Method == http.MethodGet:
			s.h.diffAudits(w, r)

		// GET /api/audits/{id}/events
		case second == "events" && r.Method == http.MethodGet:
			s.h.auditEvents(w, r, first)

		// POST /api/audits/{id}/cancel
		case second == "cancel" && r.Method == http.MethodPost:
			s.h.cancelAudit(w, r, first)

		// GET /api/audits/{id}/report.html
		case second == "report.html" && r.Method == http.MethodGet:
			s.h.serveReportHTML(w, r, first)

		// GET /api/audits/{id}/report.json
		case second == "report.json" && r.Method == http.MethodGet:
			s.h.serveReportJSON(w, r, first)

		// GET /api/audits/{id}
		case second == "" && r.Method == http.MethodGet:
			s.h.getAudit(w, r, first)

		// DELETE /api/audits/{id}
		case second == "" && r.Method == http.MethodDelete:
			s.h.deleteAudit(w, r, first)

		default:
			http.NotFound(w, r)
		}
	})

	return corsMiddleware(mux)
}

// ListenAndServe starts the HTTP server on addr (e.g. ":8080").
func (s *Server) ListenAndServe(addr string) error {
	fmt.Printf("SEO Audit Server  →  http://localhost%s\n", addr)
	fmt.Printf("Reports stored in →  %s\n", s.h.storage.BaseDir())
	return http.ListenAndServe(addr, s.Handler())
}

// corsMiddleware adds CORS headers and handles pre-flight OPTIONS requests.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
