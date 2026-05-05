package cmd

import (
	"fmt"
	"log"

	"github.com/dbvault/dbvault/internal/config"
	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var testConnCmd = &cobra.Command{
	Use:   "test-connection",
	Short: "Test the database connection",
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Override config with CLI flags
		if dbType, _ := cmd.Flags().GetString("db"); dbType != "" {
			cfg.Database.Type = models.DBType(dbType)
		}
		if host, _ := cmd.Flags().GetString("host"); host != "" {
			cfg.Database.Host = host
		}
		if user, _ := cmd.Flags().GetString("user"); user != "" {
			cfg.Database.User = user
		}
		if password, _ := cmd.Flags().GetString("password"); password != "" {
			cfg.Database.Password = password
		}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			cfg.Database.Name = name
		}

		// Create DB connector
		connector := db.NewConnector(string(cfg.Database.Type))
		if connector == nil {
			log.Fatalf("Unsupported database type: %s", cfg.Database.Type)
		}

		// Test connection
		if err := connector.TestConnection(); err != nil {
			log.Fatalf("Connection test failed: %v", err)
		}

		fmt.Println("Database connection test passed!")
	},
}

func init() {
	rootCmd.AddCommand(testConnCmd)
	testConnCmd.Flags().String("db", "", "Database type")
	testConnCmd.Flags().String("host", "", "Database host")
	testConnCmd.Flags().String("user", "", "Database user")
	testConnCmd.Flags().String("password", "", "Database password")
	testConnCmd.Flags().String("name", "", "Database name")

	// Bind flags to viper
	viper.BindPFlag("database.type", testConnCmd.Flags().Lookup("db"))
	viper.BindPFlag("database.host", testConnCmd.Flags().Lookup("host"))
	viper.BindPFlag("database.username", testConnCmd.Flags().Lookup("user"))
	viper.BindPFlag("database.password", testConnCmd.Flags().Lookup("password"))
	viper.BindPFlag("database.name", testConnCmd.Flags().Lookup("name"))
}
