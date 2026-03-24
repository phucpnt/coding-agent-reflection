package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestInsertAndQuery(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Microsecond)
	i := model.Interaction{
		ID:        uuid.New(),
		Ts:        now,
		Provider:  "claude",
		SessionID: "sess-1",
		Project:   "/home/user/project",
		UserPrompt: "fix the bug",
		AgentOutput: "done",
		Context:   `{"files":["main.go"]}`,
		TokensPrompt: sql.NullInt64{Int64: 100, Valid: true},
		TokensOutput: sql.NullInt64{Int64: 200, Valid: true},
		ToolsUsed:    sql.NullString{String: `["edit"]`, Valid: true},
	}

	if err := s.Insert(ctx, i); err != nil {
		t.Fatal(err)
	}

	from := now.Add(-time.Hour)
	to := now.Add(time.Hour)
	results, err := s.QueryByDateRange(ctx, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Provider != "claude" {
		t.Errorf("expected provider claude, got %s", results[0].Provider)
	}
	if results[0].UserPrompt != "fix the bug" {
		t.Errorf("unexpected user_prompt: %s", results[0].UserPrompt)
	}
}

func TestQueryEmptyRange(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	from := time.Now().Add(-2 * time.Hour)
	to := time.Now().Add(-time.Hour)
	results, err := s.QueryByDateRange(ctx, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestUpsertReflection(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	date := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	r := model.Reflection{
		ID:            uuid.New(),
		Date:          date,
		Summary:       "first summary",
		ShouldDo:      "do this",
		ShouldNotDo:   "not that",
		ConfigChanges: "none",
		CreatedAt:     time.Now(),
	}

	if err := s.UpsertReflection(ctx, r); err != nil {
		t.Fatal(err)
	}

	// Upsert again — should overwrite
	r2 := r
	r2.ID = uuid.New()
	r2.Summary = "updated summary"
	if err := s.UpsertReflection(ctx, r2); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := s.db.QueryRow("SELECT count(*) FROM reflections WHERE date = ?", date.Format("2006-01-02")).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 reflection, got %d", count)
	}
}

func TestPruneInteractions(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	old := model.Interaction{
		ID: uuid.New(), Ts: time.Now().AddDate(0, 0, -60),
		Provider: "claude", SessionID: "old",
	}
	recent := model.Interaction{
		ID: uuid.New(), Ts: time.Now(),
		Provider: "claude", SessionID: "recent",
	}

	s.Insert(ctx, old)
	s.Insert(ctx, recent)

	deleted, err := s.PruneInteractions(ctx, 30)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted, got %d", deleted)
	}
}

func TestHealthCheck(t *testing.T) {
	s := testStore(t)
	if err := s.HealthCheck(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestHealthCheckBadDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	// Remove the file to simulate inaccessibility
	os.RemoveAll(dir)

	if err := s.HealthCheck(context.Background()); err == nil {
		t.Fatal("expected health check to fail on closed db")
	}
}
