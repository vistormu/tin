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
	border  lipgloss.Style
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
		border: renderer.NewStyle().
			Foreground(lipgloss.Color("8")),
	}
}

func newRenderer(w io.Writer) *lipgloss.Renderer {
	return lipgloss.NewRenderer(w)
}

func ColorEnabled(w io.Writer) bool {
	return newRenderer(w).ColorProfile() != termenv.Ascii
}

func PrintCompressResult(w io.Writer, before string, after string, reduction float64, output string) {
	PrintBox(w, "compressed", []Row{
		{Label: "before", Value: before},
		{Label: "after", Value: after},
		{Label: "reduction", Value: fmt.Sprintf("%.1f%%", reduction)},
		{Label: "output", Value: output},
	})
}

func PrintSize(w io.Writer, formattedSize string, path string) {
	theme := newTheme(w)
	fmt.Fprintf(w, "%s  %s\n", theme.value.Render(formattedSize), theme.muted.Render(path))
}

func PrintTargets(w io.Writer, targets []string) {
	theme := newTheme(w)
	fmt.Fprintln(w, theme.title.Render(fmt.Sprintf("targets (%d)", len(targets))))
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
	fmt.Fprintf(w, "%s %s\n", theme.err.Render("error:"), err.Error())
}

func PrintNote(w io.Writer, message string) {
	fmt.Fprintln(w, newTheme(w).muted.Render(message))
}

func Plural(count int, singular string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	if singular == "directory" {
		return fmt.Sprintf("%d directories", count)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}

func PrintBox(w io.Writer, title string, rows []Row) {
	theme := newTheme(w)
	body := renderRows(theme, rows)
	lines := strings.Split(body, "\n")

	contentWidth := 0
	for _, line := range lines {
		if width := lipgloss.Width(line); width > contentWidth {
			contentWidth = width
		}
	}

	inner := contentWidth + 2

	var top string
	if title != "" {
		titleText := theme.title.Render(title)
		titleVisible := lipgloss.Width(titleText)
		dashes := inner - titleVisible - 3
		if dashes < 1 {
			dashes = 1
		}
		top = theme.border.Render("╭─ ") + titleText + theme.border.Render(" "+strings.Repeat("─", dashes)+"╮")
	} else {
		top = theme.border.Render("╭" + strings.Repeat("─", inner) + "╮")
	}
	bottom := theme.border.Render("╰" + strings.Repeat("─", inner) + "╯")

	fmt.Fprintln(w, top)
	for _, line := range lines {
		pad := contentWidth - lipgloss.Width(line)
		if pad < 0 {
			pad = 0
		}
		fmt.Fprintln(w, theme.border.Render("│")+" "+line+strings.Repeat(" ", pad)+" "+theme.border.Render("│"))
	}
	fmt.Fprintln(w, bottom)
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
