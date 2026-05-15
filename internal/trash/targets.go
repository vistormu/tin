package trash

import "tin/internal/shared/fsx"

func ExpandTargets(patterns []string) ([]string, error) {
	return fsx.ExpandPaths(patterns)
}
