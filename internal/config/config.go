package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	Repositories []string `json:"repositories"`
}

// SessionConfig controls session detection behavior.
type SessionConfig struct {
	IdleTimeoutMinutes int          `json:"idleTimeoutMinutes"`
	PollInterval       jsonDuration `json:"pollInterval"`
	FlushInterval      jsonDuration `json:"flushInterval"`
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

// LoadConfig loads configuration from JSON file, falling back to defaults.
func LoadConfig(path string) (Config, error) {
	cfg := defaultConfig()
	if path == "" {
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

	return cfg, nil
}

func defaultConfig() Config {
	repos := []string{}
	if cwd, err := os.Getwd(); err == nil {
		repos = append(repos, cwd)
	}

	return Config{
		ActivityWatch: ActivityWatchConfig{
			BaseURL:      "http://localhost:5600",
			BucketPrefix: "awagent",
			Machine:      hostnameOrUnknown(),
		},
		Git: GitConfig{
			Repositories: repos,
		},
		Session: SessionConfig{
			IdleTimeoutMinutes: 30,
			PollInterval:       jsonDuration{timeMS: int64((5 * time.Second).Milliseconds())},
			FlushInterval:      jsonDuration{timeMS: int64((15 * time.Second).Milliseconds())},
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
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}
