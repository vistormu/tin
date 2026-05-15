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
	for _, want := range []string{"compress", "convert", "flatten", "merge", "resize", "size", "split", "trash", "unzip", "zip"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help output missing %q:\n%s", want, got)
		}
	}
}

func TestCompressIntensitiesCommand(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"compress", "intensities"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"low", "medium", "high", "veryhigh"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compress intensities output missing %q:\n%s", want, got)
		}
	}
}

func TestCompressHelpShowsIntensityValues(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"compress", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"Values:", "low", "medium", "high", "veryhigh"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compress help output missing %q:\n%s", want, got)
		}
	}
}

func TestCommandArgumentValidation(t *testing.T) {
	tests := [][]string{
		{"compress"},
		{"convert"},
		{"flatten"},
		{"merge"},
		{"resize"},
		{"size"},
		{"split"},
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
