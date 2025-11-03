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
	var overrideAWURL string
	var overrideMachine string

	rootCmd := &cobra.Command{
		Use:   "awagent",
		Short: "ActivityWatch Git session tracker",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if overrideAWURL != "" {
				cfg.ActivityWatch.BaseURL = overrideAWURL
			}

			if overrideMachine != "" {
				cfg.ActivityWatch.Machine = overrideMachine
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
	rootCmd.PersistentFlags().StringVar(&overrideAWURL, "aw-url", "", "override ActivityWatch server URL")
	rootCmd.PersistentFlags().StringVar(&overrideMachine, "machine", "", "override machine identifier reported to ActivityWatch")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %v", err)
	}
}
