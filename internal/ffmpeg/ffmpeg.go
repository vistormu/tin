package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"tin/internal/installhint"
)

const binary = "ffmpeg"

func Check() error {
	if _, err := exec.LookPath(binary); err != nil {
		return installhint.MissingToolError("ffmpeg", installhint.PackageNames{
			"default": "ffmpeg",
		})
	}
	return nil
}

func Convert(ctx context.Context, input string, output string) error {
	if err := Check(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binary, "-loglevel", "error", "-i", input, output)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func CompressVideo(ctx context.Context, input string, output string, crf string) error {
	if err := Check(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binary,
		"-loglevel", "error",
		"-y",
		"-i", input,
		"-c:v", "libx264",
		"-crf", crf,
		"-preset", "medium",
		"-c:a", "aac",
		"-b:a", "128k",
		output,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func StripVideo(ctx context.Context, input string, output string) error {
	if err := Check(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binary, "-loglevel", "error", "-i", input, "-vn", output)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
