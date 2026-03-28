package reflection

import (
	"context"
	"log/slog"
	"time"
)

func RunScheduler(ctx context.Context, store Store, llm LLMClient, schedule string, outputDir string, retentionDays int) {
	if schedule == "off" || schedule == "" {
		slog.Info("reflection scheduler disabled")
		return
	}

	slog.Info("reflection scheduler started", "schedule", schedule)

	// Check immediately on startup for missed reflections
	runIfNeeded(ctx, store, llm, outputDir, retentionDays)

	// Then tick every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runIfNeeded(ctx, store, llm, outputDir, retentionDays)
		}
	}
}

func runIfNeeded(ctx context.Context, store Store, llm LLMClient, outputDir string, retentionDays int) {
	yesterday := time.Now().AddDate(0, 0, -1)
	dateStr := yesterday.Format("2006-01-02")

	exists, err := store.HasReflection(ctx, yesterday)
	if err != nil {
		slog.Error("check reflection existence failed", "err", err, "date", dateStr)
		return
	}
	if exists {
		slog.Debug("reflection already exists", "date", dateStr)
		return
	}

	slog.Info("running reflection for missing date", "date", dateStr)
	r, err := RunReflection(ctx, store, llm, yesterday, outputDir)
	if err != nil {
		slog.Error("scheduled reflection failed", "err", err, "date", dateStr)
		return
	}
	if r == nil {
		slog.Info("no interactions for reflection", "date", dateStr)
		return
	}

	slog.Info("reflection complete", "date", dateStr)

	if retentionDays > 0 {
		deleted, err := store.PruneInteractions(ctx, retentionDays)
		if err != nil {
			slog.Error("prune failed", "err", err)
		} else if deleted > 0 {
			slog.Info("pruned old interactions", "deleted", deleted)
		}
	}
}
