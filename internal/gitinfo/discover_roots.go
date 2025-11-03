package gitinfo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverFromRoots walks the provided root directories and returns git repositories found
// within the optional maxDepth (depth == 1 means direct children). When maxDepth <= 0 the
// traversal is unbounded. Duplicate repositories are automatically deduplicated.
func DiscoverFromRoots(roots []string, maxDepth int) ([]Info, error) {
	seen := make(map[string]struct{})
	repos := make([]Info, 0)

	for _, root := range roots {
		if root == "" {
			root = "."
		}
		if strings.TrimSpace(root) == "" {
			continue
		}

		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, fmt.Errorf("resolve root %s: %w", root, err)
		}

		if statErr := ensureDir(absRoot); statErr != nil {
			// Skip roots that are missing instead of failing the whole discovery.
			if errors.Is(statErr, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("stat root %s: %w", absRoot, statErr)
		}

		walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if errors.Is(walkErr, fs.ErrPermission) {
					return filepath.SkipDir
				}
				return walkErr
			}

			if !d.IsDir() {
				return nil
			}

			if path != absRoot {
				if maxDepth > 0 && depthExceeded(absRoot, path, maxDepth) {
					return filepath.SkipDir
				}
				if shouldSkipDiscoveryDir(d.Name()) {
					return filepath.SkipDir
				}
			}

			if d.Name() == ".git" {
				repoPath := filepath.Dir(path)
				if _, ok := seen[repoPath]; ok {
					return filepath.SkipDir
				}

				info, err := Discover(repoPath)
				if err != nil {
					return filepath.SkipDir
				}

				repos = append(repos, info)
				seen[repoPath] = struct{}{}
				return filepath.SkipDir
			}

			return nil
		})

		if walkErr != nil {
			return nil, fmt.Errorf("walk root %s: %w", absRoot, walkErr)
		}
	}

	return repos, nil
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	return nil
}

func depthExceeded(root, current string, maxDepth int) bool {
	if maxDepth <= 0 || root == current {
		return false
	}

	rel, err := filepath.Rel(root, current)
	if err != nil || rel == "." {
		return false
	}

	rel = strings.Trim(rel, string(os.PathSeparator))
	if rel == "" {
		return false
	}

	depth := 1
	if strings.Contains(rel, string(os.PathSeparator)) {
		depth = len(strings.Split(rel, string(os.PathSeparator)))
	}

	return depth > maxDepth
}

func shouldSkipDiscoveryDir(name string) bool {
	switch name {
	case "node_modules", "vendor", "target", "build", ".terraform", ".cache", ".local", ".config", ".idea", ".vscode", ".venv", "venv", "tmp", "temp":
		return true
	default:
		return false
	}
}
