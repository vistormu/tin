package cwebp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"tin/internal/installhint"
)

const binary = "cwebp"

func Check() error {
	if _, err := exec.LookPath(binary); err != nil {
		return installhint.MissingToolError("cwebp", installhint.PackageNames{
			"brew":    "webp",
			"apt":     "webp",
			"dnf":     "libwebp-tools",
			"pacman":  "libwebp",
			"apk":     "libwebp-tools",
			"zypper":  "libwebp-tools",
			"default": "webp",
		})
	}
	return nil
}

func CompressWebP(ctx context.Context, input string, output string, quality string) error {
	if err := Check(); err != nil {
		return err
	}
	if quality == "" {
		quality = "82"
	}

	cmd := exec.CommandContext(ctx, binary, "-q", quality, input, "-o", output)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cwebp failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
