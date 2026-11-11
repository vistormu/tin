package replace

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"tin/internal/shared/fsx"
)

type Options struct {
	Root       string
	Old        string
	New        string
	Extensions []string
	NameGlob   string
}

type FileChange struct {
	Path  string
	Count int
}

type Plan struct {
	Old     string
	New     string
	Changes []FileChange
}

func (p Plan) FileCount() int {
	return len(p.Changes)
}

func (p Plan) TotalMatches() int {
	total := 0
	for _, c := range p.Changes {
		total += c.Count
	}
	return total
}

type Result struct {
	Files   int
	Matches int
}

func BuildPlan(opts Options) (Plan, error) {
	if opts.Old == "" {
		return Plan{}, fmt.Errorf("search string must not be empty")
	}

	root := opts.Root
	if root == "" {
		root = "."
	}

	exts := normalizeExtensions(opts.Extensions)

	plan := Plan{Old: opts.Old, New: opts.New}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" && path != root {
				return filepath.SkipDir
			}
			return nil
		}

		if len(exts) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			if !contains(exts, ext) {
				return nil
			}
		}

		if opts.NameGlob != "" {
			ok, err := filepath.Match(opts.NameGlob, filepath.Base(path))
			if err != nil {
				return fmt.Errorf("invalid name pattern %q: %w", opts.NameGlob, err)
			}
			if !ok {
				return nil
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if isBinary(content) {
			return nil
		}

		count := bytes.Count(content, []byte(opts.Old))
		if count == 0 {
			return nil
		}

		plan.Changes = append(plan.Changes, FileChange{Path: path, Count: count})
		return nil
	})
	if err != nil {
		return Plan{}, err
	}

	return plan, nil
}

func Apply(plan Plan) (Result, error) {
	result := Result{}
	for _, change := range plan.Changes {
		content, err := os.ReadFile(change.Path)
		if err != nil {
			return result, err
		}

		count := bytes.Count(content, []byte(plan.Old))
		if count == 0 {
			continue
		}

		info, err := os.Stat(change.Path)
		if err != nil {
			return result, err
		}
		mode := info.Mode().Perm()

		updated := bytes.ReplaceAll(content, []byte(plan.Old), []byte(plan.New))

		err = fsx.ReplaceInPlace(change.Path, func(tmp string) error {
			if err := os.WriteFile(tmp, updated, mode); err != nil {
				return err
			}
			return os.Chmod(tmp, mode)
		})
		if err != nil {
			return result, err
		}

		result.Files++
		result.Matches += count
	}

	return result, nil
}

func isBinary(content []byte) bool {
	const sniff = 8000
	if len(content) > sniff {
		content = content[:sniff]
	}
	return bytes.IndexByte(content, 0) >= 0
}

func normalizeExtensions(exts []string) []string {
	if len(exts) == 0 {
		return nil
	}
	out := make([]string, 0, len(exts))
	for _, e := range exts {
		e = strings.TrimSpace(strings.ToLower(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		out = append(out, e)
	}
	return out
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
