package gitinfo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Info captures core git metadata for a repository.
type Info struct {
	Path   string
	Name   string
	Branch string
	User   string
	Email  string
	Remote string
}

// Discover collects git metadata for the provided repository root.
func Discover(path string) (Info, error) {
	root, err := FindRepoRoot(path)
	if err != nil {
		return Info{}, err
	}

	branch, err := CurrentBranch(root)
	if err != nil {
		return Info{}, err
	}

	user, _ := gitString(root, "config", "--get", "user.name")
	email, _ := gitString(root, "config", "--get", "user.email")
	remote, _ := gitString(root, "config", "--get", "remote.origin.url")

	return Info{
		Path:   root,
		Name:   filepath.Base(root),
		Branch: branch,
		User:   user,
		Email:  email,
		Remote: remote,
	}, nil
}

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch(path string) (string, error) {
	branch, err := gitString(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git branch: %w", err)
	}
	return branch, nil
}

func ensureGitRepo(path string) error {
	_, err := FindRepoRoot(path)
	return err
}

// FindRepoRoot walks up from the provided path until it finds a directory containing a .git folder.
func FindRepoRoot(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat path: %w", err)
	}

	if !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	for {
		if _, err := os.Stat(filepath.Join(abs, ".git")); err == nil {
			return abs, nil
		}

		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}

	return "", fmt.Errorf("not a git repository: %s", path)
}

func gitString(path string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", path}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Commit represents a single git commit with metadata.
type Commit struct {
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// GetCommitsSince retrieves all commits from startHash to HEAD.
// If startHash is empty, returns only the HEAD commit.
// Returns commits in chronological order (oldest first).
func GetCommitsSince(repoPath string, startHash string) ([]Commit, error) {
	var gitRange string
	if startHash == "" {
		gitRange = "HEAD"
	} else {
		gitRange = fmt.Sprintf("%s..HEAD", startHash)
	}

	// Format: hash|author|timestamp|message (one line per commit)
	// Use --reverse to get chronological order (oldest first)
	output, err := gitString(repoPath, "log", "--reverse", "--pretty=format:%H|%an <%ae>|%aI|%s", gitRange)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	if output == "" {
		return []Commit{}, nil
	}

	lines := strings.Split(output, "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}

		timestamp, err := time.Parse(time.RFC3339, parts[2])
		if err != nil {
			timestamp = time.Now()
		}

		commits = append(commits, Commit{
			Hash:      parts[0],
			Message:   parts[3],
			Author:    parts[1],
			Timestamp: timestamp,
		})
	}

	return commits, nil
}

// GetCurrentCommitHash returns the current HEAD commit hash.
func GetCurrentCommitHash(repoPath string) (string, error) {
	hash, err := gitString(repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("get HEAD hash: %w", err)
	}
	return hash, nil
}
