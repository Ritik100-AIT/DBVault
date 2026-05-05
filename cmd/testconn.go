package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testConnCmd = &cobra.Command{
	Use:   "test-connection",
	Short: "Test the database connection",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test-connection command stub")
	},
}

func init() {
	rootCmd.AddCommand(testConnCmd)
	testConnCmd.Flags().String("db", "", "Database type")
	testConnCmd.Flags().String("host", "", "Database host")
	testConnCmd.Flags().String("user", "", "Database user")
	testConnCmd.Flags().String("password", "", "Database password")
	testConnCmd.Flags().String("name", "", "Database name")
}
