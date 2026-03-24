package reflection

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type CLICompleter struct {
	Command string // e.g. "claude --print" or "codex --quiet"
	Timeout time.Duration
}

func NewCLICompleter(command string) *CLICompleter {
	return &CLICompleter{
		Command: command,
		Timeout: 5 * time.Minute,
	}
}

func (c *CLICompleter) Complete(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	parts := strings.Fields(c.Command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty reflection CLI command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("CLI subprocess timed out after %v", c.Timeout)
		}
		return "", fmt.Errorf("CLI subprocess failed: %w (stderr: %s)", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return "", fmt.Errorf("CLI subprocess returned empty output")
	}
	return output, nil
}
