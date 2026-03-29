package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cfg "github.com/phuc/coding-agent-reflection/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or edit configuration",
	RunE:  runConfigView,
}

var configSetCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigView(cmd *cobra.Command, args []string) error {
	c, err := cfg.Load()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	fmt.Printf("# Config file: %s\n\n", cfg.ConfigPath())
	fmt.Print(string(data))
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]
	if err := cfg.SetValue(key, value); err != nil {
		return err
	}
	fmt.Printf("Set %s = %s\n", key, value)
	fmt.Printf("Written to %s\n", cfg.ConfigPath())
	return nil
}
