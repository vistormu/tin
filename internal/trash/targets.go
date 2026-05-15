package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ExpandTargets(patterns []string) ([]string, error) {
	var targets []string
	seen := make(map[string]struct{})

	for _, pattern := range patterns {
		matches, err := expandPattern(pattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			clean := filepath.Clean(match)
			if _, ok := seen[clean]; ok {
				continue
			}
			seen[clean] = struct{}{}
			targets = append(targets, clean)
		}
	}

	return targets, nil
}

func expandPattern(pattern string) ([]string, error) {
	if hasGlobMeta(pattern) {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("pattern %q matched no files", pattern)
		}
		return matches, nil
	}

	if _, err := os.Lstat(pattern); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("target %q does not exist", pattern)
		}
		return nil, err
	}
	return []string{pattern}, nil
}

func hasGlobMeta(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}
