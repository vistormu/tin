package convert

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultOutputPath(t *testing.T) {
	tests := []struct {
		input  string
		target string
		want   string
	}{
		{"photo.jpg", ".webp", "photo.webp"},
		{"archive/photo.png", ".webp", "archive/photo.webp"},
		{"report.final.jpeg", ".png", "report.final.png"},
		{"/tmp/image.PNG", ".jpg", "/tmp/image.jpg"},
	}

	for _, tt := range tests {
		if got := DefaultOutputPath(tt.input, tt.target); got != tt.want {
			t.Fatalf("DefaultOutputPath(%q, %q) = %q, want %q", tt.input, tt.target, got, tt.want)
		}
	}
}

func TestConvertRequiresInput(t *testing.T) {
	_, err := Convert(context.TODO(), Options{})
	if err == nil || !strings.Contains(err.Error(), "input file is required") {
		t.Fatalf("Convert() error = %v, want input-required validation", err)
	}
}

func TestConvertRejectsUnsupportedInput(t *testing.T) {
	_, err := Convert(context.TODO(), Options{Inputs: []string{"photo.gif"}, Target: "png"})
	if err == nil {
		t.Fatal("Convert() returned nil error for unsupported input")
	}
	if !strings.Contains(err.Error(), "unsupported input type") {
		t.Fatalf("Convert() error = %v, want unsupported-input", err)
	}
}

func TestConvertRequiresTargetOrOutput(t *testing.T) {
	_, err := Convert(context.TODO(), Options{Inputs: []string{"photo.png"}})
	if err == nil {
		t.Fatal("Convert() returned nil error when target/output unspecified")
	}
	if !strings.Contains(err.Error(), "target type is required") {
		t.Fatalf("Convert() error = %v, want target-required", err)
	}
}

func TestConvertRejectsUnsupportedTarget(t *testing.T) {
	_, err := Convert(context.TODO(), Options{Inputs: []string{"photo.png"}, Target: "gif"})
	if err == nil {
		t.Fatal("Convert() returned nil error for unsupported target")
	}
	if !strings.Contains(err.Error(), "unsupported target type") {
		t.Fatalf("Convert() error = %v, want unsupported-target", err)
	}
}

func TestConvertRejectsMismatchedTargetAndOutput(t *testing.T) {
	_, err := Convert(context.TODO(), Options{Inputs: []string{"photo.jpg"}, Target: "webp", Output: "photo.png"})
	if err == nil {
		t.Fatal("Convert() returned nil error for mismatched target/output")
	}
	if !strings.Contains(err.Error(), "does not match target") {
		t.Fatalf("Convert() error = %v, want target/output mismatch", err)
	}
}

func TestConvertRejectsSameFormat(t *testing.T) {
	_, err := Convert(context.TODO(), Options{Inputs: []string{"photo.jpg"}, Target: "jpeg"})
	if err == nil {
		t.Fatal("Convert() returned nil error for same input/target")
	}
	if !strings.Contains(err.Error(), "already") {
		t.Fatalf("Convert() error = %v, want already-this-format", err)
	}
}

func TestConvertRejectsCrossFamily(t *testing.T) {
	_, err := Convert(context.TODO(), Options{Inputs: []string{"clip.mp4"}, Target: "png"})
	if err == nil {
		t.Fatal("Convert() returned nil error for unsupported cross-family conversion")
	}
	if !strings.Contains(err.Error(), "unsupported conversion") {
		t.Fatalf("Convert() error = %v, want cross-family rejection", err)
	}
}

func TestConvertRejectsMixedFamilies(t *testing.T) {
	_, err := Convert(context.TODO(), Options{Inputs: []string{"a.png", "b.mp4"}, Target: "webp"})
	if err == nil {
		t.Fatal("Convert() returned nil error for mixed input families")
	}
	if !strings.Contains(err.Error(), "same family") {
		t.Fatalf("Convert() error = %v, want mixed-family rejection", err)
	}
}

func TestConvertRefusesPDFOutputFlag(t *testing.T) {
	_, err := Convert(context.TODO(), Options{Inputs: []string{"doc.pdf"}, Target: "png", Output: "out.png"})
	if err == nil {
		t.Fatal("Convert() returned nil error when -o passed for pdf->image")
	}
	if !strings.Contains(err.Error(), "not supported when converting pdf to images") {
		t.Fatalf("Convert() error = %v, want pdf->image -o rejection", err)
	}
}

func TestConvertRefusesToOverwriteExistingOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "photo.png")
	output := filepath.Join(dir, "photo.webp")
	if err := os.WriteFile(input, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("y"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Convert(context.TODO(), Options{Inputs: []string{input}, Output: output})
	if err == nil {
		t.Fatal("Convert() returned nil error when output exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Convert() error = %v, want exists check", err)
	}
}
