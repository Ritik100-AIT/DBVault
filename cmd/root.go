package cmd

import (
	"fmt"
	"os"

	"github.com/dbvault/dbvault/internal/config"
	"github.com/spf13/cobra"
)

// rootCmd is the primary Cobra command for the DBVault CLI.
var rootCmd = &cobra.Command{
	Use:   "dbvault",
	Short: "DBVault CLI",
	Long:  "DBVault is a database backup, restore, and scheduler CLI tool.",
}

// RootCmd exposes the root Cobra command for integration tests.
var RootCmd = rootCmd

// Execute runs the root command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.dbvault/config.yaml)")
}

// initConfig initializes Viper configuration before commands execute.
func initConfig() {
	cfgPath, _ := rootCmd.PersistentFlags().GetString("config")
	if err := config.SetupViper(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize config: %v\n", err)
		os.Exit(1)
	}
}
