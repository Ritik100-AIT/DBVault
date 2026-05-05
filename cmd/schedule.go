package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage scheduled backups",
}

var scheduleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new scheduled backup",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("schedule add command stub")
	},
}

var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scheduled backups",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("schedule list command stub")
	},
}

var scheduleRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a schedule",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("schedule remove command stub")
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.AddCommand(scheduleAddCmd, scheduleListCmd, scheduleRemoveCmd)
	scheduleAddCmd.Flags().String("cron", "", "Cron expression for the schedule")
	scheduleAddCmd.Flags().String("name", "", "Friendly schedule name")
}
