package installhint

import (
	"errors"
	"testing"
)

func TestInstallHintUsesFirstDetectedManager(t *testing.T) {
	oldLookPath := lookPath
	t.Cleanup(func() {
		lookPath = oldLookPath
	})

	lookPath = func(name string) (string, error) {
		if name == "apt" {
			return "/usr/bin/apt", nil
		}
		return "", errors.New("not found")
	}

	got := InstallHint(PackageNames{"apt": "ghostscript"})
	want := "Install it with: sudo apt install ghostscript"
	if got != want {
		t.Fatalf("InstallHint() = %q, want %q", got, want)
	}
}

func TestInstallHintFallsBackWhenNoManagerIsDetected(t *testing.T) {
	oldLookPath := lookPath
	t.Cleanup(func() {
		lookPath = oldLookPath
	})

	lookPath = func(name string) (string, error) {
		return "", errors.New("not found")
	}

	got := InstallHint(PackageNames{"default": "ghostscript"})
	want := "Install it with your package manager."
	if got != want {
		t.Fatalf("InstallHint() = %q, want %q", got, want)
	}
}
