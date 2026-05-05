package cmd

import (
	"fmt"

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
		fmt.Println("config view command stub")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configViewCmd)
}
