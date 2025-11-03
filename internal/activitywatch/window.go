package activitywatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ErrWindowBucketMissing indicates that aw-watcher-window is not publishing a bucket for the configured machine.
var ErrWindowBucketMissing = errors.New("aw-watcher-window bucket not found")

// WindowEvent captures the subset of fields from aw-watcher-window we care about.
type WindowEvent struct {
	Timestamp time.Time  `json:"timestamp"`
	Duration  float64    `json:"duration"`
	Data      WindowData `json:"data"`
}

// WindowData describes the aw-watcher-window payload.
type WindowData struct {
	App   string `json:"app"`
	Title string `json:"title"`
}

// FetchWindowEvents pulls window watcher events from the ActivityWatch server since the provided instant.
// When since is zero, the server's default lookback is used. Limit bounds the maximum number of events returned.
func (c *Client) FetchWindowEvents(ctx context.Context, machine string, since time.Time, limit int) ([]WindowEvent, error) {
	bucketID := fmt.Sprintf("aw-watcher-window_%s", machine)
	values := url.Values{}
	if !since.IsZero() {
		values.Set("start", since.UTC().Format(time.RFC3339Nano))
	}
	if limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", limit))
	}

	endpoint := c.buildURL("api/0/buckets", bucketID, "events")
	if len(values) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, values.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create window request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch window events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrWindowBucketMissing
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch window events failed: status %s", resp.Status)
	}

	var events []WindowEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("decode window events: %w", err)
	}

	return events, nil
}
