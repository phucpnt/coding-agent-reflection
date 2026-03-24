package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/phuc/coding-agent-reflection/config"
	"github.com/phuc/coding-agent-reflection/internal/store"
)

func main() {
	cfg := config.Load()

	s, err := store.New(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	ctx := context.Background()

	// Default: show today's interactions
	days := 1
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "today":
			days = 1
		case "week":
			days = 7
		case "all":
			days = 3650
		case "reflections":
			printReflections(s)
			return
		case "stats":
			printStats(s)
			return
		default:
			fmt.Fprintf(os.Stderr, "Usage: query [today|week|all|reflections|stats]\n")
			os.Exit(1)
		}
	}

	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, -(days - 1))
	to := now.Add(time.Hour)

	interactions, err := s.QueryByDateRange(ctx, from, to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(1)
	}

	if len(interactions) == 0 {
		fmt.Println("No interactions found.")
		return
	}

	fmt.Printf("Found %d interactions:\n\n", len(interactions))
	for i, inter := range interactions {
		fmt.Printf("─── %d. [%s] %s ───\n", i+1, inter.Provider, inter.Ts.Format("15:04:05"))
		fmt.Printf("  Session:  %s\n", inter.SessionID)
		if inter.Project != "" {
			fmt.Printf("  Project:  %s\n", inter.Project)
		}
		fmt.Printf("  Prompt:   %s\n", truncate(inter.UserPrompt, 120))
		fmt.Printf("  Output:   %s\n", truncate(inter.AgentOutput, 120))
		if inter.ToolsUsed.Valid {
			fmt.Printf("  Tools:    %s\n", inter.ToolsUsed.String)
		}
		if inter.TokensPrompt.Valid {
			fmt.Printf("  Tokens:   %d in / %d out\n", inter.TokensPrompt.Int64, inter.TokensOutput.Int64)
		}
		fmt.Println()
	}
}

func printReflections(s *store.Store) {
	rows, err := s.DB().QueryContext(context.Background(),
		`SELECT date, summary, should_do, should_not_do FROM reflections ORDER BY date DESC LIMIT 10`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query reflections: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var dateStr string
		var summary, shouldDo, shouldNotDo string
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
}

func printStats(s *store.Store) {
	row := s.DB().QueryRowContext(context.Background(),
		`SELECT count(*), count(DISTINCT session_id), count(DISTINCT provider) FROM interactions`)
	var total, sessions, providers int
	row.Scan(&total, &sessions, &providers)

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
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
