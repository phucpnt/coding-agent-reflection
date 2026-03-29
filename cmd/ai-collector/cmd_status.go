package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	cfg "github.com/phuc/coding-agent-reflection/internal/config"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check collector health and show recent interactions",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	c, _ := cfg.Load()
	baseURL := fmt.Sprintf("http://localhost:%d", c.Port)

	// Check PID
	pid := readPid(cfg.PidPath())
	if pid > 0 && processRunning(pid) {
		fmt.Printf("Collector: running (PID %d)\n", pid)
	} else {
		fmt.Println("Collector: not running")
		return nil
	}

	// Health check
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Println("Health:    unreachable")
		return nil
	}
	defer resp.Body.Close()
	fmt.Printf("Health:    ok (port %d)\n", c.Port)

	// Interaction count
	resp2, err := http.Get(baseURL + "/interactions")
	if err != nil {
		return nil
	}
	defer resp2.Body.Close()
	body, _ := io.ReadAll(resp2.Body)
	var interactions []any
	json.Unmarshal(body, &interactions)
	fmt.Printf("Recent:    %d interactions (last 7 days)\n", len(interactions))

	fmt.Printf("Config:    %s\n", cfg.ConfigPath())
	fmt.Printf("Database:  %s\n", c.DBPath)
	fmt.Printf("Logs:      %s\n", cfg.LogPath())

	return nil
}
