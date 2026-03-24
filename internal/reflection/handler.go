package reflection

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type handlerConfig struct {
	Store         Store
	LLM           LLMClient
	RetentionDays int
}

func Handler(store Store, llm LLMClient, retentionDays int) http.HandlerFunc {
	cfg := handlerConfig{Store: store, LLM: llm, RetentionDays: retentionDays}
	return cfg.serveHTTP
}

func (c *handlerConfig) serveHTTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Date string `json:"date"`
	}
	json.NewDecoder(r.Body).Decode(&body) // ignore error — body is optional

	targetDate := time.Now().AddDate(0, 0, -1) // yesterday
	if body.Date != "" {
		parsed, err := time.Parse("2006-01-02", body.Date)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		targetDate = parsed
	}

	reflection, err := RunReflection(r.Context(), c.Store, c.LLM, targetDate)
	if err != nil {
		slog.Error("reflection job failed", "err", err, "date", targetDate.Format("2006-01-02"))
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "reflection failed: " + err.Error()})
		return
	}

	if reflection == nil {
		slog.Info("no interactions for reflection", "date", targetDate.Format("2006-01-02"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "no interactions to reflect on for " + targetDate.Format("2006-01-02"),
		})
		return
	}

	// Prune if retention configured
	if c.RetentionDays > 0 {
		deleted, err := c.Store.PruneInteractions(r.Context(), c.RetentionDays)
		if err != nil {
			slog.Error("prune failed", "err", err)
		} else if deleted > 0 {
			slog.Info("pruned old interactions", "deleted", deleted)
		}
	}

	slog.Info("reflection complete", "date", targetDate.Format("2006-01-02"))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"date":    targetDate.Format("2006-01-02"),
		"summary": reflection.Summary,
	})
}
