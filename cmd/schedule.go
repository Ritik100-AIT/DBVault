package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dbvault/dbvault/internal/backup"
	"github.com/dbvault/dbvault/internal/config"
	"github.com/dbvault/dbvault/internal/logger"
	"github.com/dbvault/dbvault/internal/notify"
	"github.com/dbvault/dbvault/internal/scheduler"
	"github.com/dbvault/dbvault/internal/storage"
	"github.com/robfig/cron/v3"
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

var scheduleRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run scheduled backups in the foreground",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := config.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		backend, err := storage.NewStorageBackend(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize storage backend: %w", err)
		}

		connector := newDBConnector(&cfg.Database)
		if connector == nil {
			return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
		}

		var notifier *notify.SlackNotifier
		if cfg.Notifications.Slack.Enabled {
			notifier = notify.NewSlackNotifier(cfg.Notifications.Slack.WebhookURL)
		}

		loggerInstance, err := logger.NewLogger(&cfg.Logging)
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		manager := backup.NewBackupManager(connector, backend, cfg, notifier, loggerInstance)

		schedules, err := schedulerInstance.List()
		if err != nil {
			return fmt.Errorf("failed to load schedules: %w", err)
		}
		if len(schedules) == 0 {
			return fmt.Errorf("no schedules configured")
		}

		cronRunner := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
		for _, schedule := range schedules {
			s := schedule
			_, err := cronRunner.AddFunc(s.Cron, func() {
				loggerInstance.Info(fmt.Sprintf("Scheduled backup triggered: %s", s.Name))
				if err := manager.Run(); err != nil {
					loggerInstance.Error(fmt.Sprintf("scheduled backup failed: %v", err))
				}
			})
			if err != nil {
				return fmt.Errorf("failed to schedule %s: %w", s.Name, err)
			}
		}

		cronRunner.Start()
		fmt.Println("Scheduler started. Press Ctrl+C to stop.")

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		<-ctx.Done()

		fmt.Println("Shutting down scheduler...")
		cronRunner.Stop()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.AddCommand(scheduleAddCmd, scheduleListCmd, scheduleRemoveCmd, scheduleRunCmd)
	scheduleAddCmd.Flags().String("cron", "", "Cron expression for the schedule")
	scheduleAddCmd.Flags().String("name", "", "Friendly schedule name")
	scheduleRemoveCmd.Flags().String("id", "", "Schedule ID to remove")
}
