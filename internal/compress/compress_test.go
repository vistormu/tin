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

func TestCompressRejectsInPlaceWithOutput(t *testing.T) {
	_, err := Compress(nil, Options{Input: "x.pdf", InPlace: true, Output: "out.pdf"})
	if err == nil {
		t.Fatal("Compress() returned nil error when --in-place combined with --output")
	}
}

func TestResolveIntensity(t *testing.T) {
	cases := map[string]intensityMapping{
		"":         {pdfPreset: "printer", imageQuality: 80, videoCRF: 23},
		"low":      {pdfPreset: "prepress", imageQuality: 90, videoCRF: 20},
		"medium":   {pdfPreset: "printer", imageQuality: 80, videoCRF: 23},
		"high":     {pdfPreset: "ebook", imageQuality: 65, videoCRF: 28},
		"veryhigh": {pdfPreset: "screen", imageQuality: 45, videoCRF: 32},
	}

	for input, want := range cases {
		got, err := resolveIntensity(input)
		if err != nil {
			t.Fatalf("resolveIntensity(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Fatalf("resolveIntensity(%q) = %+v, want %+v", input, got, want)
		}
	}

	for _, bad := range []string{"max", "82", "ebook", "VERYHIGH"} {
		if _, err := resolveIntensity(bad); err == nil {
			t.Fatalf("resolveIntensity(%q) returned nil error", bad)
		}
	}
}
