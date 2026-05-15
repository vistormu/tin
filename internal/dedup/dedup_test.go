package dedup

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func TestBuildPlanGroupsDuplicates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "hello")
	writeFile(t, filepath.Join(dir, "b.txt"), "hello")
	writeFile(t, filepath.Join(dir, "c.txt"), "different")
	writeFile(t, filepath.Join(dir, "sub/d.txt"), "hello")

	plan, err := BuildPlan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(plan.Groups))
	}
	g := plan.Groups[0]
	all := append([]string{g.Keep}, g.Trash...)
	sort.Strings(all)
	want := []string{
		filepath.Join(dir, "a.txt"),
		filepath.Join(dir, "b.txt"),
		filepath.Join(dir, "sub/d.txt"),
	}
	sort.Strings(want)
	if strings.Join(all, "|") != strings.Join(want, "|") {
		t.Fatalf("group members = %v, want %v", all, want)
	}
	if g.Keep != want[0] {
		t.Fatalf("Keep = %q, want alphabetically-first %q", g.Keep, want[0])
	}
}

func TestBuildPlanIgnoresUniqueSizes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "12345")
	writeFile(t, filepath.Join(dir, "b.txt"), "1234567")

	plan, err := BuildPlan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Groups) != 0 {
		t.Fatalf("groups = %d, want 0", len(plan.Groups))
	}
}

func TestBuildPlanRequiresDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.txt")
	writeFile(t, file, "hi")

	_, err := BuildPlan(file)
	if err == nil {
		t.Fatal("BuildPlan returned nil error for file path")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("error = %v, want not-a-directory", err)
	}
}

type recordingMover struct {
	moved []string
}

func (m *recordingMover) Move(path string) error {
	m.moved = append(m.moved, path)
	return os.Remove(path)
}

func TestApplyTrashesDuplicates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "same")
	writeFile(t, filepath.Join(dir, "b.txt"), "same")
	writeFile(t, filepath.Join(dir, "c.txt"), "same")

	plan, err := BuildPlan(dir)
	if err != nil {
		t.Fatal(err)
	}

	mover := &recordingMover{}
	result, err := Apply(plan, mover)
	if err != nil {
		t.Fatal(err)
	}
	if result.Trashed != 2 {
		t.Fatalf("Trashed = %d, want 2", result.Trashed)
	}

	keep := plan.Groups[0].Keep
	if _, err := os.Lstat(keep); err != nil {
		t.Fatalf("keep file missing after Apply: %v", err)
	}
	for _, trashed := range mover.moved {
		if _, err := os.Lstat(trashed); !os.IsNotExist(err) {
			t.Fatalf("trashed file still present: %s", trashed)
		}
	}
}

func TestPlanCounts(t *testing.T) {
	plan := Plan{Groups: []Group{
		{Keep: "a", Trash: []string{"b", "c"}, Size: 100},
		{Keep: "x", Trash: []string{"y"}, Size: 50},
	}}
	if got := plan.Removable(); got != 3 {
		t.Fatalf("Removable() = %d, want 3", got)
	}
	if got := plan.RecoverableBytes(); got != 250 {
		t.Fatalf("RecoverableBytes() = %d, want 250", got)
	}
}
