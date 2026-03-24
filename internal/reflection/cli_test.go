package reflection

import (
	"context"
	"testing"
	"time"
)

func TestCLICompleter_Echo(t *testing.T) {
	c := &CLICompleter{Command: "cat", Timeout: 10 * time.Second}
	result, err := c.Complete(context.Background(), "hello world")
	if err != nil {
		t.Fatal(err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestCLICompleter_EmptyCommand(t *testing.T) {
	c := &CLICompleter{Command: "", Timeout: 10 * time.Second}
	_, err := c.Complete(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestCLICompleter_BadCommand(t *testing.T) {
	c := &CLICompleter{Command: "nonexistent-binary-xyz", Timeout: 10 * time.Second}
	_, err := c.Complete(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}

func TestCLICompleter_Timeout(t *testing.T) {
	c := &CLICompleter{Command: "sleep 60", Timeout: 100 * time.Millisecond}
	_, err := c.Complete(context.Background(), "")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
