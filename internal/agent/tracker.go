package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/liamdn8/auto-worklog-agent/internal/activitywatch"
	"github.com/liamdn8/auto-worklog-agent/internal/config"
	"github.com/liamdn8/auto-worklog-agent/internal/gitinfo"
	"github.com/liamdn8/auto-worklog-agent/internal/session"
	"github.com/liamdn8/auto-worklog-agent/internal/watcher"
)

const (
	bucketTypeWorkSession = "app.awagent.worksession"
	windowPollLimit       = 100
)

// Tracker coordinates window activity tracking and publishes work sessions to ActivityWatch.
type Tracker struct {
	cfg      config.Config
	awClient *activitywatch.Client

	idleTimeout time.Duration
	flushEvery  time.Duration

	sessions map[string]*session.State
	mu       sync.Mutex

	repoMu sync.RWMutex
	repos  map[string]gitinfo.Info
}

// NewTracker builds a Tracker from configuration and client dependencies.
func NewTracker(cfg config.Config, awClient *activitywatch.Client) (*Tracker, error) {
	tracker := &Tracker{
		cfg:         cfg,
		awClient:    awClient,
		idleTimeout: time.Duration(cfg.Session.IdleTimeoutMinutes) * time.Minute,
		flushEvery:  cfg.Session.FlushInterval.Duration(),
		sessions:    make(map[string]*session.State),
		repos:       make(map[string]gitinfo.Info),
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

	tracker.refreshRepositories()

	log.Printf(
		"Tracker configured: repositories=%d idleTimeout=%s flushInterval=%s",
		len(tracker.repos),
		tracker.idleTimeout,
		tracker.flushEvery,
	)

	if len(tracker.repos) == 0 {
		log.Printf("Repository scan did not locate any git repositories; verify git.roots and git.maxDepth settings")
	}

	return tracker, nil
}

// RunTest runs the tracker in test mode, simulating IDE activity for discovered repositories.
func (t *Tracker) RunTest(ctx context.Context) error {
	log.Println("TEST MODE: Simulating activity for discovered repositories")

	go t.repoScanLoop(ctx)

	events := make(chan repoEvent, 64)
	go t.testActivityLoop(ctx, events)

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

func (t *Tracker) testActivityLoop(ctx context.Context, events chan<- repoEvent) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.repoMu.RLock()
			for _, repo := range t.repos {
				log.Printf("TEST: Simulating activity for repo=%s", repo.Name)
				select {
				case events <- repoEvent{repo: repo, when: time.Now(), path: "[test-activity]"}:
				case <-ctx.Done():
					t.repoMu.RUnlock()
					return
				}
			}
			t.repoMu.RUnlock()
		}
	}
}

// Run starts the tracker loop with embedded window watching.
func (t *Tracker) Run(ctx context.Context) error {
	events := make(chan repoEvent, 64)
	go t.embeddedWindowLoop(ctx, events)
	go t.repoScanLoop(ctx)

	flushTicker := time.NewTicker(t.flushEvery)
	defer flushTicker.Stop()

	log.Println("Embedded window watcher started - no aw-watcher-window required!")

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

func (t *Tracker) embeddedWindowLoop(ctx context.Context, events chan<- repoEvent) {
	interval := t.cfg.Session.PollInterval.Duration()
	if interval <= 0 {
		interval = 1 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			window, err := watcher.GetActiveWindow()
			if err != nil {
				log.Printf("Failed to get active window: %v", err)
				continue
			}

			if window.App == "" && window.Title == "" {
				continue
			}

			app := strings.ToLower(window.App)
			title := strings.ToLower(window.Title)

			if !matchesKnownIDE(app, title) {
				continue
			}

			repo, ok := t.matchCachedRepo(title)
			if !ok {
				continue
			}

			select {
			case events <- repoEvent{repo: repo, when: time.Now(), path: fmt.Sprintf("[window] %s - %s", window.App, window.Title)}:
			case <-ctx.Done():
				return
			}
		}
	}
}

var knownIDEApps = []string{
	"code",
	"code-oss",
	"visual studio code",
	"vscode",
	"idea",
	"intellij",
	"clion",
	"goland",
	"webstorm",
	"phpstorm",
	"pycharm",
	"rider",
	"dataspell",
	"rubymine",
	"vim",
	"nvim",
	"neovim",
	"emacs",
	"sublime",
}

func matchesKnownIDE(app, title string) bool {
	for _, candidate := range knownIDEApps {
		if strings.Contains(app, candidate) || strings.Contains(title, candidate) {
			return true
		}
	}
	return false
}

