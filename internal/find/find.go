package find

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Type int

const (
	TypeFile Type = iota
	TypeDir
	TypeAll
)

func ParseType(s string) (Type, error) {
	switch strings.ToLower(s) {
	case "", "f", "file", "files":
		return TypeFile, nil
	case "d", "dir", "dirs", "directory":
		return TypeDir, nil
	case "a", "all":
		return TypeAll, nil
	default:
		return 0, fmt.Errorf("unsupported type %q; use f, d, or all", s)
	}
}

type Options struct {
	Root       string
	Extensions []string
	NameGlob   string
	Type       Type
}

type Match struct {
	Path  string
	IsDir bool
}

func Find(opts Options) ([]Match, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}

	exts := normalizeExtensions(opts.Extensions)

	var matches []Match
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		isDir := d.IsDir()

		switch opts.Type {
		case TypeFile:
			if isDir {
				return nil
			}
		case TypeDir:
			if !isDir {
				return nil
			}
		}

		if len(exts) > 0 && !isDir {
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

		matches = append(matches, Match{Path: path, IsDir: isDir})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return matches, nil
}

func NormalizeExtensions(exts []string) []string {
	return normalizeExtensions(exts)
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
