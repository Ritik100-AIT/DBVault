package cmd

import (
	"fmt"

	"github.com/dbvault/dbvault/internal/config"
	"github.com/dbvault/dbvault/internal/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// testConnCmd verifies the database connection using the configured connector.
var testConnCmd = &cobra.Command{
	Use:   "test-connection",
	Short: "Test the database connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		connector := db.NewConnector(string(cfg.Database.Type))
		if connector == nil {
			return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
		}

		if err := connector.TestConnection(); err != nil {
			return fmt.Errorf("connection test failed: %w", err)
		}

		fmt.Println("Database connection test passed!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testConnCmd)
	testConnCmd.Flags().String("db", "", "Database type")
	testConnCmd.Flags().String("host", "", "Database host")
	testConnCmd.Flags().String("user", "", "Database user")
	testConnCmd.Flags().String("password", "", "Database password")
	testConnCmd.Flags().String("name", "", "Database name")

	viper.BindPFlag("database.type", testConnCmd.Flags().Lookup("db"))
	viper.BindPFlag("database.host", testConnCmd.Flags().Lookup("host"))
	viper.BindPFlag("database.username", testConnCmd.Flags().Lookup("user"))
	viper.BindPFlag("database.password", testConnCmd.Flags().Lookup("password"))
	viper.BindPFlag("database.name", testConnCmd.Flags().Lookup("name"))
}
