package reflection

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

// PromptTemplatePath can be set to a file path containing a custom prompt template.
// The template should contain {{INTERACTIONS}} as a placeholder.
var PromptTemplatePath string

type Store interface {
	QueryByDateRange(ctx context.Context, from, to time.Time) ([]model.Interaction, error)
	UpsertReflection(ctx context.Context, r model.Reflection) error
	PruneInteractions(ctx context.Context, retentionDays int) (int64, error)
	HasReflection(ctx context.Context, date time.Time) (bool, error)
}

type ReflectionResult struct {
	Summary       string `json:"summary"`
	ShouldDo      string `json:"should_do"`
	ShouldNotDo   string `json:"should_not_do"`
	ConfigChanges string `json:"config_changes"`
}

func RunReflection(ctx context.Context, store Store, llm LLMClient, targetDate time.Time, outputDir string) (*model.Reflection, error) {
	from := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.Local)
	to := from.Add(24 * time.Hour)

	interactions, err := store.QueryByDateRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("query interactions: %w", err)
	}

	if len(interactions) == 0 {
		return nil, nil // no interactions to reflect on
	}

	prompt := buildReflectionPrompt(interactions)
	response, err := llm.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm complete: %w", err)
	}

	result := parseReflectionResponse(response)

	r := model.Reflection{
		ID:            uuid.New(),
		Date:          from,
		Summary:       result.Summary,
		ShouldDo:      result.ShouldDo,
		ShouldNotDo:   result.ShouldNotDo,
		ConfigChanges: result.ConfigChanges,
		CreatedAt:     time.Now(),
	}

	if err := store.UpsertReflection(ctx, r); err != nil {
		return nil, fmt.Errorf("upsert reflection: %w", err)
	}

	if outputDir != "" {
		if _, err := WriteReflectionFile(outputDir, targetDate, r, len(interactions)); err != nil {
			return &r, fmt.Errorf("write reflection file (db saved ok): %w", err)
		}
	}

	return &r, nil
}

func buildReflectionPrompt(interactions []model.Interaction) string {
	interactionsBlock := formatInteractions(interactions)

	// Try custom template first
	if PromptTemplatePath != "" {
		if tmpl, err := os.ReadFile(PromptTemplatePath); err == nil {
			return strings.ReplaceAll(string(tmpl), "{{INTERACTIONS}}", interactionsBlock)
		}
	}

	// Default template
	return defaultPromptTemplate(interactionsBlock)
}

func formatInteractions(interactions []model.Interaction) string {
	var b strings.Builder
	for i, inter := range interactions {
		fmt.Fprintf(&b, "--- Interaction %d (%s) ---\n", i+1, inter.Provider)
		fmt.Fprintf(&b, "Project: %s\n", inter.Project)
		fmt.Fprintf(&b, "Prompt: %s\n", truncate(inter.UserPrompt, 500))
		fmt.Fprintf(&b, "Output: %s\n\n", truncate(inter.AgentOutput, 500))
	}
	return b.String()
}

func defaultPromptTemplate(interactions string) string {
	return `Analyze the following coding agent interactions from today and provide a structured reflection.

For each interaction, I'll show the provider, prompt, and output.

` + interactions + `Please respond with exactly these four sections, using these exact headers:

## Summary
A brief summary of the day's interactions — what was accomplished, overall patterns.

## Should Do
Patterns, techniques, or prompt styles that worked well and should be repeated.

## Should Not Do
Mistakes, anti-patterns, or ineffective approaches to avoid.

## Config Changes
Suggestions for updating agent configs, rules, or workflows based on today's patterns. Write "none" if no changes suggested.
`
}

func parseReflectionResponse(response string) ReflectionResult {
	sections := map[string]*string{
		"## Summary":        nil,
		"## Should Do":      nil,
		"## Should Not Do":  nil,
		"## Config Changes": nil,
	}
	result := ReflectionResult{}
	sections["## Summary"] = &result.Summary
	sections["## Should Do"] = &result.ShouldDo
	sections["## Should Not Do"] = &result.ShouldNotDo
	sections["## Config Changes"] = &result.ConfigChanges

	headers := []string{"## Summary", "## Should Do", "## Should Not Do", "## Config Changes"}

	for i, header := range headers {
		idx := strings.Index(response, header)
		if idx == -1 {
			continue
		}
		start := idx + len(header)
		end := len(response)
		if i+1 < len(headers) {
			nextIdx := strings.Index(response, headers[i+1])
			if nextIdx != -1 {
				end = nextIdx
			}
		}
		*sections[header] = strings.TrimSpace(response[start:end])
	}

	// Fallback: if no sections parsed, put everything in summary
	if result.Summary == "" && result.ShouldDo == "" {
		result.Summary = response
	}

	return result
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
