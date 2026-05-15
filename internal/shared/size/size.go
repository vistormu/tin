package size

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func Path(path string) (int64, error) {
	var total int64

	err := filepath.WalkDir(path, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		total += info.Size()
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("size %q: %w", path, err)
	}

	return total, nil
}

func Format(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	value := float64(bytes)
	for _, suffix := range []string{"KB", "MB", "GB", "TB", "PB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}

	return fmt.Sprintf("%.1f EB", value/unit)
}
