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
	var verbose bool
	var testMode bool

	rootCmd := &cobra.Command{
		Use:   "awagent",
		Short: "ActivityWatch Git session tracker",
		Long: `Auto-discovery agent that tracks development activity across Git repositories.
		
Requires aw-watcher-window to be running to detect IDE activity.
Use --test mode to simulate activity for testing without aw-watcher-window.`,
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

			log.Printf("Starting ActivityWatch agent (verbose=%v, test=%v)...", verbose, testMode)
			log.Printf("Configuration: server=%s machine=%s", cfg.ActivityWatch.BaseURL, cfg.ActivityWatch.Machine)
			log.Printf("Git scan roots: %v (maxDepth=%d, rescan=%dm)", cfg.Git.Roots, cfg.Git.MaxDepth, cfg.Git.RescanIntervalMin)

			if testMode {
				log.Println("TEST MODE: Simulating IDE activity without aw-watcher-window")
				return sessionTracker.RunTest(ctx)
			}

			return sessionTracker.Run(ctx)
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file")
	rootCmd.PersistentFlags().StringVar(&overrideAWURL, "aw-url", "", "override ActivityWatch server URL")
	rootCmd.PersistentFlags().StringVar(&overrideMachine, "machine", "", "override machine identifier reported to ActivityWatch")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&testMode, "test", false, "run in test mode (simulate activity without aw-watcher-window)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %v", err)
	}
}
