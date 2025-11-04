package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config represents application configuration.
type Config struct {
	ActivityWatch ActivityWatchConfig `json:"activityWatch"`
	Git           GitConfig           `json:"git"`
	Session       SessionConfig       `json:"session"`
}

// ActivityWatchConfig holds the aw-server integration settings.
type ActivityWatchConfig struct {
	BaseURL      string `json:"baseURL"`
	BucketPrefix string `json:"bucketPrefix"`
	Machine      string `json:"machine"`
}

// GitConfig configures git metadata resolution.
type GitConfig struct {
	Repositories      []string `json:"repositories"`
	Roots             []string `json:"roots"`
	MaxDepth          int      `json:"maxDepth"`
	RescanIntervalMin int      `json:"rescanIntervalMin"`
}

// SessionConfig controls session detection behavior.
type SessionConfig struct {
	IdleTimeoutMinutes int          `json:"idleTimeoutMinutes"`
	PollInterval       jsonDuration `json:"pollInterval"`
	FlushInterval      jsonDuration `json:"flushInterval"`
	PulseTime          jsonDuration `json:"pulseTime"`
}

type jsonDuration struct {
	timeMS int64
}

func (d *jsonDuration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		d.timeMS = dur.Milliseconds()
	case float64:
		d.timeMS = int64(value)
	default:
		return errors.New("invalid duration format")
	}
	return nil
}

func (d jsonDuration) Duration() time.Duration {
	return time.Duration(d.timeMS) * time.Millisecond
}

func newJSONDuration(d time.Duration) jsonDuration {
	return jsonDuration{timeMS: int64(d.Milliseconds())}
}

// LoadConfig loads configuration from JSON file, falling back to defaults.
func LoadConfig(path string) (Config, error) {
	cfg := defaultConfig()
	if path == "" {
		if err := cfg.normalize(); err != nil {
			return cfg, err
		}
		return cfg, nil
	}

	expanded, err := expandPath(path)
	if err != nil {
		return cfg, fmt.Errorf("expand path: %w", err)
	}

	file, err := os.ReadFile(expanded)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(file, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.normalize(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func defaultConfig() Config {
	repos := []string{}
	if cwd, err := os.Getwd(); err == nil {
		repos = append(repos, cwd)
	}

	roots := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, home)
	}

	return Config{
		ActivityWatch: ActivityWatchConfig{
			BaseURL:      "http://localhost:5600",
			BucketPrefix: "awagent",
			Machine:      hostnameOrUnknown(),
		},
		Git: GitConfig{
			Repositories:      repos,
			Roots:             roots,
			MaxDepth:          5,
			RescanIntervalMin: 5,
		},
		Session: SessionConfig{
			IdleTimeoutMinutes: 30,
			PollInterval:       newJSONDuration(5 * time.Second),
			FlushInterval:      newJSONDuration(15 * time.Second),
			PulseTime:          newJSONDuration(10 * time.Second),
		},
	}
}

func hostnameOrUnknown() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return host
}

func expandPath(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}
	path = os.ExpandEnv(path)
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

func expandPaths(paths []string) ([]string, error) {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		expanded, err := expandPath(trimmed)
		if err != nil {
			return nil, err
		}
		out = append(out, filepath.Clean(expanded))
	}
	return out, nil
}

func (cfg *Config) normalize() error {
	repos, err := expandPaths(cfg.Git.Repositories)
	if err != nil {
		return fmt.Errorf("expand git repositories: %w", err)
	}
	cfg.Git.Repositories = repos

	roots, err := expandPaths(cfg.Git.Roots)
	if err != nil {
		return fmt.Errorf("expand git roots: %w", err)
	}
	cfg.Git.Roots = roots

	// Zero or negative maxDepth means unlimited
	if cfg.Git.MaxDepth < 0 {
		cfg.Git.MaxDepth = 0
	}

	if cfg.ActivityWatch.BaseURL == "" {
		cfg.ActivityWatch.BaseURL = "http://localhost:5600"
	}
	if cfg.ActivityWatch.BucketPrefix == "" {
		cfg.ActivityWatch.BucketPrefix = "awagent"
	}
	if cfg.ActivityWatch.Machine == "" {
		cfg.ActivityWatch.Machine = hostnameOrUnknown()
	}

	return nil
}
