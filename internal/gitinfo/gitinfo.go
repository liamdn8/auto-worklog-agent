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
