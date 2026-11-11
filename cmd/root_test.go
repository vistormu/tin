package cmd

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootHelpListsCommands(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"compress", "convert", "count", "dedup", "find", "flatten", "git", "merge", "rename", "resize", "size", "split", "trash", "unzip", "zip"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help output missing %q:\n%s", want, got)
		}
	}
}

func TestCompressIntensitiesCommand(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"compress", "intensities"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"low", "medium", "high", "veryhigh"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compress intensities output missing %q:\n%s", want, got)
		}
	}
}

func TestCompressHelpShowsIntensityValues(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"compress", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"Values:", "low", "medium", "high", "veryhigh"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compress help output missing %q:\n%s", want, got)
		}
	}
}

func TestCommandArgumentValidation(t *testing.T) {
	tests := [][]string{
		{"compress"},
		{"convert"},
		{"count"},
		{"dedup"},
		{"flatten"},
		{"git", "obliterate"},
		{"merge"},
		{"rename"},
		{"resize"},
		{"size"},
		{"split"},
		{"trash"},
		{"unzip"},
		{"zip"},
	}

	for _, args := range tests {
		root := NewRootCommand()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&out)
		root.SetArgs(args)

		if err := root.Execute(); err == nil {
			t.Fatalf("%v returned nil error", args)
		}
	}
}

func TestGitHelpListsSubcommands(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"git", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"undo", "obliterate"} {
		if !strings.Contains(got, want) {
			t.Fatalf("git help output missing %q:\n%s", want, got)
		}
	}
}

func TestExecutePrintsStandardizedError(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"size"})

	code := execute(root)
	if code != 1 {
		t.Fatalf("execute() exit code = %d, want 1", code)
	}

	got := out.String()
	if !strings.HasPrefix(got, "error: ") {
		t.Fatalf("execute() error output = %q, want error prefix", got)
	}
	if strings.Contains(got, "Usage:") {
		t.Fatalf("execute() printed usage for runtime error: %q", got)
	}
}

func TestUnzipExpandsGlobAndExtractsEachArchiveIntoItsOwnDir(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"one", "two"} {
		writeTestZip(t, filepath.Join(dir, name+".zip"), name+".txt", name)
	}

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"unzip", filepath.Join(dir, "*.zip")})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"one", "two"} {
		path := filepath.Join(dir, name, name+".txt")
		if got, err := os.ReadFile(path); err != nil || string(got) != name {
			t.Fatalf("%s = %q, %v; want %q", path, got, err, name)
		}
	}
}

func TestUnzipOutputFlagIsParentOfPerArchiveDirs(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"one", "two"} {
		writeTestZip(t, filepath.Join(dir, name+".zip"), name+".txt", name)
	}
	outDir := filepath.Join(dir, "extracted")

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"unzip", filepath.Join(dir, "*.zip"), "-o", outDir})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"one", "two"} {
		path := filepath.Join(outDir, name, name+".txt")
		if got, err := os.ReadFile(path); err != nil || string(got) != name {
			t.Fatalf("%s = %q, %v; want %q", path, got, err, name)
		}
	}
}

func writeTestZip(t *testing.T, archivePath string, entryName string, content string) {
	t.Helper()

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	entry, err := zw.Create(entryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestUnzipContinuesPastFailedArchives(t *testing.T) {
	dir := t.TempDir()
	writeTestZip(t, filepath.Join(dir, "good.zip"), "good.txt", "good")
	if err := os.WriteFile(filepath.Join(dir, "broken.zip"), []byte("not a zip"), 0600); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"unzip", filepath.Join(dir, "*.zip")})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() returned nil error despite a failed archive")
	}
	if !strings.Contains(err.Error(), "1 archive failed") {
		t.Fatalf("error = %v, want it to mention 1 archive failed", err)
	}
	if got := out.String(); !strings.Contains(got, "skipped") {
		t.Fatalf("output missing skip warning:\n%s", got)
	}

	path := filepath.Join(dir, "good", "good.txt")
	if got, err := os.ReadFile(path); err != nil || string(got) != "good" {
		t.Fatalf("%s = %q, %v; want good", path, got, err)
	}
}

func TestUnzipSingleArchiveReportsErrorVerbatim(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.zip"), []byte("not a zip"), 0600); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"unzip", filepath.Join(dir, "broken.zip")})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() returned nil error for a broken archive")
	}
	if strings.Contains(err.Error(), "failed to extract") {
		t.Fatalf("error = %v, want the underlying error, not a batch summary", err)
	}
}
