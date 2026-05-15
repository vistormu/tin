package fsx

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Exists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func UniquePath(path string) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	for i := 0; ; i++ {
		candidate := filepath.Join(dir, UniqueName(base, i))
		exists, err := Exists(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
}

func UniqueName(base string, index int) string {
	if index == 0 {
		return base
	}
	return fmt.Sprintf("%s.%d", base, index)
}

func Move(src string, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	if err := copyPath(src, dst); err != nil {
		_ = os.RemoveAll(dst)
		return err
	}
	if err := os.RemoveAll(src); err != nil {
		_ = os.RemoveAll(dst)
		return err
	}
	return nil
}

func copyPath(src string, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}

	if info.IsDir() {
		return copyDir(src, dst, info.Mode().Perm())
	}

	if info.Mode().IsRegular() {
		return copyFile(src, dst, info.Mode().Perm())
	}

	return fmt.Errorf("unsupported file type %q", info.Mode().String())
}

func copyDir(src string, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(dst, mode); err != nil {
		return err
	}

	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == src {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, target)
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if info.Mode().IsRegular() {
			return copyFile(path, target, info.Mode().Perm())
		}
		return fmt.Errorf("unsupported file type %q", info.Mode().String())
	})
}

func copyFile(src string, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
