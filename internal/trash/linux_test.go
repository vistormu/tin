package trash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLinuxMoverMoveCreatesXDGTrashEntries(t *testing.T) {
	dir := t.TempDir()
	trashDir := filepath.Join(dir, "data", "Trash")
	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(source, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	mover := linuxMover{
		trashDir: trashDir,
		now: func() time.Time {
			return time.Date(2026, 5, 15, 10, 30, 45, 0, time.UTC)
		},
	}

	if err := mover.Move(source); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists or unexpected error: %v", err)
	}

	trashed := filepath.Join(trashDir, "files", "source.txt")
	if got, err := os.ReadFile(trashed); err != nil || string(got) != "hello" {
		t.Fatalf("trashed file = %q, %v; want hello", got, err)
	}

	infoPath := filepath.Join(trashDir, "info", "source.txt.trashinfo")
	info, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatal(err)
	}
	if want := "DeletionDate=2026-05-15T10:30:45"; !contains(string(info), want) {
		t.Fatalf("trash info %q missing %q", info, want)
	}
}

func TestLinuxMoverMoveUsesUniqueNameOnCollision(t *testing.T) {
	dir := t.TempDir()
	trashDir := filepath.Join(dir, "data", "Trash")
	if err := os.MkdirAll(filepath.Join(trashDir, "files"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(trashDir, "info"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trashDir, "files", "source.txt"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trashDir, "info", "source.txt.trashinfo"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}

	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(source, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}

	mover := linuxMover{
		trashDir: trashDir,
		now: func() time.Time {
			return time.Date(2026, 5, 15, 10, 30, 45, 0, time.UTC)
		},
	}
	if err := mover.Move(source); err != nil {
		t.Fatal(err)
	}

	trashed := filepath.Join(trashDir, "files", "source.txt.1")
	if got, err := os.ReadFile(trashed); err != nil || string(got) != "new" {
		t.Fatalf("collided trashed file = %q, %v; want new", got, err)
	}
	if _, err := os.Stat(filepath.Join(trashDir, "info", "source.txt.1.trashinfo")); err != nil {
		t.Fatal(err)
	}
}

func TestLinuxMoverMoveDirectory(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "folder")
	if err := os.Mkdir(source, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "child.txt"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	trashDir := filepath.Join(dir, "data", "Trash")
	mover := linuxMover{trashDir: trashDir, now: time.Now}
	if err := mover.Move(source); err != nil {
		t.Fatal(err)
	}

	child := filepath.Join(trashDir, "files", "folder", "child.txt")
	if got, err := os.ReadFile(child); err != nil || string(got) != "hello" {
		t.Fatalf("trashed child = %q, %v; want hello", got, err)
	}
	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("source directory still exists or unexpected error: %v", err)
	}
}

func contains(s string, sub string) bool {
	return strings.Contains(s, sub)
}
