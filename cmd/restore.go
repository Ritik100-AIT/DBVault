package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup from local storage or remote backend",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("restore command stub")
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().String("source", "", "Backup source path or key")
	restoreCmd.Flags().String("db", "", "Database type: mysql | postgres | mongodb | sqlite")
	restoreCmd.Flags().Bool("verify", true, "Verify checksum before restore")
}
