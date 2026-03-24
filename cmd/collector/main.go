package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phuc/coding-agent-reflection/config"
	"github.com/phuc/coding-agent-reflection/internal/ingest"
	"github.com/phuc/coding-agent-reflection/internal/reflection"
	"github.com/phuc/coding-agent-reflection/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := config.Load()

	s, err := store.New(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open store", "err", err)
		os.Exit(1)
	}
	defer s.Close()

	llm := &reflection.ClaudeClient{
		APIKey: cfg.LLMAPIKey,
		Model:  cfg.LLMModel,
	}

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := s.HealthCheck(r.Context()); err != nil {
			slog.Error("health check failed", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Ingest
	mux.HandleFunc("POST /ingest/claude", ingest.ClaudeHandler(s))
	mux.HandleFunc("POST /ingest/gemini", ingest.GeminiHandler(s))
	mux.HandleFunc("POST /v1/traces", ingest.CodexHandler(s))

	// Query
	mux.HandleFunc("GET /interactions", func(w http.ResponseWriter, r *http.Request) {
		from := time.Now().AddDate(0, 0, -7)
		to := time.Now().Add(time.Hour)
		interactions, err := s.QueryByDateRange(r.Context(), from, to)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(interactions)
	})

	// Reflection
	mux.HandleFunc("POST /jobs/daily-reflection", reflection.Handler(s, llm, cfg.RetentionDays))

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("starting collector", "port", cfg.Port, "db", cfg.DBPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
