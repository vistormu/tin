package compress

import (
	"fmt"
	"strconv"

	"tin/internal/shared/helptext"
)

type intensityMapping struct {
	pdfPreset    string
	imageQuality int
}

var intensities = map[string]intensityMapping{
	"low":      {pdfPreset: "prepress", imageQuality: 90},
	"medium":   {pdfPreset: "printer", imageQuality: 80},
	"high":     {pdfPreset: "ebook", imageQuality: 65},
	"veryhigh": {pdfPreset: "screen", imageQuality: 45},
}

const defaultIntensity = "medium"

func resolveIntensity(intensity string) (intensityMapping, error) {
	if intensity == "" {
		intensity = defaultIntensity
	}
	mapping, ok := intensities[intensity]
	if !ok {
		return intensityMapping{}, fmt.Errorf("unsupported intensity %q; use low, medium, high, or veryhigh", intensity)
	}
	return mapping, nil
}

func pdfPresetForIntensity(intensity string) (string, error) {
	mapping, err := resolveIntensity(intensity)
	if err != nil {
		return "", err
	}
	return mapping.pdfPreset, nil
}

func imageQualityForIntensity(intensity string) (string, error) {
	mapping, err := resolveIntensity(intensity)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(mapping.imageQuality), nil
}

func IntensityHelpSections() []helptext.Section {
	return []helptext.Section{
		{
			Title: "Intensity",
			Items: []helptext.Item{
				{Value: "low", Description: "least compression (PDF prepress, image quality 90)"},
				{Value: "medium", Description: "balanced, default (PDF printer, image quality 80)"},
				{Value: "high", Description: "more compression (PDF ebook, image quality 65)"},
				{Value: "veryhigh", Description: "maximum compression (PDF screen, image quality 45)"},
			},
		},
	}
}
