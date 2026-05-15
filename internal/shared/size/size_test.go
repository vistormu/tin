package size

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := map[int64]string{
		0:           "0 B",
		512:         "512 B",
		1024:        "1.0 KB",
		1536:        "1.5 KB",
		1024 * 1024: "1.0 MB",
	}

	for input, want := range tests {
		if got := Format(input); got != want {
			t.Fatalf("Format(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "b.txt"), []byte("defg"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := Path(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Fatalf("Path() = %d, want 7", got)
	}
}
