package reflection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

// PromptTemplatePath can be set to a file path containing a custom prompt template.
// The template should contain {{INTERACTIONS_FILE}} as a placeholder for the file path.
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

	// Write interactions to a temp file so the CLI agent can read it
	interactionsFile, err := writeInteractionsFile(interactions, targetDate)
	if err != nil {
		return nil, fmt.Errorf("write interactions file: %w", err)
	}
	defer os.Remove(interactionsFile)

	prompt := buildReflectionPrompt(interactionsFile, len(interactions))
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

// InteractionsDir is where temp interaction files are written.
// Defaults to ./data but can be overridden (must be readable by the CLI agent).
var InteractionsDir = "./data"

func writeInteractionsFile(interactions []model.Interaction, targetDate time.Time) (string, error) {
	dir := InteractionsDir
	os.MkdirAll(dir, 0o755)
	filename := fmt.Sprintf("interactions-%s.md", targetDate.Format("20060102"))
	path := filepath.Join(dir, filename)

	var b strings.Builder
	fmt.Fprintf(&b, "# Coding Agent Interactions — %s\n\n", targetDate.Format("2006-01-02"))
	fmt.Fprintf(&b, "Total interactions: %d\n\n", len(interactions))

	for i, inter := range interactions {
		fmt.Fprintf(&b, "## Interaction %d\n\n", i+1)
		fmt.Fprintf(&b, "- **Provider**: %s\n", inter.Provider)
		fmt.Fprintf(&b, "- **Time**: %s\n", inter.Ts.Format("15:04:05"))
		fmt.Fprintf(&b, "- **Session**: %s\n", inter.SessionID)
		if inter.Project != "" {
			fmt.Fprintf(&b, "- **Project**: %s\n", inter.Project)
		}
		if inter.ToolsUsed.Valid && inter.ToolsUsed.String != "" {
			fmt.Fprintf(&b, "- **Tools**: %s\n", inter.ToolsUsed.String)
		}
		b.WriteString("\n### User Prompt\n\n")
		b.WriteString(inter.UserPrompt)
		b.WriteString("\n\n### Agent Output\n\n")
		b.WriteString(inter.AgentOutput)
		b.WriteString("\n\n---\n\n")
	}

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	// Return absolute path so CLI agents can find it regardless of cwd
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return absPath, nil
}

func buildReflectionPrompt(interactionsFile string, count int) string {
	// Try custom template first
	if PromptTemplatePath != "" {
		if tmpl, err := os.ReadFile(PromptTemplatePath); err == nil {
			return strings.ReplaceAll(string(tmpl), "{{INTERACTIONS_FILE}}", interactionsFile)
		}
	}

	return defaultPromptTemplate(interactionsFile, count)
}

func defaultPromptTemplate(interactionsFile string, count int) string {
	return fmt.Sprintf(`Read the file at %s which contains %d coding agent interactions from today.

Analyze the interactions and provide a structured reflection. Please respond with exactly these four sections, using these exact headers:

## Summary
A brief summary of the day's interactions — what was accomplished, overall patterns.

## Should Do
Patterns, techniques, or prompt styles that worked well and should be repeated.

## Should Not Do
Mistakes, anti-patterns, or ineffective approaches to avoid.

## Config Changes
Suggestions for updating agent configs, rules, or workflows based on today's patterns. Write "none" if no changes suggested.
`, interactionsFile, count)
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
