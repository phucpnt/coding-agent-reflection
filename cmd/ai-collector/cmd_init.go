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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive first-run setup wizard",
	RunE:  runInit,
}

var initDefaults bool

func init() {
	initCmd.Flags().BoolVar(&initDefaults, "defaults", false, "Use all default values without prompting")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Welcome to AI Interaction Collector!")
	fmt.Println()

	// Load existing config or defaults
	c, _ := cfg.Load()
	reader := bufio.NewReader(os.Stdin)

	if !initDefaults {
		// Reflection CLI
		fmt.Println("Which CLI do you use for reflections?")
		fmt.Println("  1) claude --print")
		fmt.Println("  2) codex --quiet")
		fmt.Println("  3) gemini")
		fmt.Printf("Choose [1/2/3] (current: %s): ", c.Reflection.CLI)
		if input := readLine(reader); input != "" {
			switch input {
			case "1":
				c.Reflection.CLI = "claude --print"
			case "2":
				c.Reflection.CLI = "codex --quiet"
			case "3":
				c.Reflection.CLI = "gemini"
			default:
				c.Reflection.CLI = input
			}
		}

		// Reflection output directory
		fmt.Printf("\nWhere should reflections be saved?\n  [%s]: ", c.Reflection.OutputDir)
		if input := readLine(reader); input != "" {
			c.Reflection.OutputDir = input
		}

		// Port
		fmt.Printf("\nCollector port?\n  [%d]: ", c.Port)
		if input := readLine(reader); input != "" {
			fmt.Sscanf(input, "%d", &c.Port)
		}

		// Reflection schedule
		fmt.Println("\nReflection schedule?")
		fmt.Println("  1) daily (check on startup + every hour)")
		fmt.Println("  2) off (manual only)")
		fmt.Printf("Choose [1/2] (current: %s): ", c.Reflection.Schedule)
		if input := readLine(reader); input != "" {
			switch input {
			case "1":
				c.Reflection.Schedule = "daily"
			case "2":
				c.Reflection.Schedule = "off"
			default:
				c.Reflection.Schedule = input
			}
		}
	}

	// Save config
	if err := cfg.Save(c); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("\n✓ Config written to %s\n", cfg.ConfigPath())

	// Create data directories
	os.MkdirAll(cfg.DataDir(), 0o755)
	os.MkdirAll(c.Reflection.OutputDir, 0o755)
	fmt.Printf("✓ Data directory created at %s\n", cfg.DataDir())

	// Setup hooks
	if !initDefaults {
		fmt.Println("\nSet up hooks for which providers?")
		fmt.Println("  1) Claude Code")
		fmt.Println("  2) Gemini CLI")
		fmt.Println("  3) Codex")
		fmt.Println("  4) Skip")
		fmt.Print("Choose (comma-separated, e.g. 1,2): ")
		input := readLine(reader)

		if strings.Contains(input, "1") {
			scope := "global"
			fmt.Println("\nInstall Claude hooks globally or for this project?")
			fmt.Println("  1) global  — all sessions")
			fmt.Println("  2) project — this directory only")
			fmt.Print("Choose [1/2]: ")
			if s := readLine(reader); s == "2" {
				scope = "project"
			}

			hookScript := setup.FindHookScript()
			if hookScript != "" {
				if err := setup.SetupClaude(scope, hookScript); err != nil {
					fmt.Printf("  ✗ Claude setup failed: %v\n", err)
				} else {
					fmt.Printf("✓ Claude Code hooks configured (%s)\n", scope)
				}
			} else {
				fmt.Println("  ✗ Could not find claude-hook.sh — run from the project directory")
			}
		}

		if strings.Contains(input, "2") {
			if err := setup.SetupGemini(c.Port); err != nil {
				fmt.Printf("  ✗ Gemini setup failed: %v\n", err)
			} else {
				fmt.Println("✓ Gemini CLI hooks configured")
			}
		}

		if strings.Contains(input, "3") {
			if err := setup.SetupCodex(c.Port); err != nil {
				fmt.Printf("  ✗ Codex setup failed: %v\n", err)
			} else {
				fmt.Println("✓ Codex OTel export configured")
			}
		}
	}

	// Offer to start
	if !initDefaults {
		fmt.Print("\nStart the collector now? [Y/n]: ")
		input := readLine(reader)
		if input == "" || strings.ToLower(input) == "y" {
			return runStart(cmd, nil)
		}
	}

	fmt.Println("\nRun `ai-collector start` to start the collector.")
	return nil
}

func readLine(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
