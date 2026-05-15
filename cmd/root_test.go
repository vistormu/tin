package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpListsCommands(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"compress", "flatten", "resize", "size", "trash", "unzip", "zip"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help output missing %q:\n%s", want, got)
		}
	}
}

func TestCompressQualitiesCommand(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"compress", "qualities"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"screen", "ebook", "printer", "prepress", "1-100"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compress qualities output missing %q:\n%s", want, got)
		}
	}
}

func TestCompressHelpShowsQualityValues(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"compress", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"Quality Values:", "screen", "1-100"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compress help output missing %q:\n%s", want, got)
		}
	}
}

func TestCommandArgumentValidation(t *testing.T) {
	tests := [][]string{
		{"compress"},
		{"flatten"},
		{"resize"},
		{"size"},
		{"trash"},
		{"unzip"},
		{"zip"},
	}

	for _, args := range tests {
		root := NewRootCommand()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&out)
		root.SetArgs(args)

		if err := root.Execute(); err == nil {
			t.Fatalf("%v returned nil error", args)
		}
	}
}

func TestExecutePrintsStandardizedError(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"size"})

	code := execute(root)
	if code != 1 {
		t.Fatalf("execute() exit code = %d, want 1", code)
	}

	got := out.String()
	if !strings.HasPrefix(got, "Error: ") {
		t.Fatalf("execute() error output = %q, want Error prefix", got)
	}
	if strings.Contains(got, "Usage:") {
		t.Fatalf("execute() printed usage for runtime error: %q", got)
	}
}
