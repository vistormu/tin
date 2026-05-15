package rename

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestParseStyle(t *testing.T) {
	for _, s := range []string{"kebab", "snake", "camel", "pascal", "screaming", ""} {
		if _, err := ParseStyle(s); err != nil {
			t.Fatalf("ParseStyle(%q) returned error: %v", s, err)
		}
	}
	if got, _ := ParseStyle(""); got != DefaultStyle {
		t.Fatalf("ParseStyle(\"\") = %q, want %q", got, DefaultStyle)
	}
	if got, _ := ParseStyle("KEBAB"); got != StyleKebab {
		t.Fatalf("ParseStyle(\"KEBAB\") = %q, want %q", got, StyleKebab)
	}
	if _, err := ParseStyle("wat"); err == nil {
		t.Fatal("ParseStyle(\"wat\") returned nil error")
	}
}

func TestTokenize(t *testing.T) {
	tests := map[string][]string{
		"My Photo Album":   {"My", "Photo", "Album"},
		"my-photo_album":   {"my", "photo", "album"},
		"camelCaseWord":    {"camel", "Case", "Word"},
		"HTMLParserTest":   {"HTML", "Parser", "Test"},
		"already-fine":     {"already", "fine"},
		"weird!!!chars???": {"weird", "chars"},
		"한국어 파일":           {"한국어", "파일"},
		"Mixed 한글 + ASCII": {"Mixed", "한글", "ASCII"},
		"123 Numbers 456":  {"123", "Numbers", "456"},
		"":                 nil,
	}

	for input, want := range tests {
		got := tokenize(input)
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Fatalf("tokenize(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestNormalizeBasenameStyles(t *testing.T) {
	type want struct {
		kebab, snake, camel, pascal, screaming string
	}
	tests := map[string]want{
		"My Photo Album.JPG": {
			kebab: "my-photo-album.jpg", snake: "my_photo_album.jpg",
			camel: "myPhotoAlbum.jpg", pascal: "MyPhotoAlbum.jpg",
			screaming: "MY_PHOTO_ALBUM.jpg",
		},
		"camelCaseWord.md": {
			kebab: "camel-case-word.md", snake: "camel_case_word.md",
			camel: "camelCaseWord.md", pascal: "CamelCaseWord.md",
			screaming: "CAMEL_CASE_WORD.md",
		},
		"HTMLParser.go": {
			kebab: "html-parser.go", snake: "html_parser.go",
			camel: "htmlParser.go", pascal: "HtmlParser.go",
			screaming: "HTML_PARSER.go",
		},
		"한국어 파일.TXT": {
			kebab: "한국어-파일.txt", snake: "한국어_파일.txt",
			camel: "한국어파일.txt", pascal: "한국어파일.txt",
			screaming: "한국어_파일.txt",
		},
	}

	for input, w := range tests {
		check := func(style Style, want string) {
			got, err := NormalizeBasename(input, style)
			if err != nil {
				t.Fatalf("NormalizeBasename(%q, %s) error: %v", input, style, err)
			}
			if got != want {
				t.Fatalf("NormalizeBasename(%q, %s) = %q, want %q", input, style, got, want)
			}
		}
		check(StyleKebab, w.kebab)
		check(StyleSnake, w.snake)
		check(StyleCamel, w.camel)
		check(StylePascal, w.pascal)
		check(StyleScreaming, w.screaming)
	}
}

func TestNormalizeBasenamePreservesNFDCombiningMarks(t *testing.T) {
	// macOS APFS/HFS+ surfaces filenames in NFD form. A combining mark
	// must not be treated as a token separator.
	const nfd = "Información Importante.TXT"
	got, err := NormalizeBasename(nfd, StyleKebab)
	if err != nil {
		t.Fatal(err)
	}
	if want := "información-importante.txt"; got != want {
		t.Fatalf("NormalizeBasename(NFD) = %q, want %q", got, want)
	}
}

func TestNormalizeBasenamePreservesLeadingDots(t *testing.T) {
	got, err := NormalizeBasename(".config.JSON", StyleKebab)
	if err != nil {
		t.Fatal(err)
	}
	if want := ".config.json"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalizeBasenameRejectsEmptyStem(t *testing.T) {
	_, err := NormalizeBasename("$$$.txt", StyleKebab)
	if err == nil {
		t.Fatal("NormalizeBasename returned nil error for stem with no letters or digits")
	}
}

func TestBuildPlanSkipsAlreadyNormalized(t *testing.T) {
	dir := t.TempDir()
	already := filepath.Join(dir, "already-fine.txt")
	rough := filepath.Join(dir, "Rough Name.TXT")
	for _, path := range []string{already, rough} {
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	plan, err := BuildPlan([]string{already, rough}, StyleKebab)
	if err != nil {
		t.Fatal(err)
	}

	if len(plan.NoOps) != 1 || plan.NoOps[0] != already {
		t.Fatalf("NoOps = %v, want [%s]", plan.NoOps, already)
	}
	if len(plan.Moves) != 1 || plan.Moves[0].Source != rough {
		t.Fatalf("Moves = %v, want one rename of %s", plan.Moves, rough)
	}
	if got, want := plan.Moves[0].Destination, filepath.Join(dir, "rough-name.txt"); got != want {
		t.Fatalf("Destination = %q, want %q", got, want)
	}
}

func TestBuildPlanRespectsStyle(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "My Photo Album.JPG")
	if err := os.WriteFile(source, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan([]string{source}, StyleSnake)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Moves) != 1 {
		t.Fatalf("expected one move, got %v", plan.Moves)
	}
	if got, want := plan.Moves[0].Destination, filepath.Join(dir, "my_photo_album.jpg"); got != want {
		t.Fatalf("Destination = %q, want %q", got, want)
	}
}

func TestBuildPlanCollidesWithExistingFile(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "Weird!! Name.TXT")
	occupied := filepath.Join(dir, "weird-name.txt")
	for _, path := range []string{source, occupied} {
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	plan, err := BuildPlan([]string{source}, StyleKebab)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Moves) != 1 {
		t.Fatalf("expected one move, got %v", plan.Moves)
	}
	if got, want := plan.Moves[0].Destination, filepath.Join(dir, "weird-name.1.txt"); got != want {
		t.Fatalf("Destination = %q, want %q", got, want)
	}
}

func TestBuildPlanCollidesWithinPlan(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "Hello World.txt")
	b := filepath.Join(dir, "Hello_World.txt")
	for _, path := range []string{a, b} {
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	plan, err := BuildPlan([]string{a, b}, StyleKebab)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Moves) != 2 {
		t.Fatalf("expected two moves, got %v", plan.Moves)
	}

	dests := []string{plan.Moves[0].Destination, plan.Moves[1].Destination}
	sort.Strings(dests)
	want := []string{filepath.Join(dir, "hello-world.txt"), filepath.Join(dir, "hello-world.1.txt")}
	sort.Strings(want)
	if strings.Join(dests, "|") != strings.Join(want, "|") {
		t.Fatalf("destinations = %v, want %v", dests, want)
	}
}

func TestBuildPlanCaseOnlyRenameIsSingleMove(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "FILE.TXT")
	if err := os.WriteFile(source, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan([]string{source}, StyleKebab)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Moves) != 1 {
		t.Fatalf("expected one move, got %v", plan.Moves)
	}
	if got, want := plan.Moves[0].Destination, filepath.Join(dir, "file.txt"); got != want {
		t.Fatalf("Destination = %q, want %q", got, want)
	}
}

func TestApplyRenamesFiles(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "Hello World.TXT")
	if err := os.WriteFile(source, []byte("hi"), 0600); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan([]string{source}, StyleKebab)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Apply(plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.Moved != 1 {
		t.Fatalf("Moved = %d, want 1", result.Moved)
	}

	dest := filepath.Join(dir, "hello-world.txt")
	if got, err := os.ReadFile(dest); err != nil || string(got) != "hi" {
		t.Fatalf("renamed file = %q, %v; want hi", got, err)
	}
	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("original still exists or unexpected error: %v", err)
	}
}
