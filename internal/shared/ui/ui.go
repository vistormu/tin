package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type Row struct {
	Label string
	Value string
}

type theme struct {
	title   lipgloss.Style
	muted   lipgloss.Style
	label   lipgloss.Style
	value   lipgloss.Style
	err     lipgloss.Style
	success lipgloss.Style
	warning lipgloss.Style
	box     lipgloss.Style
}

func newTheme(w io.Writer) theme {
	renderer := newRenderer(w)
	return theme{
		title: renderer.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")),
		muted: renderer.NewStyle().
			Foreground(lipgloss.Color("8")),
		label: renderer.NewStyle().
			Foreground(lipgloss.Color("14")),
		value: renderer.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")),
		err: renderer.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9")),
		success: renderer.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")),
		warning: renderer.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11")),
		box: renderer.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1),
	}
}

func newRenderer(w io.Writer) *lipgloss.Renderer {
	return lipgloss.NewRenderer(w)
}

func ColorEnabled(w io.Writer) bool {
	return newRenderer(w).ColorProfile() != termenv.Ascii
}

func PrintCompressResult(w io.Writer, before string, after string, reduction float64, output string) {
	PrintBox(w, "Compressed", []Row{
		{Label: "Before", Value: before},
		{Label: "After", Value: after},
		{Label: "Reduction", Value: fmt.Sprintf("%.1f%%", reduction)},
		{Label: "Output", Value: output},
	})
}

func PrintSize(w io.Writer, formattedSize string, path string) {
	theme := newTheme(w)
	fmt.Fprintf(w, "%s  %s\n", theme.value.Render(formattedSize), theme.muted.Render(path))
}

func PrintTargets(w io.Writer, targets []string) {
	theme := newTheme(w)
	fmt.Fprintln(w, theme.title.Render(fmt.Sprintf("Targets (%d)", len(targets))))
	for _, target := range targets {
		fmt.Fprintf(w, "  %s %s\n", theme.muted.Render("-"), target)
	}
}

func PrintSuccess(w io.Writer, message string) {
	fmt.Fprintln(w, newTheme(w).success.Render(message))
}

func PrintWarning(w io.Writer, message string) {
	fmt.Fprintln(w, newTheme(w).warning.Render(message))
}

func PrintError(w io.Writer, err error) {
	if err == nil {
		return
	}
	theme := newTheme(w)
	fmt.Fprintf(w, "%s %s\n", theme.err.Render("Error:"), err.Error())
}

func PrintNote(w io.Writer, message string) {
	fmt.Fprintln(w, newTheme(w).muted.Render(message))
}

func PrintBox(w io.Writer, title string, rows []Row) {
	theme := newTheme(w)
	fmt.Fprintln(w, theme.title.Render(title))
	fmt.Fprintln(w, theme.box.Render(renderRows(theme, rows)))
}

func renderRows(theme theme, rows []Row) string {
	width := 0
	for _, row := range rows {
		if len(row.Label) > width {
			width = len(row.Label)
		}
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		label := fmt.Sprintf("%-*s", width, row.Label)
		lines = append(lines, fmt.Sprintf("%s  %s", theme.label.Render(label), theme.value.Render(row.Value)))
	}
	return strings.Join(lines, "\n")
}
