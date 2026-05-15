package trash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tin/internal/shared/confirm"
)

func TestNewMoverForOS(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	tests := []struct {
		goos    string
		wantErr bool
	}{
		{goos: "darwin"},
		{goos: "linux"},
		{goos: "windows", wantErr: true},
		{goos: "freebsd", wantErr: true},
	}

	for _, tt := range tests {
		_, err := NewMoverForOS(tt.goos)
		if tt.wantErr && err == nil {
			t.Fatalf("NewMoverForOS(%q) returned nil error", tt.goos)
		}
		if !tt.wantErr && err != nil {
			t.Fatalf("NewMoverForOS(%q) returned error: %v", tt.goos, err)
		}
	}
}

func TestExpandTargets(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.md")
	b := filepath.Join(dir, "b.md")
	c := filepath.Join(dir, "c.txt")
	for _, path := range []string{a, b, c} {
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	targets, err := ExpandTargets([]string{filepath.Join(dir, "*.md"), a})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{filepath.Clean(a), filepath.Clean(b)}
	if len(targets) != len(want) {
		t.Fatalf("ExpandTargets returned %d targets, want %d: %#v", len(targets), len(want), targets)
	}
	for i := range want {
		if targets[i] != want[i] {
			t.Fatalf("target %d = %q, want %q", i, targets[i], want[i])
		}
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name string
		yes  bool
		in   string
		want bool
	}{
		{name: "yes flag", yes: true, want: true},
		{name: "y answer", in: "y\n", want: true},
		{name: "yes answer", in: "YES\n", want: true},
		{name: "default no", in: "\n", want: false},
		{name: "other answer", in: "no\n", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := confirm.Ask(tt.yes, "Move 2 item(s) to trash?", strings.NewReader(tt.in), &strings.Builder{})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrashInfoContent(t *testing.T) {
	deletedAt := time.Date(2026, 5, 15, 10, 30, 45, 0, time.UTC)
	got := trashInfoContent("/home/me/file with spaces.txt", deletedAt)

	if !strings.Contains(got, "[Trash Info]\n") {
		t.Fatalf("trash info missing header: %q", got)
	}
	if !strings.Contains(got, "Path=/home/me/file%20with%20spaces.txt\n") {
		t.Fatalf("trash info path was not URL escaped correctly: %q", got)
	}
	if !strings.Contains(got, "DeletionDate=2026-05-15T10:30:45\n") {
		t.Fatalf("trash info deletion date missing: %q", got)
	}
}

func TestDarwinMoverMoveUsesTrashDirectory(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(source, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	trashDir := filepath.Join(dir, ".Trash")
	mover := darwinMover{trashDir: trashDir}
	if err := mover.Move(source); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists or unexpected error: %v", err)
	}

	trashed := filepath.Join(trashDir, "source.txt")
	if got, err := os.ReadFile(trashed); err != nil || string(got) != "hello" {
		t.Fatalf("trashed file = %q, %v; want hello", got, err)
	}
}

func TestDarwinMoverMoveUsesUniqueNameOnCollision(t *testing.T) {
	dir := t.TempDir()
	trashDir := filepath.Join(dir, ".Trash")
	if err := os.Mkdir(trashDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trashDir, "source.txt"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}

	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(source, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}

	mover := darwinMover{trashDir: trashDir}
	if err := mover.Move(source); err != nil {
		t.Fatal(err)
	}

	trashed := filepath.Join(trashDir, "source.txt.1")
	if got, err := os.ReadFile(trashed); err != nil || string(got) != "new" {
		t.Fatalf("collided trashed file = %q, %v; want new", got, err)
	}
}

func TestDarwinMoverMoveSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("target"), 0600); err != nil {
		t.Fatal(err)
	}

	source := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, source); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	trashDir := filepath.Join(dir, ".Trash")
	mover := darwinMover{trashDir: trashDir}
	if err := mover.Move(source); err != nil {
		t.Fatal(err)
	}

	trashed := filepath.Join(trashDir, "link.txt")
	gotTarget, err := os.Readlink(trashed)
	if err != nil {
		t.Fatal(err)
	}
	if gotTarget != target {
		t.Fatalf("trashed symlink target = %q, want %q", gotTarget, target)
	}
	if got, err := os.ReadFile(target); err != nil || string(got) != "target" {
		t.Fatalf("symlink target content = %q, %v; want target", got, err)
	}
}
