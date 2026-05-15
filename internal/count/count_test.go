package count

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCountSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	writeFile(t, path, "one\ntwo\nthree\n")

	got, err := Count(Options{Paths: []string{path}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Files != 1 || got.Lines != 3 || got.Bytes != 14 {
		t.Fatalf("got %+v, want Files=1 Lines=3 Bytes=14", got)
	}
}

func TestCountNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	writeFile(t, path, "hi")

	got, err := Count(Options{Paths: []string{path}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Lines != 0 {
		t.Fatalf("Lines = %d, want 0 (wc semantics)", got.Lines)
	}
	if got.Bytes != 2 {
		t.Fatalf("Bytes = %d, want 2", got.Bytes)
	}
}

func TestCountDirectoryRecursive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), "x\ny\n")
	writeFile(t, filepath.Join(dir, "b.md"), "hello\n")
	writeFile(t, filepath.Join(dir, "sub/c.go"), "z\n")

	got, err := Count(Options{Paths: []string{dir}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Files != 3 || got.Lines != 4 {
		t.Fatalf("got %+v, want Files=3 Lines=4", got)
	}
}

func TestCountExtensionFilter(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), "x\ny\n")
	writeFile(t, filepath.Join(dir, "b.md"), "hello\n")
	writeFile(t, filepath.Join(dir, "sub/c.go"), "z\n")
	writeFile(t, filepath.Join(dir, "sub/d.txt"), "ignore\n")

	got, err := Count(Options{Paths: []string{dir}, Extensions: []string{"go"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Files != 2 {
		t.Fatalf("Files = %d, want 2 (go only)", got.Files)
	}
	if got.Lines != 3 {
		t.Fatalf("Lines = %d, want 3", got.Lines)
	}
}

func TestCountEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	writeFile(t, path, "")

	got, err := Count(Options{Paths: []string{path}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Files != 1 || got.Lines != 0 || got.Bytes != 0 {
		t.Fatalf("got %+v, want Files=1 Lines=0 Bytes=0", got)
	}
}
