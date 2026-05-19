package archive

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultZipOutput(t *testing.T) {
	tests := []struct {
		inputs []string
		want   string
	}{
		{inputs: []string{"docs"}, want: "docs.zip"},
		{inputs: []string{"photo.jpg"}, want: "photo.zip"},
		{inputs: []string{"nested/photo.jpg"}, want: "nested/photo.zip"},
		{inputs: []string{"a", "b"}, want: "archive.zip"},
	}

	for _, tt := range tests {
		if got := DefaultZipOutput(tt.inputs); got != tt.want {
			t.Fatalf("DefaultZipOutput(%v) = %q, want %q", tt.inputs, got, tt.want)
		}
	}
}

func TestDefaultUnzipOutput(t *testing.T) {
	tests := map[string]string{
		"archive.zip":        "archive",
		"nested/archive.zip": "nested/archive",
		"notzip":             "notzip",
	}

	for input, want := range tests {
		if got := DefaultUnzipOutput(input); got != want {
			t.Fatalf("DefaultUnzipOutput(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestZipAndUnzipRoundTrip(t *testing.T) {
	dir := t.TempDir()
	inputDir := filepath.Join(dir, "input")
	if err := os.Mkdir(inputDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "a.txt"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(inputDir, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "nested", "b.txt"), []byte("world"), 0600); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(dir, "out.zip")
	zipResult, err := Zip(ZipOptions{Inputs: []string{inputDir}, Output: archivePath})
	if err != nil {
		t.Fatal(err)
	}
	if zipResult.Files != 2 || zipResult.Bytes != 10 {
		t.Fatalf("Zip result = %#v, want 2 files and 10 bytes", zipResult)
	}

	outDir := filepath.Join(dir, "out")
	unzipResult, err := Unzip(UnzipOptions{Input: archivePath, OutputDir: outDir})
	if err != nil {
		t.Fatal(err)
	}
	if unzipResult.Files != 2 || unzipResult.Bytes != 10 {
		t.Fatalf("Unzip result = %#v, want 2 files and 10 bytes", unzipResult)
	}

	if got, err := os.ReadFile(filepath.Join(outDir, "input", "nested", "b.txt")); err != nil || string(got) != "world" {
		t.Fatalf("unzipped file = %q, %v; want world", got, err)
	}
}

func TestZipDoesNotOverwriteOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(input, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "out.zip")
	if err := os.WriteFile(output, []byte("exists"), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := Zip(ZipOptions{Inputs: []string{input}, Output: output}); err == nil {
		t.Fatal("Zip() returned nil error for existing output")
	}
}

func TestUnzipRejectsUnsafePaths(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "bad.zip")

	out, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(out)
	w, err := zw.Create("../evil.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("evil")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := Unzip(UnzipOptions{Input: archivePath, OutputDir: filepath.Join(dir, "out")}); err == nil {
		t.Fatal("Unzip() returned nil error for unsafe path")
	}
}

func TestUnzipDoesNotOverwriteFiles(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(input, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(dir, "out.zip")
	if _, err := Zip(ZipOptions{Inputs: []string{input}, Output: archivePath}); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(dir, "out")
	if err := os.Mkdir(outDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "a.txt"), []byte("exists"), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := Unzip(UnzipOptions{Input: archivePath, OutputDir: outDir}); err == nil {
		t.Fatal("Unzip() returned nil error for existing destination file")
	}
}
