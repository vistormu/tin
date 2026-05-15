package helptext

import (
	"fmt"
	"strings"
)

type Section struct {
	Title string
	Items []Item
}

type Item struct {
	Value       string
	Description string
}

func RenderSections(sections []Section) string {
	var out strings.Builder
	for i, section := range sections {
		if i > 0 {
			out.WriteString("\n")
		}
		fmt.Fprintf(&out, "  %s:\n", section.Title)
		for _, item := range section.Items {
			fmt.Fprintf(&out, "    %-8s %s\n", item.Value, item.Description)
		}
	}
	return strings.TrimRight(out.String(), "\n")
}
