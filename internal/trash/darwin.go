package trash

import (
	"fmt"
	"os"
	"path/filepath"

	"tin/internal/shared/fsx"
)

type darwinMover struct {
	trashDir string
}

func newDarwinMover() (Mover, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil, fmt.Errorf("cannot determine home directory for macOS Trash")
	}
	return darwinMover{trashDir: filepath.Join(home, ".Trash")}, nil
}

func (m darwinMover) Move(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(abs); err != nil {
		return err
	}
	if err := os.MkdirAll(m.trashDir, 0700); err != nil {
		return err
	}

	destPath, err := fsx.UniquePath(filepath.Join(m.trashDir, filepath.Base(abs)))
	if err != nil {
		return err
	}

	return fsx.Move(abs, destPath)
}
