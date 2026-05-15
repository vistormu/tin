package ghostscript

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"tin/internal/installhint"
)

const binary = "gs"

func Check() error {
	if _, err := exec.LookPath(binary); err != nil {
		return installhint.MissingToolError("ghostscript", installhint.PackageNames{
			"default": "ghostscript",
		})
	}
	return nil
}

func CompressPDF(ctx context.Context, input string, output string, quality string) error {
	if err := Check(); err != nil {
		return err
	}

	setting, err := pdfSetting(quality)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binary,
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=/"+setting,
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-sOutputFile="+output,
		input,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ghostscript failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func MergePDFs(ctx context.Context, inputs []string, output string) error {
	if err := Check(); err != nil {
		return err
	}

	args := []string{
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-sOutputFile=" + output,
	}
	args = append(args, inputs...)

	cmd := exec.CommandContext(ctx, binary, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ghostscript failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func SplitPDF(ctx context.Context, input string, outputPattern string) error {
	if err := Check(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binary,
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-dSAFER",
		"-sOutputFile="+outputPattern,
		input,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ghostscript failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func pdfSetting(quality string) (string, error) {
	if quality == "" {
		return "ebook", nil
	}

	switch quality {
	case "screen", "ebook", "printer", "prepress", "default":
		return quality, nil
	case "print":
		return "printer", nil
	default:
		return "", fmt.Errorf("unsupported PDF quality %q; use screen, ebook, printer, prepress, or default", quality)
	}
}
