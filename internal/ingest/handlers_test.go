package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phuc/coding-agent-reflection/internal/model"
)

type mockInserter struct {
	interactions []model.Interaction
}

func (m *mockInserter) Insert(_ context.Context, i model.Interaction) error {
	m.interactions = append(m.interactions, i)
	return nil
}

func TestClaudeHandler_Valid(t *testing.T) {
	m := &mockInserter{}
	h := ClaudeHandler(m)

	body := `{"session_id":"s1","cwd":"/home/user/proj","user_prompt":"fix bug","agent_output":"done","tools_used":["edit"]}`
	req := httptest.NewRequest(http.MethodPost, "/ingest/claude", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(m.interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(m.interactions))
	}
	if m.interactions[0].Provider != "claude" {
		t.Errorf("expected provider claude, got %s", m.interactions[0].Provider)
	}
}

func TestClaudeHandler_MissingSessionID(t *testing.T) {
	m := &mockInserter{}
	h := ClaudeHandler(m)

	body := `{"cwd":"/home/user/proj"}`
	req := httptest.NewRequest(http.MethodPost, "/ingest/claude", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestClaudeHandler_MalformedJSON(t *testing.T) {
	m := &mockInserter{}
	h := ClaudeHandler(m)

	req := httptest.NewRequest(http.MethodPost, "/ingest/claude", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGeminiHandler_Valid(t *testing.T) {
	m := &mockInserter{}
	h := GeminiHandler(m)

	body := `{"session_id":"g1","cwd":"/proj","user_prompt":"add feature","agent_output":"added"}`
	req := httptest.NewRequest(http.MethodPost, "/ingest/gemini", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if m.interactions[0].Provider != "gemini" {
		t.Errorf("expected provider gemini, got %s", m.interactions[0].Provider)
	}
}

func TestCodexHandler_ValidSpan(t *testing.T) {
	m := &mockInserter{}
	h := CodexHandler(m)

	payload := otlpTraceRequest{
		ResourceSpans: []resourceSpan{{
			ScopeSpans: []scopeSpan{{
				Spans: []span{{
					Name: "llm.call",
					Attributes: []attribute{
						{Key: "user_prompt", Value: attributeValue{StringValue: "write tests"}},
						{Key: "completion", Value: attributeValue{StringValue: "here are tests"}},
						{Key: "code.filepath", Value: attributeValue{StringValue: "/home/user/repo"}},
					},
				}},
			}},
		}},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/traces", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(m.interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(m.interactions))
	}
	if m.interactions[0].Provider != "codex" {
		t.Errorf("expected provider codex, got %s", m.interactions[0].Provider)
	}
}

func TestCodexHandler_NoRelevantSpans(t *testing.T) {
	m := &mockInserter{}
	h := CodexHandler(m)

	payload := otlpTraceRequest{
		ResourceSpans: []resourceSpan{{
			ScopeSpans: []scopeSpan{{
				Spans: []span{{
					Name:       "http.request",
					Attributes: []attribute{{Key: "url", Value: attributeValue{StringValue: "https://api.example.com"}}},
				}},
			}},
		}},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/traces", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(m.interactions) != 0 {
		t.Fatalf("expected 0 interactions, got %d", len(m.interactions))
	}
}

func TestCodexHandler_MalformedJSON(t *testing.T) {
	m := &mockInserter{}
	h := CodexHandler(m)

	req := httptest.NewRequest(http.MethodPost, "/v1/traces", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
