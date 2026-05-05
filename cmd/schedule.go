package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/dbvault/dbvault/internal/scheduler"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage scheduled backups",
}

var schedulerInstance = scheduler.NewScheduler()

var scheduleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new scheduled backup",
	Run: func(cmd *cobra.Command, args []string) {
		cron, _ := cmd.Flags().GetString("cron")
		name, _ := cmd.Flags().GetString("name")
		if cron == "" {
			log.Fatal("Cron expression is required")
		}
		if name == "" {
			name = fmt.Sprintf("schedule-%d", time.Now().Unix())
		}

		sch := scheduler.Schedule{
			ID:      name,
			Cron:    cron,
			Name:    name,
			NextRun: time.Now().Add(time.Hour), // Placeholder
		}

		if err := schedulerInstance.Add(sch); err != nil {
			log.Fatalf("Failed to add schedule: %v", err)
		}
	},
}

var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scheduled backups",
	Run: func(cmd *cobra.Command, args []string) {
		schedules, err := schedulerInstance.List()
		if err != nil {
			log.Fatalf("Failed to list schedules: %v", err)
		}

		if len(schedules) == 0 {
			fmt.Println("No schedules found.")
			return
		}

		fmt.Println("Scheduled backups:")
		for _, sch := range schedules {
			fmt.Printf("- ID: %s, Cron: %s, Name: %s\n", sch.ID, sch.Cron, sch.Name)
		}
	},
}

var scheduleRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a schedule",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		if id == "" {
			log.Fatal("Schedule ID is required")
		}

		if err := schedulerInstance.Remove(id); err != nil {
			log.Fatalf("Failed to remove schedule: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.AddCommand(scheduleAddCmd, scheduleListCmd, scheduleRemoveCmd)
	scheduleAddCmd.Flags().String("cron", "", "Cron expression for the schedule")
	scheduleAddCmd.Flags().String("name", "", "Friendly schedule name")
	scheduleRemoveCmd.Flags().String("id", "", "Schedule ID to remove")
}
