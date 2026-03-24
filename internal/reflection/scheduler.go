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

	for {
		var wait time.Duration
		switch schedule {
		case "daily":
			wait = untilNextHour(2) // 02:00 local time
		case "hourly":
			wait = untilNextHour(-1) // next full hour
		default:
			slog.Error("unknown reflection schedule", "schedule", schedule)
			return
		}

		slog.Info("next reflection scheduled", "in", wait.Round(time.Minute))

		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		targetDate := time.Now()
		if schedule == "daily" {
			targetDate = targetDate.AddDate(0, 0, -1) // reflect on yesterday
		}

		slog.Info("running scheduled reflection", "date", targetDate.Format("2006-01-02"))
		r, err := RunReflection(ctx, store, llm, targetDate, outputDir)
		if err != nil {
			slog.Error("scheduled reflection failed", "err", err)
			continue
		}
		if r == nil {
			slog.Info("no interactions for scheduled reflection", "date", targetDate.Format("2006-01-02"))
			continue
		}

		slog.Info("scheduled reflection complete", "date", targetDate.Format("2006-01-02"))

		if retentionDays > 0 {
			deleted, err := store.PruneInteractions(ctx, retentionDays)
			if err != nil {
				slog.Error("scheduled prune failed", "err", err)
			} else if deleted > 0 {
				slog.Info("pruned old interactions", "deleted", deleted)
			}
		}
	}
}

func untilNextHour(targetHour int) time.Duration {
	now := time.Now()
	if targetHour < 0 {
		// Next full hour
		next := now.Truncate(time.Hour).Add(time.Hour)
		return next.Sub(now)
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), targetHour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next.Sub(now)
}
