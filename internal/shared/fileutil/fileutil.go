package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Parent(path string) string {
	return filepath.Dir(path)
}

func Base(path string) string {
	return filepath.Base(path)
}

func Ext(path string) string {
	return filepath.Ext(path)
}

func Stem(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func Join(parts ...string) string {
	return filepath.Join(parts...)
}

func DefaultOutputPath(input string, suffix string) string {
	parent := Parent(input)
	stem := Stem(input)
	ext := Ext(input)
	return filepath.Join(parent, stem+suffix+ext)
}

func ReplaceExt(input string, newExt string) string {
	parent := Parent(input)
	stem := Stem(input)
	if newExt != "" && !strings.HasPrefix(newExt, ".") {
		newExt = "." + newExt
	}
	return filepath.Join(parent, stem+newExt)
}

func FileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("stat %q: %w", path, err)
	}
	if info.IsDir() {
		return 0, fmt.Errorf("%q is a directory", path)
	}
	return info.Size(), nil
}

func LowerExt(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

func IsSupportedImageExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
