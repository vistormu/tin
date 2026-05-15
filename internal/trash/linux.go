package trash

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tin/internal/shared/fsx"
)

type linuxMover struct {
	trashDir string
	now      func() time.Time
}

func newLinuxMover() (Mover, error) {
	trashDir, err := linuxTrashDir()
	if err != nil {
		return nil, err
	}
	return linuxMover{
		trashDir: trashDir,
		now:      time.Now,
	}, nil
}

func linuxTrashDir() (string, error) {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "Trash"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("cannot determine home directory for XDG Trash")
	}
	return filepath.Join(home, ".local", "share", "Trash"), nil
}

func (m linuxMover) Move(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(abs); err != nil {
		return err
	}

	filesDir := filepath.Join(m.trashDir, "files")
	infoDir := filepath.Join(m.trashDir, "info")
	if err := os.MkdirAll(filesDir, 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(infoDir, 0700); err != nil {
		return err
	}

	destPath, infoPath, err := m.uniquePaths(abs)
	if err != nil {
		return err
	}

	if err := os.WriteFile(infoPath, []byte(trashInfoContent(abs, m.now())), 0600); err != nil {
		return err
	}

	if err := fsx.Move(abs, destPath); err != nil {
		_ = os.Remove(infoPath)
		return err
	}

	return nil
}

func (m linuxMover) uniquePaths(originalPath string) (string, string, error) {
	base := filepath.Base(originalPath)
	for i := 0; ; i++ {
		name := fsx.UniqueName(base, i)

		destPath := filepath.Join(m.trashDir, "files", name)
		infoPath := filepath.Join(m.trashDir, "info", name+".trashinfo")
		destExists, err := fsx.Exists(destPath)
		if err != nil {
			return "", "", err
		}
		infoExists, err := fsx.Exists(infoPath)
		if err != nil {
			return "", "", err
		}
		if !destExists && !infoExists {
			return destPath, infoPath, nil
		}
	}
}

func trashInfoContent(originalPath string, deletedAt time.Time) string {
	return fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		escapeTrashPath(originalPath),
		deletedAt.Format("2006-01-02T15:04:05"),
	)
}

func escapeTrashPath(path string) string {
	slashed := filepath.ToSlash(path)
	parts := strings.Split(slashed, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
