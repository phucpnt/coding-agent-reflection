package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/phuc/coding-agent-reflection/internal/ingest"
	"github.com/phuc/coding-agent-reflection/internal/reflection"
	"github.com/phuc/coding-agent-reflection/internal/store"
)

func setupTestServer(t *testing.T) (*store.Store, *http.ServeMux) {
	t.Helper()
	dir := t.TempDir()
	s, err := store.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest/claude", ingest.ClaudeHandler(s))
	mux.HandleFunc("POST /ingest/gemini", ingest.GeminiHandler(s))
	mux.HandleFunc("POST /v1/traces", ingest.CodexHandler(s))

	mockLLM := &testLLM{response: "## Summary\nTest reflection.\n\n## Should Do\nKeep testing.\n\n## Should Not Do\nSkip tests.\n\n## Config Changes\nnone"}
	reflDir := filepath.Join(dir, "reflections")
	mux.HandleFunc("POST /jobs/daily-reflection", reflection.Handler(s, mockLLM, 0, reflDir))

	return s, mux
}

type testLLM struct {
	response string
}

func (m *testLLM) Complete(_ context.Context, _ string) (string, error) {
	return m.response, nil
}

func TestE2E_ClaudeIngest(t *testing.T) {
	s, mux := setupTestServer(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body := `{"session_id":"e2e-claude","cwd":"/proj","user_prompt":"test prompt","agent_output":"test output"}`
	resp, err := http.Post(srv.URL+"/ingest/claude", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Verify in DB
	rows, err := s.QueryByDateRange(context.Background(),
		timeZero(), timeFuture())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Provider != "claude" {
		t.Fatalf("expected 1 claude row, got %d", len(rows))
	}
}

func TestE2E_GeminiIngest(t *testing.T) {
	s, mux := setupTestServer(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body := `{"session_id":"e2e-gemini","cwd":"/proj","user_prompt":"gemini prompt","agent_output":"gemini output"}`
	resp, err := http.Post(srv.URL+"/ingest/gemini", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	rows, err := s.QueryByDateRange(context.Background(), timeZero(), timeFuture())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Provider != "gemini" {
		t.Fatalf("expected 1 gemini row, got %d", len(rows))
	}
}

func TestE2E_CodexIngest(t *testing.T) {
	s, mux := setupTestServer(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	payload := map[string]any{
		"resourceSpans": []map[string]any{{
			"scopeSpans": []map[string]any{{
				"spans": []map[string]any{{
					"name": "llm.call",
					"attributes": []map[string]any{
						{"key": "user_prompt", "value": map[string]string{"stringValue": "codex prompt"}},
						{"key": "completion", "value": map[string]string{"stringValue": "codex output"}},
					},
				}},
			}},
		}},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(srv.URL+"/v1/traces", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	rows, err := s.QueryByDateRange(context.Background(), timeZero(), timeFuture())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Provider != "codex" {
		t.Fatalf("expected 1 codex row, got %d", len(rows))
	}
}

func TestE2E_ReflectionAfterIngest(t *testing.T) {
	_, mux := setupTestServer(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Insert an interaction first
	body := `{"session_id":"refl-test","cwd":"/proj","user_prompt":"do thing","agent_output":"did thing"}`
	resp, err := http.Post(srv.URL+"/ingest/claude", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("ingest: expected 200, got %d", resp.StatusCode)
	}

	// Trigger reflection for today (since we just inserted)
	today := timeNowDateStr()
	reflBody := `{"date":"` + today + `"}`
	resp, err = http.Post(srv.URL+"/jobs/daily-reflection", "application/json", bytes.NewBufferString(reflBody))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("reflection: expected 200, got %d: %v", resp.StatusCode, errResp)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	if result["summary"] != "Test reflection." {
		t.Errorf("unexpected summary: %v", result["summary"])
	}
}
