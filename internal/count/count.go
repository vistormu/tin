package count

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"tin/internal/find"
)

type Options struct {
	Paths      []string
	Extensions []string
}

type Result struct {
	Files int
	Lines int
	Bytes int64
}

func Count(opts Options) (Result, error) {
	exts := find.NormalizeExtensions(opts.Extensions)
	var result Result

	for _, root := range opts.Paths {
		info, err := os.Lstat(root)
		if err != nil {
			return result, err
		}

		if !info.IsDir() {
			if !info.Mode().IsRegular() {
				continue
			}
			if err := countOne(root, exts, &result); err != nil {
				return result, err
			}
			continue
		}

		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if !d.Type().IsRegular() {
				return nil
			}
			return countOne(path, exts, &result)
		})
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

func countOne(path string, exts []string, result *Result) error {
	if len(exts) > 0 && !matchesExt(path, exts) {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	lines, bytes, err := countReader(f)
	if err != nil {
		return err
	}
	result.Files++
	result.Lines += lines
	result.Bytes += bytes
	return nil
}

func matchesExt(path string, exts []string) bool {
	e := filepath.Ext(path)
	if e == "" {
		return false
	}
	target := lower(e)
	for _, want := range exts {
		if target == want {
			return true
		}
	}
	return false
}

func lower(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		out[i] = c
	}
	return string(out)
}

func countReader(r io.Reader) (int, int64, error) {
	br := bufio.NewReader(r)
	var lines int
	var bytes int64
	buf := make([]byte, 64*1024)
	for {
		n, err := br.Read(buf)
		bytes += int64(n)
		for i := 0; i < n; i++ {
			if buf[i] == '\n' {
				lines++
			}
		}
		if errors.Is(err, io.EOF) {
			return lines, bytes, nil
		}
		if err != nil {
			return lines, bytes, err
		}
	}
}
