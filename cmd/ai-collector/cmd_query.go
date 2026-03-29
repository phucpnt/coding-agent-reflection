package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cfg "github.com/phuc/coding-agent-reflection/internal/config"
	"github.com/phuc/coding-agent-reflection/internal/store"
)

var queryCmd = &cobra.Command{
	Use:   "query [today|week|all|reflections|stats]",
	Short: "Query interactions from the database",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)
}

func runQuery(cmd *cobra.Command, args []string) error {
	c, _ := cfg.Load()

	s, err := store.New(c.DBPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer s.Close()

	mode := "today"
	if len(args) > 0 {
		mode = args[0]
	}

	ctx := context.Background()

	switch mode {
	case "today":
		return queryInteractions(ctx, s, 1)
	case "week":
		return queryInteractions(ctx, s, 7)
	case "all":
		return queryInteractions(ctx, s, 3650)
	case "reflections":
		return queryReflections(s)
	case "stats":
		return queryStats(s)
	default:
		return fmt.Errorf("unknown query mode: %s (use today|week|all|reflections|stats)", mode)
	}
}

func queryInteractions(ctx context.Context, s *store.Store, days int) error {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, -(days - 1))
	to := now.Add(time.Hour)

	interactions, err := s.QueryByDateRange(ctx, from, to)
	if err != nil {
		return err
	}

	if len(interactions) == 0 {
		fmt.Println("No interactions found.")
		return nil
	}

	fmt.Printf("Found %d interactions:\n\n", len(interactions))
	for i, inter := range interactions {
		fmt.Printf("─── %d. [%s] %s ───\n", i+1, inter.Provider, inter.Ts.Format("15:04:05"))
		fmt.Printf("  Session:  %s\n", inter.SessionID)
		if inter.Project != "" {
			fmt.Printf("  Project:  %s\n", inter.Project)
		}
		fmt.Printf("  Prompt:   %s\n", truncStr(inter.UserPrompt, 120))
		fmt.Printf("  Output:   %s\n", truncStr(inter.AgentOutput, 120))
		if inter.ToolsUsed.Valid {
			fmt.Printf("  Tools:    %s\n", inter.ToolsUsed.String)
		}
		fmt.Println()
	}
	return nil
}

func queryReflections(s *store.Store) error {
	rows, err := s.DB().QueryContext(context.Background(),
		`SELECT date, summary, should_do, should_not_do FROM reflections ORDER BY date DESC LIMIT 10`)
	if err != nil {
		return err
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var dateStr, summary, shouldDo, shouldNotDo string
		rows.Scan(&dateStr, &summary, &shouldDo, &shouldNotDo)
		fmt.Printf("═══ %s ═══\n", dateStr)
		fmt.Printf("Summary:       %s\n", summary)
		fmt.Printf("Should Do:     %s\n", shouldDo)
		fmt.Printf("Should Not Do: %s\n", shouldNotDo)
		fmt.Println()
	}
	if !found {
		fmt.Println("No reflections found.")
	}
	return nil
}

func queryStats(s *store.Store) error {
	var total, sessions, providers int
	s.DB().QueryRowContext(context.Background(),
		`SELECT count(*), count(DISTINCT session_id), count(DISTINCT provider) FROM interactions`).
		Scan(&total, &sessions, &providers)

	fmt.Printf("Total interactions: %d\n", total)
	fmt.Printf("Unique sessions:   %d\n", sessions)
	fmt.Printf("Providers:         %d\n", providers)

	rows, _ := s.DB().QueryContext(context.Background(),
		`SELECT provider, count(*) as cnt FROM interactions GROUP BY provider ORDER BY cnt DESC`)
	if rows != nil {
		defer rows.Close()
		fmt.Println("\nBy provider:")
		for rows.Next() {
			var provider string
			var cnt int
			rows.Scan(&provider, &cnt)
			fmt.Printf("  %-10s %d\n", provider, cnt)
		}
	}

	rows2, _ := s.DB().QueryContext(context.Background(),
		`SELECT date(ts) as day, count(*) FROM interactions GROUP BY day ORDER BY day DESC LIMIT 7`)
	if rows2 != nil {
		defer rows2.Close()
		fmt.Println("\nLast 7 days:")
		for rows2.Next() {
			var day string
			var cnt int
			rows2.Scan(&day, &cnt)
			fmt.Printf("  %s  %s\n", day, strings.Repeat("█", cnt))
		}
	}
	return nil
}

func truncStr(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
