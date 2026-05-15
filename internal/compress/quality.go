package compress

import "tin/internal/shared/helptext"

func QualityHelpSections() []helptext.Section {
	return []helptext.Section{
		{
			Title: "PDF",
			Items: []helptext.Item{
				{Value: "screen", Description: "smallest, lowest quality"},
				{Value: "ebook", Description: "default PDF preset"},
				{Value: "printer", Description: "higher quality, larger output"},
				{Value: "print", Description: "alias for printer"},
				{Value: "prepress", Description: "high quality, largest output"},
				{Value: "default", Description: "Ghostscript default preset"},
			},
		},
		{
			Title: "Images (.jpg, .jpeg, .png, .webp)",
			Items: []helptext.Item{
				{Value: "1-100", Description: "numeric quality, default 82"},
			},
		},
	}
}
