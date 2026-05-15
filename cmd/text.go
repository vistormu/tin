package cmd

import "fmt"

func plural(count int, singular string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	if singular == "directory" {
		return fmt.Sprintf("%d directories", count)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}
