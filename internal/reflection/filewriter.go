package reflection

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/phuc/coding-agent-reflection/internal/model"
)

func WriteReflectionFile(dir string, targetDate time.Time, r model.Reflection, interactionCount int) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create reflection dir: %w", err)
	}

	datePrefix := targetDate.Format("20060102")
	seq := nextSequence(dir, datePrefix)
	filename := fmt.Sprintf("%s-%03d.md", datePrefix, seq)
	path := filepath.Join(dir, filename)

	content := formatReflectionFile(targetDate, r, interactionCount)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write reflection file: %w", err)
	}

	return path, nil
}

func nextSequence(dir, datePrefix string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 1
	}

	var seqs []int
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, datePrefix+"-") || !strings.HasSuffix(name, ".md") {
			continue
		}
		// Extract NNN from YYYYMMDD-NNN.md
		numStr := strings.TrimSuffix(strings.TrimPrefix(name, datePrefix+"-"), ".md")
		if n, err := strconv.Atoi(numStr); err == nil {
			seqs = append(seqs, n)
		}
	}

	if len(seqs) == 0 {
		return 1
	}
	sort.Ints(seqs)
	return seqs[len(seqs)-1] + 1
}

func formatReflectionFile(targetDate time.Time, r model.Reflection, interactionCount int) string {
	var b strings.Builder

	b.WriteString("---\n")
	fmt.Fprintf(&b, "date: %s\n", targetDate.Format("2006-01-02"))
	fmt.Fprintf(&b, "interactions: %d\n", interactionCount)
	fmt.Fprintf(&b, "created: %s\n", r.CreatedAt.Format(time.RFC3339))
	b.WriteString("---\n\n")

	fmt.Fprintf(&b, "# Reflection — %s\n\n", targetDate.Format("2006-01-02"))

	b.WriteString("## Summary\n\n")
	b.WriteString(r.Summary)
	b.WriteString("\n\n")

	b.WriteString("## Should Do\n\n")
	b.WriteString(r.ShouldDo)
	b.WriteString("\n\n")

	b.WriteString("## Should Not Do\n\n")
	b.WriteString(r.ShouldNotDo)
	b.WriteString("\n\n")

	b.WriteString("## Config Changes\n\n")
	b.WriteString(r.ConfigChanges)
	b.WriteString("\n")

	return b.String()
}