func (t *Tracker) matchCachedRepo(lowerTitle string) (gitinfo.Info, bool) {
	lowerTitle = strings.ToLower(lowerTitle)
	t.repoMu.RLock()
	defer t.repoMu.RUnlock()
	for _, info := range t.repos {
		name := strings.ToLower(info.Name)
		path := strings.ToLower(info.Path)
		if strings.Contains(lowerTitle, name) || strings.Contains(lowerTitle, filepath.Base(path)) || strings.Contains(lowerTitle, path) {
			return info, true
		}
	}
	return gitinfo.Info{}, false
}

func (t *Tracker) repoScanLoop(ctx context.Context) {
	rescanInterval := time.Duration(t.cfg.Git.RescanIntervalMin) * time.Minute
	if rescanInterval == 0 {
		rescanInterval = 5 * time.Minute
	}

	ticker := time.NewTicker(rescanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.refreshRepositories()
		}
	}
}

func (t *Tracker) refreshRepositories() {
	newRepos := make(map[string]gitinfo.Info)

	for _, repoPath := range t.cfg.Git.Repositories {
		repo, err := gitinfo.Discover(repoPath)
		if err != nil {
			log.Printf("skip configured repository %s: %v", repoPath, err)
			continue
		}
		newRepos[repo.Path] = repo
	}

	t.scanRoots(newRepos)

	t.repoMu.Lock()
	t.repos = newRepos
	t.repoMu.Unlock()

	log.Printf("Repository scan complete: indexed %d repositories", len(newRepos))
}

func (t *Tracker) scanRoots(dest map[string]gitinfo.Info) {
	maxDepth := t.cfg.Git.MaxDepth
	for _, root := range t.cfg.Git.Roots {
		t.scanRoot(root, maxDepth, dest)
	}
}

type scanEntry struct {
	path  string
	depth int
}

func (t *Tracker) scanRoot(root string, maxDepth int, dest map[string]gitinfo.Info) {
	if root == "" {
		return
	}

	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		log.Printf("repository scan: skip root %s: %v", root, err)
		return
	}
	if !info.IsDir() {
		root = filepath.Dir(root)
	}

	queue := []scanEntry{{path: root, depth: 0}}

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		if _, seen := dest[entry.path]; seen {
			continue
		}

		if _, err := os.Stat(filepath.Join(entry.path, ".git")); err == nil {
			if repo, err := gitinfo.Discover(entry.path); err == nil {
				dest[repo.Path] = repo
			}
			continue
		}

		if maxDepth > 0 && entry.depth >= maxDepth {
			continue
		}

		children, err := os.ReadDir(entry.path)
		if err != nil {
			continue
		}

		for _, child := range children {
			if !child.IsDir() {
				continue
			}
			if child.Name() == ".git" {
				continue
			}
			if child.Type()&os.ModeSymlink != 0 {
				continue
			}
			childPath := filepath.Join(entry.path, child.Name())
			queue = append(queue, scanEntry{path: childPath, depth: entry.depth + 1})
		}
	}
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
		log.Printf("Session started repo=%s branch=%s source=%s", sess.Repo.Name, sess.Branch, evt.path)
		return
	}

	sess.Touch(branch, evt.when)
	log.Printf("Activity detected repo=%s branch=%s source=%s totalEvents=%d", sess.Repo.Name, sess.Branch, evt.path, sess.Events)
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
		log.Printf("Flushing idle session repo=%s duration=%s events=%d", sess.Repo.Name, sess.Duration(), sess.Events)
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
		log.Printf("Flushing remaining session repo=%s duration=%s events=%d", sess.Repo.Name, sess.Duration(), sess.Events)
		if err := t.publishSession(ctx, sess); err != nil {
			log.Printf("publish session %s: %v", sess.Repo.Path, err)
		}
	}
}

func (t *Tracker) publishSession(ctx context.Context, sess *session.State) error {
	if sess.Duration() <= 0 {
		return nil
	}

	// Bucket name format: user.repoName.branch
	bucketID := bucketIDForSession(sess.Repo.User, sess.Repo.Name, sess.Branch)

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

	log.Printf("Session published repo=%s branch=%s duration=%s events=%d bucket=%s", sess.Repo.Name, sess.Branch, sess.Duration(), sess.Events, bucketID)

	return nil
}

var bucketSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func bucketIDForSession(user, repoName, branch string) string {
	// Format: user.repoName.branch
	user = bucketSanitizer.ReplaceAllString(strings.ToLower(user), "-")
	user = strings.Trim(user, "-")
	if user == "" {
		user = "unknown"
	}

	repoName = bucketSanitizer.ReplaceAllString(strings.ToLower(repoName), "-")
	repoName = strings.Trim(repoName, "-")
	if repoName == "" {
		repoName = "unknown"
	}

	branch = bucketSanitizer.ReplaceAllString(strings.ToLower(branch), "-")
	branch = strings.Trim(branch, "-")
	if branch == "" {
		branch = "unknown"
	}

	return fmt.Sprintf("%s.%s.%s", user, repoName, branch)
}

type repoEvent struct {
	repo gitinfo.Info
	when time.Time
	path string
}
