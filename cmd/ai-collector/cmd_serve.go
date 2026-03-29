package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	cfg "github.com/phuc/coding-agent-reflection/internal/config"
	"github.com/phuc/coding-agent-reflection/internal/ingest"
	"github.com/phuc/coding-agent-reflection/internal/reflection"
	"github.com/phuc/coding-agent-reflection/internal/store"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the collector HTTP server in the foreground",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	c, err := cfg.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Auto-create data directories
	os.MkdirAll(cfg.DataDir(), 0o755)
	os.MkdirAll(c.Reflection.OutputDir, 0o755)

	s, err := store.New(c.DBPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	llm := reflection.NewCLICompleter(c.Reflection.CLI)
	if c.Reflection.Prompt != "" {
		reflection.PromptTemplatePath = c.Reflection.Prompt
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := s.HealthCheck(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /ingest/claude", ingest.ClaudeHandler(s))
	mux.HandleFunc("POST /ingest/gemini", ingest.GeminiHandler(s))
	mux.HandleFunc("POST /v1/traces", ingest.CodexHandler(s))

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

	mux.HandleFunc("POST /jobs/daily-reflection", reflection.Handler(s, llm, c.RetentionDays, c.Reflection.OutputDir))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go reflection.RunScheduler(ctx, s, llm, c.Reflection.Schedule, c.Reflection.OutputDir, c.RetentionDays)

	go func() {
		slog.Info("starting collector", "port", c.Port, "db", c.DBPath, "reflection_cli", c.Reflection.CLI, "schedule", c.Reflection.Schedule)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
