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
}

// NewState constructs a fresh session state.
func NewState(repo gitinfo.Info, branch string, ts time.Time) *State {
	return &State{
		Repo:         repo,
		Branch:       branch,
		Start:        ts,
		LastActivity: ts,
		Events:       1,
	}
}

// Touch updates the session's last activity timestamp.
func (s *State) Touch(branch string, ts time.Time) {
	if branch != "" {
		s.Branch = branch
	}
	s.LastActivity = ts
	s.Events++
}

// Duration returns the elapsed active duration of the session.
func (s *State) Duration() time.Duration {
	return s.LastActivity.Sub(s.Start)
}
