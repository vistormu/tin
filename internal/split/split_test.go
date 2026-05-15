package split

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultOutputDir(t *testing.T) {
	tests := map[string]string{
		"report.pdf":               "report-pages",
		"docs/report.final.pdf":    "docs/report.final-pages",
		"/tmp/Quarterly Report.PDF": "/tmp/Quarterly Report-pages",
	}

	for input, want := range tests {
		if got := DefaultOutputDir(input); got != want {
			t.Fatalf("DefaultOutputDir(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSplitRequiresInput(t *testing.T) {
	_, err := Split(context.TODO(), Options{})
	if err == nil || !strings.Contains(err.Error(), "input file is required") {
		t.Fatalf("Split() error = %v, want input-required", err)
	}
}

func TestSplitRejectsNonPDF(t *testing.T) {
	_, err := Split(context.TODO(), Options{Input: "image.png"})
	if err == nil {
		t.Fatal("Split() returned nil error for non-pdf")
	}
	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("Split() error = %v, want unsupported-type", err)
	}
}

func TestSplitRefusesNonEmptyOutputDir(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(input, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "pages")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "stale.pdf"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Split(context.TODO(), Options{Input: input, OutputDir: outDir})
	if err == nil {
		t.Fatal("Split() returned nil error for non-empty output dir")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("Split() error = %v, want non-empty-dir", err)
	}
}

func TestSplitRejectsFileAsOutputDir(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(input, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(outDir, []byte("y"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Split(context.TODO(), Options{Input: input, OutputDir: outDir})
	if err == nil {
		t.Fatal("Split() returned nil error when output path is a file")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("Split() error = %v, want not-a-directory", err)
	}
}
