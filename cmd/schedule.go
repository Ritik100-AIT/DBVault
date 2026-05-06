package cmd

import (
	"fmt"
	"time"

	"github.com/dbvault/dbvault/internal/scheduler"
	"github.com/spf13/cobra"
)

// scheduleCmd is the parent command for schedule operations.
var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage scheduled backups",
}

var schedulerInstance = scheduler.NewScheduler()

// scheduleAddCmd adds a new backup schedule to the persistent schedule registry.
var scheduleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new scheduled backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		cron, _ := cmd.Flags().GetString("cron")
		name, _ := cmd.Flags().GetString("name")
		if cron == "" {
			return fmt.Errorf("cron expression is required")
		}
		if name == "" {
			name = fmt.Sprintf("schedule-%d", time.Now().Unix())
		}

		sch := scheduler.Schedule{
			ID:      name,
			Cron:    cron,
			Name:    name,
			NextRun: time.Now().Add(time.Hour),
		}

		if err := schedulerInstance.Add(sch); err != nil {
			return fmt.Errorf("failed to add schedule: %w", err)
		}
		fmt.Printf("Added schedule %s (%s)\n", sch.Name, sch.Cron)
		return nil
	},
}

// scheduleListCmd prints the persisted schedule list.
var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scheduled backups",
	RunE: func(cmd *cobra.Command, args []string) error {
		schedules, err := schedulerInstance.List()
		if err != nil {
			return fmt.Errorf("failed to list schedules: %w", err)
		}

		if len(schedules) == 0 {
			fmt.Println("No schedules found.")
			return nil
		}

		fmt.Println("Scheduled backups:")
		for _, sch := range schedules {
			fmt.Printf("- ID: %s, Cron: %s, Name: %s\n", sch.ID, sch.Cron, sch.Name)
		}
		return nil
	},
}

// scheduleRemoveCmd removes a schedule by ID from persistent storage.
var scheduleRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		if id == "" {
			return fmt.Errorf("schedule ID is required")
		}

		if err := schedulerInstance.Remove(id); err != nil {
			return fmt.Errorf("failed to remove schedule: %w", err)
		}
		fmt.Printf("Removed schedule %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.AddCommand(scheduleAddCmd, scheduleListCmd, scheduleRemoveCmd)
	scheduleAddCmd.Flags().String("cron", "", "Cron expression for the schedule")
	scheduleAddCmd.Flags().String("name", "", "Friendly schedule name")
	scheduleRemoveCmd.Flags().String("id", "", "Schedule ID to remove")
}
