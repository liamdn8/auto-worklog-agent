package agent

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/liamdn8/auto-worklog-agent/internal/activitywatch"
	"github.com/liamdn8/auto-worklog-agent/internal/config"
	"github.com/liamdn8/auto-worklog-agent/internal/gitinfo"
	"github.com/liamdn8/auto-worklog-agent/internal/session"
)

const (
	bucketTypeWorkSession = "app.awagent.worksession"
)

// Tracker coordinates filesystem activity tracking and publishes work sessions to ActivityWatch.
type Tracker struct {
	cfg          config.Config
	awClient     *activitywatch.Client
	repos        []gitinfo.Info
	idleTimeout  time.Duration
	flushEvery   time.Duration
	sessions     map[string]*session.State
	mu           sync.Mutex
	stopWatchers context.CancelFunc
}

// NewTracker builds a Tracker from configuration and client dependencies.
func NewTracker(cfg config.Config, awClient *activitywatch.Client) (*Tracker, error) {
	if len(cfg.Git.Repositories) == 0 {
		return nil, errors.New("no repositories configured")
	}

	repos := make([]gitinfo.Info, 0, len(cfg.Git.Repositories))
	for _, path := range cfg.Git.Repositories {
		repo, err := gitinfo.Discover(path)
		if err != nil {
			return nil, fmt.Errorf("discover repo %s: %w", path, err)
		}
		repos = append(repos, repo)
	}

	tracker := &Tracker{
		cfg:         cfg,
		awClient:    awClient,
		repos:       repos,
		idleTimeout: time.Duration(cfg.Session.IdleTimeoutMinutes) * time.Minute,
		flushEvery:  cfg.Session.FlushInterval.Duration(),
		sessions:    make(map[string]*session.State),
	}

	if tracker.flushEvery == 0 {
		tracker.flushEvery = cfg.Session.PollInterval.Duration()
	}

	if tracker.flushEvery == 0 {
		tracker.flushEvery = 30 * time.Second
	}

	if tracker.idleTimeout == 0 {
		tracker.idleTimeout = 30 * time.Minute
	}

	return tracker, nil
}

// Run starts the tracker loop.
func (t *Tracker) Run(ctx context.Context) error {
	watchCtx, cancel := context.WithCancel(ctx)
	t.stopWatchers = cancel
	defer cancel()

	events := make(chan repoEvent, 64)

	if err := t.startWatchers(watchCtx, events); err != nil {
		return err
	}

	flushTicker := time.NewTicker(t.flushEvery)
	defer flushTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.flushAll(context.Background())
			return ctx.Err()
		case evt := <-events:
			t.recordEvent(evt)
		case <-flushTicker.C:
			t.flushExpired(ctx)
		}
	}
}

func (t *Tracker) startWatchers(ctx context.Context, events chan<- repoEvent) error {
	var wg sync.WaitGroup
	repoCtx, repoCancel := context.WithCancel(ctx)

	for _, repo := range t.repos {
		repo := repo
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			repoCancel()
			return fmt.Errorf("create watcher: %w", err)
		}

		if err := addRecursive(watcher, repo.Path); err != nil {
			watcher.Close()
			repoCancel()
			return fmt.Errorf("watch repo %s: %w", repo.Path, err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer watcher.Close()
			for {
				select {
				case <-repoCtx.Done():
					return
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
						select {
						case events <- repoEvent{repo: repo, when: time.Now(), path: event.Name}:
						case <-repoCtx.Done():
							return
						}
					}
					if event.Op&fsnotify.Create != 0 {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							_ = watcher.Add(event.Name)
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Printf("watcher error for %s: %v", repo.Path, err)
				}
			}
		}()
	}

	go func() {
		<-ctx.Done()
		repoCancel()
		wg.Wait()
	}()

	return nil
}

func (t *Tracker) recordEvent(evt repoEvent) {
	branch, err := gitinfo.CurrentBranch(evt.repo.Path)
	if err != nil {
		log.Printf("resolve branch for %s: %v", evt.repo.Path, err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	repoKey := evt.repo.Path
	sess, ok := t.sessions[repoKey]
	if !ok {
		sess = session.NewState(evt.repo, branch, evt.when)
		t.sessions[repoKey] = sess
		return
	}

	sess.Touch(branch, evt.when)
}

func (t *Tracker) flushExpired(ctx context.Context) {
	t.mu.Lock()
	sessionsCopy := make([]*session.State, 0, len(t.sessions))
	keys := make([]string, 0, len(t.sessions))
	for key, sess := range t.sessions {
		if time.Since(sess.LastActivity) >= t.idleTimeout {
			sessionsCopy = append(sessionsCopy, sess)
			keys = append(keys, key)
		}
	}
	t.mu.Unlock()

	for i, sess := range sessionsCopy {
		if err := t.publishSession(ctx, sess); err != nil {
			log.Printf("publish session %s: %v", sess.Repo.Path, err)
			continue
		}
		t.mu.Lock()
		delete(t.sessions, keys[i])
		t.mu.Unlock()
	}
}

func (t *Tracker) flushAll(ctx context.Context) {
	t.mu.Lock()
	sessionsCopy := make([]*session.State, 0, len(t.sessions))
	for _, sess := range t.sessions {
		sessionsCopy = append(sessionsCopy, sess)
	}
	t.sessions = make(map[string]*session.State)
	t.mu.Unlock()

	for _, sess := range sessionsCopy {
		if err := t.publishSession(ctx, sess); err != nil {
			log.Printf("publish session %s: %v", sess.Repo.Path, err)
		}
	}
}

func (t *Tracker) publishSession(ctx context.Context, sess *session.State) error {
	if sess.Duration() <= 0 {
		return nil
	}

	bucketID := bucketIDForRepo(t.cfg.ActivityWatch.BucketPrefix, sess.Repo.Name)

	data := map[string]any{
		"gitUser":    sess.Repo.User,
		"gitEmail":   sess.Repo.Email,
		"repoName":   sess.Repo.Name,
		"repoPath":   sess.Repo.Path,
		"branch":     sess.Branch,
		"remote":     sess.Repo.Remote,
		"eventCount": sess.Events,
	}

	event := activitywatch.Event{
		Timestamp: sess.Start,
		End:       sess.LastActivity,
		Duration:  sess.Duration(),
		Data:      data,
	}

	if err := t.awClient.RecordEvent(ctx, bucketID, bucketTypeWorkSession, event); err != nil {
		return fmt.Errorf("record event: %w", err)
	}

	return nil
}

func addRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir(d.Name()) {
				return filepath.SkipDir
			}
			if err := w.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}

func skipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "target", "build":
		return true
	}
	return false
}

var bucketSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func bucketIDForRepo(prefix, repoName string) string {
	repoID := bucketSanitizer.ReplaceAllString(strings.ToLower(repoName), "-")
	repoID = strings.Trim(repoID, "-")
	if prefix == "" {
		return repoID
	}
	return fmt.Sprintf("%s.%s", prefix, repoID)
}

type repoEvent struct {
	repo gitinfo.Info
	when time.Time
	path string
}
