package gitinfo

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Scanner discovers Git repositories within directory trees.
type Scanner struct {
	roots    []string
	maxDepth int
}

// NewScanner creates a repository scanner with the given roots and depth limit.
func NewScanner(roots []string, maxDepth int) *Scanner {
	return &Scanner{
		roots:    roots,
		maxDepth: maxDepth,
	}
}

// Scan walks the configured roots and returns all discovered Git repositories.
func (s *Scanner) Scan() ([]Info, error) {
	seen := make(map[string]bool)
	var repos []Info

	for _, root := range s.roots {
		abs, err := filepath.Abs(root)
		if err != nil {
			log.Printf("skip invalid root %s: %v", root, err)
			continue
		}

		if err := s.walk(abs, 0, seen, &repos); err != nil {
			return nil, fmt.Errorf("scan %s: %w", abs, err)
		}
	}

	return repos, nil
}

func (s *Scanner) walk(path string, depth int, seen map[string]bool, repos *[]Info) error {
	if s.maxDepth > 0 && depth > s.maxDepth {
		return nil
	}

	gitDir := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		if seen[path] {
			return nil
		}
		seen[path] = true

		info, err := Discover(path)
		if err != nil {
			log.Printf("discover repo %s: %v", path, err)
		} else {
			*repos = append(*repos, info)
		}
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsPermission(err) || os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if skipScanDir(entry.Name()) {
			continue
		}
		childPath := filepath.Join(path, entry.Name())
		if err := s.walk(childPath, depth+1, seen, repos); err != nil {
			return err
		}
	}

	return nil
}

func skipScanDir(name string) bool {
	if strings.HasPrefix(name, ".") && name != ".git" {
		return true
	}
	switch name {
	case "node_modules", "vendor", "target", "build", "dist", "venv", "__pycache__":
		return true
	}
	return false
}
