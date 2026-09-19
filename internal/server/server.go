package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"jazz-tools/web"
)

// Config holds configuration for the HTTP server.
type Config struct {
	Host    string
	Port    int
	DevMode bool
}

// Server provides the HTTP API and web interface.
type Server struct {
	cfg      Config
	mux      *http.ServeMux
	staticFS fs.FS
}

// NewServer initializes a new jazz-tools web server.
func NewServer(cfg Config) (*Server, error) {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	assets, err := web.Assets()
	if err != nil {
		return nil, fmt.Errorf("failed to load web assets: %w", err)
	}

	s := &Server{
		cfg:      cfg,
		mux:      http.NewServeMux(),
		staticFS: assets,
	}

	s.registerRoutes()
	return s, nil
}

func (s *Server) registerRoutes() {
	// API Endpoints
	s.mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","version":"1.0.0"}`))
	})
	s.mux.HandleFunc("GET /api/samples", s.handleSamples)
	s.mux.HandleFunc("GET /api/standards", s.handleStandards)
	s.mux.HandleFunc("POST /api/analyze", s.handleAnalyze)
	s.mux.HandleFunc("POST /api/companion/xml", s.handleCompanionXML)
	s.mux.HandleFunc("POST /api/companion/pdf", s.handleCompanionPDF)

	// Static SPA file handler
	fileServer := http.FileServerFS(s.staticFS)
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in embedded assets
		if f, err := s.staticFS.Open(path); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA client-side routing
		indexData, err := fs.ReadFile(s.staticFS, "index.html")
		if err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(indexData)
			return
		}

		http.NotFound(w, r)
	})
}

// ServeHTTP implements http.Handler with CORS and logging.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for frontend development
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	fmt.Printf("🎶 jazz-tools web server running at http://localhost:%d\n", s.cfg.Port)
	return http.ListenAndServe(addr, s)
}
