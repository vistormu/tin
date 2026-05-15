package flatten

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPlanMovesNestedFilesOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "root.txt"), []byte("root"), 0600); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "child.txt"), []byte("child"), 0600); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Moves) != 1 {
		t.Fatalf("BuildPlan() returned %d moves, want 1: %#v", len(plan.Moves), plan.Moves)
	}
	if len(plan.Dirs) != 2 {
		t.Fatalf("BuildPlan() returned %d dirs, want 2: %#v", len(plan.Dirs), plan.Dirs)
	}
	if want := filepath.Join(root, "child.txt"); plan.Moves[0].Destination != want {
		t.Fatalf("destination = %q, want %q", plan.Moves[0].Destination, want)
	}
}

func TestBuildPlanUsesExtensionPreservingCollisionNames(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "same.txt"), []byte("root"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "a"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "b"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "same.txt"), []byte("a"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b", "same.txt"), []byte("b"), 0600); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Moves) != 2 {
		t.Fatalf("BuildPlan() returned %d moves, want 2: %#v", len(plan.Moves), plan.Moves)
	}

	got := []string{filepath.Base(plan.Moves[0].Destination), filepath.Base(plan.Moves[1].Destination)}
	want := []string{"same.1.txt", "same.2.txt"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("destination %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestApplyMovesFilesAndRemovesEmptyDirectories(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "child.txt"), []byte("child"), 0600); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan(root)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Apply(plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.MovedFiles != 1 {
		t.Fatalf("MovedFiles = %d, want 1", result.MovedFiles)
	}
	if result.RemovedDirs != 2 {
		t.Fatalf("RemovedDirs = %d, want 2", result.RemovedDirs)
	}

	if got, err := os.ReadFile(filepath.Join(root, "child.txt")); err != nil || string(got) != "child" {
		t.Fatalf("flattened file = %q, %v; want child", got, err)
	}
	if _, err := os.Lstat(filepath.Join(root, "a")); !os.IsNotExist(err) {
		t.Fatalf("nested directory still exists or unexpected error: %v", err)
	}
}

func TestApplyRemovesEmptyDirectoriesWithoutMoves(t *testing.T) {
	root := t.TempDir()
	empty := filepath.Join(root, "empty", "nested")
	if err := os.MkdirAll(empty, 0700); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Moves) != 0 || len(plan.Dirs) != 2 {
		t.Fatalf("plan = %#v, want 0 moves and 2 dirs", plan)
	}

	result, err := Apply(plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.MovedFiles != 0 || result.RemovedDirs != 2 {
		t.Fatalf("result = %#v, want 0 moved and 2 removed", result)
	}
	if _, err := os.Lstat(filepath.Join(root, "empty")); !os.IsNotExist(err) {
		t.Fatalf("empty directory still exists or unexpected error: %v", err)
	}
}

func TestBuildPlanRejectsFiles(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := BuildPlan(file); err == nil {
		t.Fatal("BuildPlan() returned nil error for file root")
	}
}
