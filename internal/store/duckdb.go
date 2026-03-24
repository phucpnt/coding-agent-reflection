package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

type Store struct {
	db *sql.DB
	mu sync.Mutex
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
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
			id           UUID,
			ts           TIMESTAMP,
			provider     TEXT,
			session_id   TEXT,
			project      TEXT,
			user_prompt  TEXT,
			agent_output TEXT,
			context      TEXT,
			tokens_prompt BIGINT,
			tokens_output BIGINT,
			tools_used   TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS reflections (
			id             UUID,
			date           DATE UNIQUE,
			summary        TEXT,
			should_do      TEXT,
			should_not_do  TEXT,
			config_changes TEXT,
			created_at     TIMESTAMP
		)`,
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
		i.ID.String(), i.Ts, i.Provider, i.SessionID, i.Project,
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
		 ORDER BY ts ASC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.Interaction
	for rows.Next() {
		var i model.Interaction
		var idStr string
		if err := rows.Scan(&idStr, &i.Ts, &i.Provider, &i.SessionID, &i.Project,
			&i.UserPrompt, &i.AgentOutput, &i.Context,
			&i.TokensPrompt, &i.TokensOutput, &i.ToolsUsed); err != nil {
			return nil, err
		}
		i.ID, _ = uuid.Parse(idStr)
		results = append(results, i)
	}
	return results, rows.Err()
}

func (s *Store) UpsertReflection(ctx context.Context, r model.Reflection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		`DELETE FROM reflections WHERE date = ?`, r.Date)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO reflections (id, date, summary, should_do, should_not_do, config_changes, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.ID.String(), r.Date, r.Summary, r.ShouldDo, r.ShouldNotDo, r.ConfigChanges, r.CreatedAt,
	)
	return err
}

func (s *Store) PruneInteractions(ctx context.Context, retentionDays int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM interactions WHERE ts < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Store) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
