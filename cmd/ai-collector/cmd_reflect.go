package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cfg "github.com/phuc/coding-agent-reflection/internal/config"
)

var reflectCmd = &cobra.Command{
	Use:   "reflect",
	Short: "Trigger a reflection for today or a specific date",
	RunE:  runReflect,
}

func init() {
	reflectCmd.Flags().String("date", "", "Date to reflect on (YYYY-MM-DD, default: today)")
	rootCmd.AddCommand(reflectCmd)
}

func runReflect(cmd *cobra.Command, args []string) error {
	c, _ := cfg.Load()
	baseURL := fmt.Sprintf("http://localhost:%d", c.Port)

	date, _ := cmd.Flags().GetString("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	fmt.Printf("Triggering reflection for %s...\n", date)

	body := fmt.Sprintf(`{"date":"%s"}`, date)
	resp, err := http.Post(baseURL+"/jobs/daily-reflection", "application/json", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("collector not reachable — is it running? (ai-collector start)")
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(respBody, &result)

	if resp.StatusCode != 200 {
		return fmt.Errorf("reflection failed: %s", result["error"])
	}

	if msg, ok := result["message"]; ok {
		fmt.Println(msg)
	} else if summary, ok := result["summary"]; ok {
		fmt.Printf("\nReflection for %s:\n%s\n", date, summary)
	}

	return nil
}
