package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

type Store struct {
	db *sql.DB
	mu sync.Mutex
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS interactions (
			id           TEXT,
			ts           TEXT,
			provider     TEXT,
			session_id   TEXT,
			project      TEXT,
			user_prompt  TEXT,
			agent_output TEXT,
			context      TEXT,
			tokens_prompt INTEGER,
			tokens_output INTEGER,
			tools_used   TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS reflections (
			id             TEXT,
			date           TEXT UNIQUE,
			summary        TEXT,
			should_do      TEXT,
			should_not_do  TEXT,
			config_changes TEXT,
			created_at     TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_interactions_ts ON interactions(ts)`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", q[:40], err)
		}
	}
	return nil
}

func (s *Store) Insert(ctx context.Context, i model.Interaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO interactions (id, ts, provider, session_id, project, user_prompt, agent_output, context, tokens_prompt, tokens_output, tools_used)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		i.ID.String(), i.Ts.Format(time.RFC3339Nano), i.Provider, i.SessionID, i.Project,
		i.UserPrompt, i.AgentOutput, i.Context,
		i.TokensPrompt, i.TokensOutput, i.ToolsUsed,
	)
	return err
}

func (s *Store) QueryByDateRange(ctx context.Context, from, to time.Time) ([]model.Interaction, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, ts, provider, session_id, project, user_prompt, agent_output, context, tokens_prompt, tokens_output, tools_used
		 FROM interactions
		 WHERE ts >= ? AND ts < ?
		 ORDER BY ts ASC`, from.Format(time.RFC3339Nano), to.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.Interaction
	for rows.Next() {
		var i model.Interaction
		var idStr, tsStr string
		if err := rows.Scan(&idStr, &tsStr, &i.Provider, &i.SessionID, &i.Project,
			&i.UserPrompt, &i.AgentOutput, &i.Context,
			&i.TokensPrompt, &i.TokensOutput, &i.ToolsUsed); err != nil {
			return nil, err
		}
		i.ID, _ = uuid.Parse(idStr)
		i.Ts, _ = time.Parse(time.RFC3339Nano, tsStr)
		results = append(results, i)
	}
	return results, rows.Err()
}

func (s *Store) UpsertReflection(ctx context.Context, r model.Reflection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO reflections (id, date, summary, should_do, should_not_do, config_changes, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(date) DO UPDATE SET
		   id=excluded.id, summary=excluded.summary, should_do=excluded.should_do,
		   should_not_do=excluded.should_not_do, config_changes=excluded.config_changes,
		   created_at=excluded.created_at`,
		r.ID.String(), r.Date.Format("2006-01-02"), r.Summary, r.ShouldDo, r.ShouldNotDo, r.ConfigChanges, r.CreatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) PruneInteractions(ctx context.Context, retentionDays int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM interactions WHERE ts < ?`, cutoff.Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Store) HasReflection(ctx context.Context, date time.Time) (bool, error) {
	dateStr := date.Format("2006-01-02")
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM reflections WHERE date = ?`, dateStr).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
