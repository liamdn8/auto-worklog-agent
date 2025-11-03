package gitinfo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	abs, err := filepath.Abs(path)
	if err != nil {
		return Info{}, fmt.Errorf("resolve path: %w", err)
	}

	if err := ensureGitRepo(abs); err != nil {
		return Info{}, err
	}

	branch, err := CurrentBranch(abs)
	if err != nil {
		return Info{}, err
	}

	user, _ := gitString(abs, "config", "--get", "user.name")
	email, _ := gitString(abs, "config", "--get", "user.email")
	remote, _ := gitString(abs, "config", "--get", "remote.origin.url")

	return Info{
		Path:   abs,
		Name:   filepath.Base(abs),
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
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}
	return nil
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
