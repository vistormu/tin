package archive

import (
	"archive/zip"
	"io/fs"
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

func TestZipSkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	inputDir := filepath.Join(dir, "input")
	if err := os.Mkdir(inputDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "real.txt"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(inputDir, "real.txt"), filepath.Join(inputDir, "link.txt")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	if err := os.Symlink(filepath.Join(dir, "missing"), filepath.Join(inputDir, "broken.txt")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	archivePath := filepath.Join(dir, "out.zip")
	result, err := Zip(ZipOptions{Inputs: []string{inputDir}, Output: archivePath})
	if err != nil {
		t.Fatalf("Zip() returned error, want symlinks skipped: %v", err)
	}
	if result.Files != 1 || result.Bytes != 5 {
		t.Fatalf("Zip result = %#v, want 1 file and 5 bytes", result)
	}
	if result.Skipped != 2 {
		t.Fatalf("Zip skipped = %d, want 2", result.Skipped)
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

func TestUnzipDecodesNonUTF8EntryNames(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "legacy.zip")

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(file)
	// 0x9e is CP437 '₧'; raw, it is invalid UTF-8 and macOS rejects it.
	entry, err := zw.CreateHeader(&zip.FileHeader{Name: "Large (2\x9e)/icon.png", Method: zip.Deflate})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("png")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(dir, "out")
	if _, err := Unzip(UnzipOptions{Input: archivePath, OutputDir: outDir}); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(outDir, "Large (2₧)", "icon.png")
	if got, err := os.ReadFile(want); err != nil || string(got) != "png" {
		t.Fatalf("%s = %q, %v; want png", want, got, err)
	}
}

func TestDecodeEntryNameLeavesValidUTF8Alone(t *testing.T) {
	for _, name := range []string{"plain/file.txt", "café/naïve.txt", "日本語/x.png"} {
		if got := decodeEntryName(name); got != name {
			t.Fatalf("decodeEntryName(%q) = %q, want unchanged", name, got)
		}
	}
}

func TestUnzipExtractsIntoReadOnlyDirectoryEntries(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "dos.zip")

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(file)
	// Zips written on Windows store the DOS read-only attribute as mode 0555.
	dirHeader := &zip.FileHeader{Name: "Textures/"}
	dirHeader.SetMode(fs.ModeDir | 0555)
	if _, err := zw.CreateHeader(dirHeader); err != nil {
		t.Fatal(err)
	}
	entry, err := zw.CreateHeader(&zip.FileHeader{Name: "Textures/button.png", Method: zip.Deflate})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("png")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(dir, "out")
	if _, err := Unzip(UnzipOptions{Input: archivePath, OutputDir: outDir}); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(outDir, "Textures", "button.png")
	if got, err := os.ReadFile(want); err != nil || string(got) != "png" {
		t.Fatalf("%s = %q, %v; want png", want, got, err)
	}
}
