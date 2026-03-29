package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "ai-collector",
	Short: "AI Interaction Collector — capture and reflect on coding agent interactions",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
