package magick

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"tin/internal/installhint"
)

const binary = "magick"

func Check() error {
	if _, err := exec.LookPath(binary); err != nil {
		return installhint.MissingToolError("magick", installhint.PackageNames{
			"brew":    "imagemagick",
			"apt":     "imagemagick",
			"dnf":     "ImageMagick",
			"pacman":  "imagemagick",
			"apk":     "imagemagick",
			"zypper":  "ImageMagick",
			"default": "imagemagick",
		})
	}
	return nil
}

func CompressImage(ctx context.Context, input string, output string, quality string) error {
	if err := Check(); err != nil {
		return err
	}
	if quality == "" {
		quality = "82"
	}

	cmd := exec.CommandContext(ctx, binary, input, "-strip", "-quality", quality, output)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("magick failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func ConvertImage(ctx context.Context, input string, output string) error {
	if err := Check(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binary, input, "-strip", output)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("magick failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func PDFToImages(ctx context.Context, input string, outputPattern string) error {
	if err := Check(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binary,
		"-density", "150",
		input,
		"-scene", "1",
		"-strip",
		outputPattern,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("magick failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func ImagesToPDF(ctx context.Context, inputs []string, output string) error {
	if err := Check(); err != nil {
		return err
	}

	args := append([]string{}, inputs...)
	args = append(args, "-strip", output)
	cmd := exec.CommandContext(ctx, binary, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("magick failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func ResizeImage(ctx context.Context, input string, output string, width int, height int) error {
	if err := Check(); err != nil {
		return err
	}

	geometry := fmt.Sprintf("%dx", width)
	if height > 0 {
		geometry = fmt.Sprintf("%dx%d", width, height)
	}

	cmd := exec.CommandContext(ctx, binary, input, "-resize", geometry, "-strip", output)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("magick failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
