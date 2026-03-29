package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cfg "github.com/phuc/coding-agent-reflection/internal/config"
	"github.com/phuc/coding-agent-reflection/internal/setup"
)

var setupCmd = &cobra.Command{
	Use:   "setup [claude|gemini|codex]",
	Short: "Configure hooks for a coding CLI",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetup,
}

var (
	setupGlobal  bool
	setupProject bool
)

func init() {
	setupCmd.Flags().BoolVar(&setupGlobal, "global", false, "Install hooks globally")
	setupCmd.Flags().BoolVar(&setupProject, "project", false, "Install hooks for current project only")
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	c, _ := cfg.Load()
	provider := strings.ToLower(args[0])

	switch provider {
	case "claude":
		scope := resolveScope()
		hookScript := setup.FindHookScript()
		if hookScript == "" {
			return fmt.Errorf("cannot find scripts/claude-hook.sh — run from the project directory or ensure the script exists")
		}
		if err := setup.SetupClaude(scope, hookScript); err != nil {
			return err
		}
		fmt.Printf("✓ Claude Code hooks configured (%s)\n", scope)
		fmt.Println("Restart your Claude Code session for hooks to take effect.")

	case "gemini":
		if err := setup.SetupGemini(c.Port); err != nil {
			return err
		}
		fmt.Println("✓ Gemini CLI hooks configured")
		fmt.Println("Restart your Gemini CLI session for hooks to take effect.")

	case "codex":
		if err := setup.SetupCodex(c.Port); err != nil {
			return err
		}
		fmt.Println("✓ Codex OTel export configured")
		fmt.Println("Restart Codex for changes to take effect.")

	default:
		return fmt.Errorf("unknown provider: %s (use claude, gemini, or codex)", provider)
	}

	return nil
}

func resolveScope() string {
	if setupGlobal {
		return "global"
	}
	if setupProject {
		return "project"
	}

	fmt.Println("Install hooks globally or for this project?")
	fmt.Println("  1) global  — all Claude Code sessions")
	fmt.Println("  2) project — this directory only")
	fmt.Print("Choose [1/2]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "2", "project":
		return "project"
	default:
		return "global"
	}
}
