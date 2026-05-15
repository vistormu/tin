package flatten

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tin/internal/shared/fsx"
)

type Move struct {
	Source      string
	Destination string
}

type Plan struct {
	Root  string
	Moves []Move
	Dirs  []string
}

type Result struct {
	MovedFiles       int
	RemovedDirs      int
	SkippedRootFiles int
}

func BuildPlan(root string) (Plan, error) {
	if root == "" {
		return Plan{}, fmt.Errorf("root directory is required")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Plan{}, err
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return Plan{}, fmt.Errorf("stat %q: %w", root, err)
	}
	if !info.IsDir() {
		return Plan{}, fmt.Errorf("%q is not a directory", root)
	}

	var files []string
	var dirs []string
	if err := filepath.WalkDir(absRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == absRoot {
			return nil
		}
		if entry.IsDir() {
			dirs = append(dirs, path)
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("unsupported file type %q", path)
		}

		files = append(files, path)
		return nil
	}); err != nil {
		return Plan{}, err
	}
	sort.Strings(files)
	sort.Slice(dirs, func(i int, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	reserved := make(map[string]struct{})
	moves := make([]Move, 0, len(files))
	for _, path := range files {
		if filepath.Dir(path) == absRoot {
			reserved[path] = struct{}{}
			continue
		}

		dest := uniqueDestination(absRoot, filepath.Base(path), reserved)
		reserved[dest] = struct{}{}
		moves = append(moves, Move{Source: path, Destination: dest})
	}

	return Plan{Root: absRoot, Moves: moves, Dirs: dirs}, nil
}

func Apply(plan Plan) (Result, error) {
	result := Result{}
	for _, move := range plan.Moves {
		if err := fsx.Move(move.Source, move.Destination); err != nil {
			return result, fmt.Errorf("move %q to %q: %w", move.Source, move.Destination, err)
		}
		result.MovedFiles++
	}

	removed, err := removeEmptyDirs(plan.Dirs)
	if err != nil {
		return result, err
	}
	result.RemovedDirs = removed
	return result, nil
}

func uniqueDestination(root string, base string, reserved map[string]struct{}) string {
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	if stem == "" {
		stem = base
		ext = ""
	}

	for i := 0; ; i++ {
		name := base
		if i > 0 {
			name = fmt.Sprintf("%s.%d%s", stem, i, ext)
		}
		candidate := filepath.Join(root, name)
		if _, ok := reserved[candidate]; ok {
			continue
		}
		if exists, err := fsx.Exists(candidate); err == nil && exists {
			continue
		}
		return candidate
	}
}

func removeEmptyDirs(dirs []string) (int, error) {
	removed := 0
	for _, dir := range dirs {
		if err := os.Remove(dir); err == nil {
			removed++
		} else if !isNotEmpty(err) {
			return removed, fmt.Errorf("remove directory %q: %w", dir, err)
		}
	}
	return removed, nil
}

func isNotEmpty(err error) bool {
	return strings.Contains(err.Error(), "directory not empty")
}
