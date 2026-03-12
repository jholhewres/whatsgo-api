package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jholhewres/whatsgo-api/internal/api"
	"github.com/jholhewres/whatsgo-api/internal/auth"
	"github.com/jholhewres/whatsgo-api/internal/config"
	"github.com/jholhewres/whatsgo-api/internal/store"
	"github.com/jholhewres/whatsgo-api/internal/webui"
	"github.com/jholhewres/whatsgo-api/internal/webhook"
	"github.com/jholhewres/whatsgo-api/internal/whatsapp"
)

type Server struct {
	httpServer *http.Server
	manager    *whatsapp.Manager
	dispatcher *webhook.Dispatcher
	authMw     *auth.Middleware
	logger     *slog.Logger
	cfg        *config.Config
	stopClean  context.CancelFunc
}

func New(cfg *config.Config, db store.Store, logger *slog.Logger) (*Server, error) {
	mux := http.NewServeMux()

	authMw := auth.NewMiddleware(db, cfg.Auth.GlobalAPIKey, logger)

	// Create webhook dispatcher
	dispatcher := webhook.NewDispatcher(db, cfg.Webhook, logger)

	// Create WhatsApp instance manager
	manager, err := whatsapp.NewManager(db, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("creating whatsapp manager: %w", err)
	}

	// Register API handlers
	handlers := api.NewHandlers(db, manager, authMw, cfg, logger)
	handlers.RegisterRoutes(mux)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// 404 for unknown API routes (returns JSON instead of HTML)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	// Serve embedded frontend (SPA fallback)
	mux.Handle("/", webui.Handler())

	// Middleware chain
	var handler http.Handler = mux
	handler = securityHeadersMiddleware(handler)
	handler = corsMiddleware(cfg.Server.BaseURL)(handler)
	handler = recoveryMiddleware(logger, handler)
	handler = loggingMiddleware(logger, handler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		httpServer: srv,
		manager:    manager,
		dispatcher: dispatcher,
		authMw:     authMw,
		logger:     logger,
		cfg:        cfg,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	// Load existing instances and reconnect
	go s.manager.LoadAll(ctx)

	// Start webhook dispatcher
	go s.dispatcher.Start(ctx, s.manager.Events())

	// Start cache cleanup with cancellable context
	cleanCtx, cleanCancel := context.WithCancel(ctx)
	s.stopClean = cleanCancel
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-cleanCtx.Done():
				return
			case <-ticker.C:
				s.authMw.CleanupCache()
			}
		}
	}()

	s.logger.Info("server listening", "addr", s.httpServer.Addr, "base_url", s.cfg.Server.BaseURL)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if s.stopClean != nil {
		s.stopClean()
	}

	s.manager.DisconnectAll()

	return s.httpServer.Shutdown(shutdownCtx)
}
