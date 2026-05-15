package find

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func setupTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := []string{
		"a.go",
		"b.md",
		"README",
		"sub/c.go",
		"sub/d.txt",
		"sub/nested/e.GO",
	}
	for _, rel := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func paths(matches []Match) []string {
	out := make([]string, len(matches))
	for i, m := range matches {
		out[i] = m.Path
	}
	sort.Strings(out)
	return out
}

func TestFindByExtension(t *testing.T) {
	dir := setupTree(t)
	got, err := Find(Options{Root: dir, Extensions: []string{"go"}, Type: TypeFile})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(dir, "a.go"),
		filepath.Join(dir, "sub/c.go"),
		filepath.Join(dir, "sub/nested/e.GO"),
	}
	sort.Strings(want)
	if strings.Join(paths(got), "|") != strings.Join(want, "|") {
		t.Fatalf("paths = %v, want %v", paths(got), want)
	}
}

func TestFindByMultipleExtensions(t *testing.T) {
	dir := setupTree(t)
	got, err := Find(Options{Root: dir, Extensions: []string{".go", ".md"}, Type: TypeFile})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d matches, want 4: %v", len(got), paths(got))
	}
}

func TestFindByNameGlob(t *testing.T) {
	dir := setupTree(t)
	got, err := Find(Options{Root: dir, NameGlob: "*.txt", Type: TypeFile})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || filepath.Base(got[0].Path) != "d.txt" {
		t.Fatalf("got %v, want one match d.txt", paths(got))
	}
}

func TestFindDirectoriesOnly(t *testing.T) {
	dir := setupTree(t)
	got, err := Find(Options{Root: dir, Type: TypeDir})
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range got {
		if !m.IsDir {
			t.Fatalf("non-dir match returned: %v", m)
		}
	}
	if len(got) != 2 {
		t.Fatalf("got %d directories, want 2: %v", len(got), paths(got))
	}
}

func TestFindAllTypes(t *testing.T) {
	dir := setupTree(t)
	got, err := Find(Options{Root: dir, Type: TypeAll})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 8 {
		t.Fatalf("got %d matches, want 8: %v", len(got), paths(got))
	}
}

func TestParseType(t *testing.T) {
	tests := map[string]Type{
		"":    TypeFile,
		"f":   TypeFile,
		"d":   TypeDir,
		"all": TypeAll,
	}
	for in, want := range tests {
		got, err := ParseType(in)
		if err != nil {
			t.Fatalf("ParseType(%q) error: %v", in, err)
		}
		if got != want {
			t.Fatalf("ParseType(%q) = %v, want %v", in, got, want)
		}
	}
	if _, err := ParseType("wat"); err == nil {
		t.Fatal("ParseType(wat) returned nil error")
	}
}

func TestNormalizeExtensions(t *testing.T) {
	got := NormalizeExtensions([]string{"go", ".MD", " txt ", ""})
	want := []string{".go", ".md", ".txt"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %v, want %v", got, want)
	}
}
