package opener

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"tin/internal/installhint"
)

func Open(ctx context.Context, target string) error {
	binary, err := binaryForOS(runtime.GOOS)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath(binary); err != nil {
		return missingToolError(binary)
	}

	cmd := exec.CommandContext(ctx, binary, target)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s failed: %w: %s", binary, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func binaryForOS(goos string) (string, error) {
	switch goos {
	case "darwin":
		return "open", nil
	case "linux":
		return "xdg-open", nil
	default:
		return "", fmt.Errorf("tin open supports macOS and Linux only; current OS is %q", goos)
	}
}

func missingToolError(binary string) error {
	if binary != "xdg-open" {
		return fmt.Errorf("%s is not installed", binary)
	}
	return installhint.MissingToolError("xdg-open", installhint.PackageNames{
		"apt":     "xdg-utils",
		"dnf":     "xdg-utils",
		"pacman":  "xdg-utils",
		"apk":     "xdg-utils",
		"zypper":  "xdg-utils",
		"default": "xdg-utils",
	})
}
