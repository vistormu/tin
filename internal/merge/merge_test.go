package merge

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeRequiresAtLeastTwoInputs(t *testing.T) {
	_, err := Merge(context.TODO(), Options{Inputs: []string{"a.pdf"}})
	if err == nil || !strings.Contains(err.Error(), "at least two") {
		t.Fatalf("Merge() error = %v, want at-least-two", err)
	}
}

func TestMergeRejectsNonPDFInput(t *testing.T) {
	_, err := Merge(context.TODO(), Options{Inputs: []string{"a.pdf", "b.txt"}})
	if err == nil {
		t.Fatal("Merge() returned nil error for non-pdf input")
	}
	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("Merge() error = %v, want unsupported-type", err)
	}
}

func TestMergeRejectsNonPDFOutput(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.pdf")
	b := filepath.Join(dir, "b.pdf")
	for _, path := range []string{a, b} {
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	_, err := Merge(context.TODO(), Options{Inputs: []string{a, b}, Output: filepath.Join(dir, "out.txt")})
	if err == nil {
		t.Fatal("Merge() returned nil error for non-pdf output")
	}
	if !strings.Contains(err.Error(), ".pdf extension") {
		t.Fatalf("Merge() error = %v, want pdf-extension", err)
	}
}

func TestMergeRefusesToOverwriteExistingOutput(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.pdf")
	b := filepath.Join(dir, "b.pdf")
	out := filepath.Join(dir, "out.pdf")
	for _, path := range []string{a, b, out} {
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	_, err := Merge(context.TODO(), Options{Inputs: []string{a, b}, Output: out})
	if err == nil {
		t.Fatal("Merge() returned nil error when output exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Merge() error = %v, want exists", err)
	}
}
