package compress

import "testing"

func TestDefaultOutputPath(t *testing.T) {
	tests := map[string]string{
		"photo.jpg":              "photo-compressed.jpg",
		"archive/photo.jpeg":     "archive/photo-compressed.jpeg",
		"report.final.pdf":       "report.final-compressed.pdf",
		"no-extension":           "no-extension-compressed",
		"/tmp/screens/image.PNG": "/tmp/screens/image-compressed.PNG",
	}

	for input, want := range tests {
		if got := DefaultOutputPath(input); got != want {
			t.Fatalf("DefaultOutputPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestReductionPercent(t *testing.T) {
	got := (Result{BeforeSize: 100, AfterSize: 75}).ReductionPercent()
	if got != 25 {
		t.Fatalf("ReductionPercent() = %v, want 25", got)
	}

	got = (Result{BeforeSize: 0, AfterSize: 75}).ReductionPercent()
	if got != 0 {
		t.Fatalf("ReductionPercent() with zero before size = %v, want 0", got)
	}
}

func TestValidateImageQuality(t *testing.T) {
	for _, quality := range []string{"", "1", "82", "100"} {
		if err := validateImageQuality(quality); err != nil {
			t.Fatalf("validateImageQuality(%q) returned error: %v", quality, err)
		}
	}

	for _, quality := range []string{"0", "101", "ebook", "75.5"} {
		if err := validateImageQuality(quality); err == nil {
			t.Fatalf("validateImageQuality(%q) returned nil error", quality)
		}
	}
}
