package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/liamdn8/auto-worklog-agent/internal/activitywatch"
	"github.com/liamdn8/auto-worklog-agent/internal/agent"
	"github.com/liamdn8/auto-worklog-agent/internal/config"
)

func main() {
	var cfgFile string

	rootCmd := &cobra.Command{
		Use:   "awagent",
		Short: "ActivityWatch Git session tracker",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			awClient := activitywatch.NewClient(cfg.ActivityWatch)

			sessionTracker, err := agent.NewTracker(cfg, awClient)
			if err != nil {
				return fmt.Errorf("init tracker: %w", err)
			}

			log.Println("Starting ActivityWatch agent...")
			return sessionTracker.Run(ctx)
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %v", err)
	}
}
