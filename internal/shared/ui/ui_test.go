package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestColorEnabledHonorsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CLICOLOR_FORCE", "1")

	if ColorEnabled(&bytes.Buffer{}) {
		t.Fatal("ColorEnabled returned true with NO_COLOR set")
	}
}

func TestColorEnabledHonorsCliColorForce(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("CLICOLOR", "")
	t.Setenv("CLICOLOR_FORCE", "1")

	if !ColorEnabled(&bytes.Buffer{}) {
		t.Fatal("ColorEnabled returned false with CLICOLOR_FORCE=1")
	}
}

func TestPrintSizeUsesColorWhenForced(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("CLICOLOR", "")
	t.Setenv("CLICOLOR_FORCE", "1")

	var out bytes.Buffer
	PrintSize(&out, "3.7 KB", "docs/specs.md")

	if got := out.String(); !strings.Contains(got, "\x1b[") {
		t.Fatalf("PrintSize did not emit ANSI with CLICOLOR_FORCE=1: %q", got)
	}
}

func TestPrintSizeIsPlainWhenColorIsDisabled(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var out bytes.Buffer
	PrintSize(&out, "3.7 KB", "docs/specs.md")

	got := out.String()
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("PrintSize emitted ANSI with NO_COLOR set: %q", got)
	}
	if got != "3.7 KB  docs/specs.md\n" {
		t.Fatalf("PrintSize() = %q", got)
	}
}

func TestPrintError(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var out bytes.Buffer
	PrintError(&out, errors.New("boom"))

	if got := out.String(); got != "error: boom\n" {
		t.Fatalf("PrintError() = %q", got)
	}
}
