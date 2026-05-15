package fsx

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestExpandPathsExpandsGlobsAndKeepsLiterals(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.md")
	b := filepath.Join(dir, "b.md")
	c := filepath.Join(dir, "c.txt")
	for _, path := range []string{a, b, c} {
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := ExpandPaths([]string{filepath.Join(dir, "*.md"), c})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{filepath.Clean(a), filepath.Clean(b), filepath.Clean(c)}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("ExpandPaths = %v, want %v", got, want)
	}
}

func TestExpandPathsDeduplicates(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.md")
	if err := os.WriteFile(a, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := ExpandPaths([]string{a, a, filepath.Join(dir, "*.md")})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != filepath.Clean(a) {
		t.Fatalf("ExpandPaths = %v, want single %q", got, a)
	}
}

func TestExpandPathsErrorOnMissingLiteral(t *testing.T) {
	dir := t.TempDir()
	_, err := ExpandPaths([]string{filepath.Join(dir, "does-not-exist.txt")})
	if err == nil {
		t.Fatal("ExpandPaths returned nil error for missing literal path")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("ExpandPaths error = %v, want 'does not exist'", err)
	}
}

func TestReplaceInPlaceWritesNewContent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(target, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}

	err := ReplaceInPlace(target, func(tmp string) error {
		if tmp == target {
			t.Fatalf("tmp path equals target: %q", tmp)
		}
		return os.WriteFile(tmp, []byte("new"), 0600)
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("target content = %q, want %q", got, "new")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tin-tmp-") {
			t.Fatalf("leftover temp file: %s", e.Name())
		}
	}
}

func TestReplaceInPlaceLeavesTargetOnWriteFailure(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(target, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}

	err := ReplaceInPlace(target, func(tmp string) error {
		return os.ErrPermission
	})
	if err == nil {
		t.Fatal("ReplaceInPlace returned nil error when write failed")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original" {
		t.Fatalf("target content = %q, want %q (must not be clobbered)", got, "original")
	}
}

func TestExpandPathsErrorOnEmptyGlob(t *testing.T) {
	dir := t.TempDir()
	_, err := ExpandPaths([]string{filepath.Join(dir, "*.nope")})
	if err == nil {
		t.Fatal("ExpandPaths returned nil error for empty glob")
	}
	if !strings.Contains(err.Error(), "matched no files") {
		t.Fatalf("ExpandPaths error = %v, want 'matched no files'", err)
	}
}
