package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup for a supported database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("backup command stub")
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().String("db", "", "Database type: mysql | postgres | mongodb | sqlite")
	backupCmd.Flags().String("host", "", "Database host")
	backupCmd.Flags().String("user", "", "Database user")
	backupCmd.Flags().String("password", "", "Database password")
	backupCmd.Flags().String("name", "", "Database name")
	backupCmd.Flags().String("type", "full", "Backup type: full | incremental | differential")
	backupCmd.Flags().String("storage", "local", "Storage backend: local | s3")
	backupCmd.Flags().String("compress", "gzip", "Compression method: gzip | none")
}
