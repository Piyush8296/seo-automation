package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Server wires together storage, the hub, the manager, and the HTTP mux.
type Server struct {
	h     *Handlers
	addr  string
	uiDir string // path to the built frontend (ui/dist)
}

// New creates a Server ready to serve.
// baseDir is where audit reports are stored; defaults to ~/.seo-reports when empty.
// uiDir is the path to the built frontend assets (e.g. "ui/dist"); leave empty to disable UI serving.
func New(baseDir, uiDir string) (*Server, error) {
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

	return &Server{h: handlers, uiDir: uiDir}, nil
}

// Handler builds and returns the HTTP mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Settings: GET to read, PUT to write
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.h.getSettings(w, r)
		case http.MethodPut:
			s.h.updateSettings(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Check catalog (read-only introspection)
	mux.HandleFunc("/api/checks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.listChecks(w, r)
	})

	// External check catalog (read-only introspection for integration-backed checks)
	mux.HandleFunc("/api/external-checks/catalog", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.listExternalCheckCatalog(w, r)
	})

	// Local SEO / Google Business Profile workspace.
	mux.HandleFunc("/api/local-seo/gbp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.getLocalSEO(w, r)
	})

	// Placeholder write endpoint for future GBP update/post operations.
	mux.HandleFunc("/api/local-seo/gbp/actions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.submitGBPAction(w, r)
	})

	// Search integrations workspace for GSC and Bing Webmaster.
	mux.HandleFunc("/api/search-integrations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.getSearchIntegrations(w, r)
	})

	// Placeholder OAuth connect endpoint for GSC and Bing.
	mux.HandleFunc("/api/search-integrations/oauth/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.connectSearchOAuth(w, r)
	})

	// Placeholder POST operations for GSC URL inspection and sitemap submission.
	mux.HandleFunc("/api/search-integrations/actions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.submitSearchIntegrationAction(w, r)
	})

	// Crawler evidence workspace and bounded crawler-backed report.
	mux.HandleFunc("/api/crawler-evidence", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.getCrawlerEvidence(w, r)
	})

	mux.HandleFunc("/api/crawler-evidence/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.h.runCrawlerEvidence(w, r)
	})

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

	// Serve the React UI for all non-API routes (SPA fallback).
	if s.uiDir != "" {
		uiFS := http.Dir(s.uiDir)
		fileServer := http.FileServer(uiFS)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the exact file (JS, CSS, images, etc.)
			p := strings.TrimPrefix(r.URL.Path, "/")
			if p == "" {
				p = "index.html"
			}
			if _, err := fs.Stat(os.DirFS(s.uiDir), p); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
			// Fall back to index.html for client-side routing
			http.ServeFile(w, r, filepath.Join(s.uiDir, "index.html"))
		})
	}

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
