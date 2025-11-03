package activitywatch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/liamdn8/auto-worklog-agent/internal/config"
)

// Client wraps interactions with an ActivityWatch server.
type Client struct {
	http       *http.Client
	cfg        config.ActivityWatchConfig
	bucketOnce sync.Map
}

// Event represents a generic ActivityWatch event payload.
type Event struct {
	Timestamp time.Time      `json:"timestamp"`
	End       time.Time      `json:"end"`
	Duration  time.Duration  `json:"duration"`
	Data      map[string]any `json:"data"`
}

// NewClient prepares a new ActivityWatch client.
func NewClient(cfg config.ActivityWatchConfig) *Client {
	return &Client{
		http: &http.Client{Timeout: 10 * time.Second},
		cfg:  cfg,
	}
}

// RecordEvent ensures a bucket exists and posts the given event.
func (c *Client) RecordEvent(ctx context.Context, bucketID, bucketType string, event Event) error {
	if err := c.ensureBucket(ctx, bucketID, bucketType); err != nil {
		return err
	}

	payload := map[string]any{
		"timestamp": event.Timestamp.UTC(),
		"duration":  event.Duration.Seconds(),
		"data":      event.Data,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	url := c.buildURL("api/0/buckets", bucketID, "events")
	log.Printf("ActivityWatch: POST %s payload=%s", url, body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("post event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("post event failed: status %s", resp.Status)
	}

	log.Printf("ActivityWatch: recorded event bucket=%s duration=%s events=%v branch=%v", bucketID, event.Duration, event.Data["eventCount"], event.Data["branch"])

	return nil
}

func (c *Client) ensureBucket(ctx context.Context, bucketID, bucketType string) error {
	if _, ok := c.bucketOnce.Load(bucketID); ok {
		return nil
	}

	payload := map[string]any{
		"client":   "awagent",
		"type":     bucketType,
		"hostname": c.cfg.Machine,
		"name":     bucketID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal bucket payload: %w", err)
	}

	url := c.buildURL("api/0/buckets", bucketID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create bucket request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusNotModified {
		c.bucketOnce.Store(bucketID, struct{}{})
		return nil
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("create bucket failed: status %s", resp.Status)
	}

	c.bucketOnce.Store(bucketID, struct{}{})
	log.Printf("ActivityWatch: ensured bucket %s (type=%s)", bucketID, bucketType)

	return nil
}

func (c *Client) buildURL(parts ...string) string {
	trimmed := strings.TrimSuffix(c.cfg.BaseURL, "/")
	joined := path.Join(parts...)
	return fmt.Sprintf("%s/%s", trimmed, joined)
}
