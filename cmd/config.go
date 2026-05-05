package cmd

import (
	"encoding/json"
	"fmt"
	"log"

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
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Pretty print as JSON
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal config: %v", err)
		}

		fmt.Println("Current configuration:")
		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configViewCmd)
}
