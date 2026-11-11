package poppler

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"tin/internal/installhint"
)

const (
	convertBinary = "pdftocairo"
	infoBinary    = "pdfinfo"
)

var packages = installhint.PackageNames{
	"brew":    "poppler",
	"apt":     "poppler-utils",
	"dnf":     "poppler-utils",
	"pacman":  "poppler",
	"apk":     "poppler-utils",
	"zypper":  "poppler-tools",
	"default": "poppler-utils",
}

func Check() error {
	for _, binary := range []string{convertBinary, infoBinary} {
		if _, err := exec.LookPath(binary); err != nil {
			return installhint.MissingToolError(binary, packages)
		}
	}
	return nil
}

func PageCount(ctx context.Context, input string) (int, error) {
	if err := Check(); err != nil {
		return 0, err
	}

	cmd := exec.CommandContext(ctx, infoBinary, input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("pdfinfo failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	for _, line := range strings.Split(string(out), "\n") {
		value, ok := strings.CutPrefix(line, "Pages:")
		if !ok {
			continue
		}
		pages, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, fmt.Errorf("could not read page count of %q: %w", input, err)
		}
		return pages, nil
	}
	return 0, fmt.Errorf("could not read page count of %q", input)
}

// PDFToSVG renders a single page. pdftocairo writes every page into one SVG
// when given a range, so pages must be converted one at a time.
func PDFToSVG(ctx context.Context, input string, page int, output string) error {
	if err := Check(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, convertBinary,
		"-svg",
		"-f", strconv.Itoa(page),
		"-l", strconv.Itoa(page),
		input,
		output,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pdftocairo failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
