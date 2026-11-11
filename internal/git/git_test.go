package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUndoSoftKeepsCommitChangesStaged(t *testing.T) {
	repo := newTestRepository(t)
	first := commitFile(t, repo, "note.txt", "first\n", "first")
	commitFile(t, repo, "note.txt", "second\n", "second")
	writeFile(t, filepath.Join(repo, "untracked.txt"), "keep\n")

	client := NewClient(repo)
	plan, err := client.BuildUndoPlan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Commit.Subject != "second" {
		t.Fatalf("subject = %q, want second", plan.Commit.Subject)
	}
	if plan.UntrackedFiles != 1 {
		t.Fatalf("untracked files = %d, want 1", plan.UntrackedFiles)
	}

	if err := client.Undo(context.Background(), plan, false); err != nil {
		t.Fatal(err)
	}

	if got := gitOutput(t, repo, "rev-parse", "HEAD"); got != first {
		t.Fatalf("HEAD = %q, want %q", got, first)
	}
	if got := gitOutput(t, repo, "diff", "--cached", "--", "note.txt"); !strings.Contains(got, "+second") {
		t.Fatalf("staged diff does not contain second commit changes:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(repo, "untracked.txt")); err != nil {
		t.Fatalf("untracked file was removed: %v", err)
	}
}

func TestUndoHardDiscardsTrackedChangesButKeepsUntrackedFiles(t *testing.T) {
	repo := newTestRepository(t)
	first := commitFile(t, repo, "note.txt", "first\n", "first")
	commitFile(t, repo, "note.txt", "second\n", "second")
	writeFile(t, filepath.Join(repo, "note.txt"), "unstaged\n")
	writeFile(t, filepath.Join(repo, "staged.txt"), "staged\n")
	runGit(t, repo, "add", "staged.txt")
	writeFile(t, filepath.Join(repo, "untracked.txt"), "keep\n")

	client := NewClient(repo)
	plan, err := client.BuildUndoPlan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if plan.TrackedChanges != 2 {
		t.Fatalf("tracked changes = %d, want 2", plan.TrackedChanges)
	}
	if plan.UntrackedFiles != 1 {
		t.Fatalf("untracked files = %d, want 1", plan.UntrackedFiles)
	}

	if err := client.Undo(context.Background(), plan, true); err != nil {
		t.Fatal(err)
	}

	if got := gitOutput(t, repo, "rev-parse", "HEAD"); got != first {
		t.Fatalf("HEAD = %q, want %q", got, first)
	}
	if got := readFile(t, filepath.Join(repo, "note.txt")); got != "first\n" {
		t.Fatalf("note.txt = %q, want first commit contents", got)
	}
	if _, err := os.Stat(filepath.Join(repo, "staged.txt")); !os.IsNotExist(err) {
		t.Fatalf("staged file still exists or stat failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "untracked.txt")); err != nil {
		t.Fatalf("untracked file was removed: %v", err)
	}
}

func TestUndoRejectsRootCommit(t *testing.T) {
	repo := newTestRepository(t)
	commitFile(t, repo, "note.txt", "first\n", "first")

	client := NewClient(repo)
	_, err := client.BuildUndoPlan(context.Background())
	if err == nil || !strings.Contains(err.Error(), "root commit") {
		t.Fatalf("error = %v, want root commit error", err)
	}
}

func TestUndoRejectsRepositoryWithoutCommits(t *testing.T) {
	repo := newTestRepository(t)

	client := NewClient(repo)
	_, err := client.BuildUndoPlan(context.Background())
	if err == nil || !strings.Contains(err.Error(), "no commits") {
		t.Fatalf("error = %v, want no commits error", err)
	}
}

func TestUndoRejectsChangedHead(t *testing.T) {
	repo := newTestRepository(t)
	commitFile(t, repo, "note.txt", "first\n", "first")
	commitFile(t, repo, "note.txt", "second\n", "second")

	client := NewClient(repo)
	plan, err := client.BuildUndoPlan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	commitFile(t, repo, "other.txt", "third\n", "third")

	err = client.Undo(context.Background(), plan, false)
	if err == nil || !strings.Contains(err.Error(), "HEAD changed") {
		t.Fatalf("error = %v, want changed HEAD error", err)
	}
	if got := gitOutput(t, repo, "show", "-s", "--format=%s", "HEAD"); got != "third" {
		t.Fatalf("HEAD subject = %q, want third", got)
	}
}

func TestBuildObliteratePlanNormalizesPathFromSubdirectory(t *testing.T) {
	repo := newTestRepository(t)
	commitFile(t, repo, "config/secret.txt", "secret\n", "add secret")
	commitFile(t, repo, "README.md", "hello\n", "add readme")
	subdir := filepath.Join(repo, "config")

	installFilterRepoShim(t)

	client := NewClient(subdir)
	plan, err := client.BuildObliteratePlan(context.Background(), "secret.txt")
	if err != nil {
		t.Fatal(err)
	}
	if plan.Path != "config/secret.txt" {
		t.Fatalf("path = %q, want config/secret.txt", plan.Path)
	}
	if plan.AffectedCommits != 1 {
		t.Fatalf("affected commits = %d, want 1", plan.AffectedCommits)
	}
}

func TestBuildObliteratePlanRequiresCleanWorkingTree(t *testing.T) {
	repo := newTestRepository(t)
	commitFile(t, repo, "secret.txt", "secret\n", "add secret")
	writeFile(t, filepath.Join(repo, "dirty.txt"), "dirty\n")
	installFilterRepoShim(t)

	client := NewClient(repo)
	_, err := client.BuildObliteratePlan(context.Background(), "secret.txt")
	if err == nil || !strings.Contains(err.Error(), "working tree must be clean") {
		t.Fatalf("error = %v, want clean working tree error", err)
	}
}

func TestObliteratePassesExactPathAndForceFlag(t *testing.T) {
	repo := newTestRepository(t)
	argsFile := installFilterRepoShim(t)
	t.Setenv("TIN_FILTER_REPO_ARGS", argsFile)

	client := NewClient(repo)
	plan := ObliteratePlan{Root: repo, Path: "config/secret.txt"}
	if _, err := client.Obliterate(context.Background(), plan, true); err != nil {
		t.Fatal(err)
	}

	got := readFile(t, argsFile)
	want := "--force\n--path\nconfig/secret.txt\n--invert-paths\n"
	if got != want {
		t.Fatalf("git-filter-repo args:\n%s\nwant:\n%s", got, want)
	}
}

func newTestRepository(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.name", "Tin Test")
	runGit(t, repo, "config", "user.email", "tin@example.test")
	runGit(t, repo, "config", "core.excludesFile", "/dev/null")
	return repo
}

func commitFile(t *testing.T, repo string, path string, contents string, message string) string {
	t.Helper()
	writeFile(t, filepath.Join(repo, filepath.FromSlash(path)), contents)
	runGit(t, repo, "add", "--", path)
	runGit(t, repo, "commit", "-q", "-m", message)
	return gitOutput(t, repo, "rev-parse", "HEAD")
}

func installFilterRepoShim(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	shim := filepath.Join(dir, "git-filter-repo")
	argsFile := filepath.Join(dir, "args.txt")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$TIN_FILTER_REPO_ARGS\"\n"
	if err := os.WriteFile(shim, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return argsFile
}

func runGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func gitOutput(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
