package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/dbvault/dbvault/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage DBVault configuration",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the current config",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Println("Current configuration:")
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configViewCmd)
}
