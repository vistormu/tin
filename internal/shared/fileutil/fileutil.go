package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DefaultOutputPath(input string, suffix string) string {
	ext := filepath.Ext(input)
	stem := strings.TrimSuffix(input, ext)
	return stem + suffix + ext
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
