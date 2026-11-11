package replace

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestBuildPlanCountsMatches(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), "foo and foo")
	writeFile(t, filepath.Join(dir, "sub", "b.go"), "foo once")
	writeFile(t, filepath.Join(dir, "c.txt"), "foo ignored by ext filter")

	plan, err := BuildPlan(Options{Root: dir, Old: "foo", New: "bar", Extensions: []string{"go"}})
	if err != nil {
		t.Fatal(err)
	}

	if plan.FileCount() != 2 {
		t.Fatalf("FileCount = %d, want 2", plan.FileCount())
	}
	if plan.TotalMatches() != 3 {
		t.Fatalf("TotalMatches = %d, want 3", plan.TotalMatches())
	}
}

func TestApplyReplacesContent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a.go")
	writeFile(t, target, "example.com/old/foo")

	plan, err := BuildPlan(Options{
		Root: dir,
		Old:  "example.com/old",
		New:  "example.com/new",
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := Apply(plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.Files != 1 || result.Matches != 1 {
		t.Fatalf("result = %+v, want 1 file / 1 match", result)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "example.com/new/foo" {
		t.Fatalf("content = %q", string(got))
	}
}

func TestBuildPlanSkipsBinaryAndGit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "bin"), "foo\x00foo")
	writeFile(t, filepath.Join(dir, ".git", "config"), "foo in git")
	writeFile(t, filepath.Join(dir, "ok.txt"), "foo here")

	plan, err := BuildPlan(Options{Root: dir, Old: "foo", New: "bar"})
	if err != nil {
		t.Fatal(err)
	}

	if plan.FileCount() != 1 {
		t.Fatalf("FileCount = %d, want 1 (only ok.txt)", plan.FileCount())
	}
	if plan.Changes[0].Path != filepath.Join(dir, "ok.txt") {
		t.Fatalf("unexpected path %q", plan.Changes[0].Path)
	}
}

func TestBuildPlanEmptyOldErrors(t *testing.T) {
	_, err := BuildPlan(Options{Root: t.TempDir(), Old: "", New: "x"})
	if err == nil {
		t.Fatal("expected error for empty search string")
	}
}

func TestApplyPreservesMode(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "script.sh")
	writeFile(t, target, "foo")
	if err := os.Chmod(target, 0755); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan(Options{Root: dir, Old: "foo", New: "bar"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(plan); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Fatalf("mode = %v, want 0755", info.Mode().Perm())
	}
}
