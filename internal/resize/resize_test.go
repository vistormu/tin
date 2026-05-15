package resize

import "testing"

func TestDefaultOutputPath(t *testing.T) {
	tests := map[string]string{
		"photo.jpg":          "photo-resized.jpg",
		"images/photo.jpeg":  "images/photo-resized.jpeg",
		"screen.final.webp":  "screen.final-resized.webp",
		"/tmp/image.PNG":     "/tmp/image-resized.PNG",
		"image-no-extension": "image-no-extension-resized",
	}

	for input, want := range tests {
		if got := DefaultOutputPath(input); got != want {
			t.Fatalf("DefaultOutputPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestResizeValidatesDimensionsBeforeStat(t *testing.T) {
	_, err := Resize(nil, Options{Input: "missing.jpg"})
	if err == nil || err.Error() != "width must be greater than 0" {
		t.Fatalf("Resize() error = %v, want width validation", err)
	}

	_, err = Resize(nil, Options{Input: "missing.jpg", Width: 10, Height: -1})
	if err == nil || err.Error() != "height must be greater than or equal to 0" {
		t.Fatalf("Resize() error = %v, want height validation", err)
	}
}

func TestResizeRejectsUnsupportedExtensionBeforeStat(t *testing.T) {
	_, err := Resize(nil, Options{Input: "missing.gif", Width: 10})
	if err == nil {
		t.Fatal("Resize() returned nil error")
	}
	if want := `unsupported file type ".gif"`; len(err.Error()) < len(want) || err.Error()[:len(want)] != want {
		t.Fatalf("Resize() error = %v, want unsupported extension", err)
	}
}
