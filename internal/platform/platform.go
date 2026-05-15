package platform

import (
	"fmt"
	"runtime"
)

func EnsureSupported() error {
	return EnsureSupportedOS(runtime.GOOS)
}

func EnsureSupportedOS(goos string) error {
	switch goos {
	case "darwin", "linux":
		return nil
	default:
		return fmt.Errorf("tin supports macOS and Linux only; current OS is %q", goos)
	}
}
