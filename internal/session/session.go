package session

import (
	"time"

	"github.com/liamdn8/auto-worklog-agent/internal/gitinfo"
)

// State tracks the lifecycle of a work session.
type State struct {
	Repo         gitinfo.Info
	Branch       string
	Start        time.Time
	LastActivity time.Time
	Events       int
	StartCommit  string           // Commit hash at session start
	Commits      []gitinfo.Commit // All commits made during this session
	App          string           // Application name (IDE) where activity was detected
}

// NewState constructs a fresh session state.
func NewState(repo gitinfo.Info, branch string, ts time.Time, app string) *State {
	return &State{
		Repo:         repo,
		Branch:       branch,
		Start:        ts,
		LastActivity: ts,
		Events:       1,
		StartCommit:  "",
		Commits:      []gitinfo.Commit{},
		App:          app,
	}
}

// Touch updates the session's last activity timestamp.
func (s *State) Touch(branch string, app string, ts time.Time) {
	if branch != "" {
		s.Branch = branch
	}
	if app != "" {
		s.App = app
	}
	s.LastActivity = ts
	s.Events++
}

// Duration returns the elapsed active duration of the session.
func (s *State) Duration() time.Duration {
	return s.LastActivity.Sub(s.Start)
}
