package reflection

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

func TestRunIfNeeded_NoExistingReflection(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	store := &mockStore{
		interactions: []model.Interaction{
			{ID: uuid.New(), Ts: yesterday, Provider: "claude", UserPrompt: "test", AgentOutput: "done"},
		},
	}
	llm := &mockLLM{response: "## Summary\ntest\n\n## Should Do\ntest\n\n## Should Not Do\ntest\n\n## Config Changes\nnone"}

	runIfNeeded(context.Background(), store, llm, t.TempDir(), 0)

	if len(store.reflections) != 1 {
		t.Fatalf("expected 1 reflection, got %d", len(store.reflections))
	}
}

func TestRunIfNeeded_AlreadyExists(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	store := &mockStore{
		reflections: []model.Reflection{
			{ID: uuid.New(), Date: yesterday, Summary: "already done"},
		},
	}
	llm := &mockLLM{response: "should not be called"}

	runIfNeeded(context.Background(), store, llm, t.TempDir(), 0)

	// Should still be 1 — no new reflection created
	if len(store.reflections) != 1 {
		t.Fatalf("expected 1 reflection (unchanged), got %d", len(store.reflections))
	}
}
