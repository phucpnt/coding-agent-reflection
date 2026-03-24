package reflection

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

type mockLLM struct {
	response string
	err      error
}

func (m *mockLLM) Complete(_ context.Context, _ string) (string, error) {
	return m.response, m.err
}

type mockStore struct {
	interactions []model.Interaction
	reflections  []model.Reflection
	pruned       int
}

func (m *mockStore) QueryByDateRange(_ context.Context, from, to time.Time) ([]model.Interaction, error) {
	var result []model.Interaction
	for _, i := range m.interactions {
		if !i.Ts.Before(from) && i.Ts.Before(to) {
			result = append(result, i)
		}
	}
	return result, nil
}

func (m *mockStore) UpsertReflection(_ context.Context, r model.Reflection) error {
	m.reflections = append(m.reflections, r)
	return nil
}

func (m *mockStore) PruneInteractions(_ context.Context, days int) (int64, error) {
	m.pruned = days
	return 0, nil
}

func TestRunReflection_WithInteractions(t *testing.T) {
	today := time.Now()
	store := &mockStore{
		interactions: []model.Interaction{
			{
				ID: uuid.New(), Ts: today, Provider: "claude",
				UserPrompt: "fix the bug", AgentOutput: "fixed it",
			},
		},
	}

	llm := &mockLLM{
		response: `## Summary
Worked on bug fixes.

## Should Do
Use clear prompts.

## Should Not Do
Avoid vague instructions.

## Config Changes
none`,
	}

	r, err := RunReflection(context.Background(), store, llm, today)
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("expected reflection, got nil")
	}
	if r.Summary != "Worked on bug fixes." {
		t.Errorf("unexpected summary: %q", r.Summary)
	}
	if r.ShouldDo != "Use clear prompts." {
		t.Errorf("unexpected should_do: %q", r.ShouldDo)
	}
	if len(store.reflections) != 1 {
		t.Errorf("expected 1 upserted reflection, got %d", len(store.reflections))
	}
}

func TestRunReflection_NoInteractions(t *testing.T) {
	store := &mockStore{}
	llm := &mockLLM{response: "should not be called"}

	r, err := RunReflection(context.Background(), store, llm, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Fatal("expected nil reflection for no interactions")
	}
}

func TestParseReflectionResponse(t *testing.T) {
	input := `## Summary
Did things.

## Should Do
Keep doing them.

## Should Not Do
Stop doing bad things.

## Config Changes
Update the timeout setting.`

	result := parseReflectionResponse(input)
	if result.Summary != "Did things." {
		t.Errorf("summary: %q", result.Summary)
	}
	if result.ShouldDo != "Keep doing them." {
		t.Errorf("should_do: %q", result.ShouldDo)
	}
	if result.ShouldNotDo != "Stop doing bad things." {
		t.Errorf("should_not_do: %q", result.ShouldNotDo)
	}
	if result.ConfigChanges != "Update the timeout setting." {
		t.Errorf("config_changes: %q", result.ConfigChanges)
	}
}
