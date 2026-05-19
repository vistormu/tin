package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tin/internal/resize"
	"tin/internal/shared/fsx"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewResizeCommand() *cobra.Command {
	var width int
	var height int
	var output string
	var inPlace bool

	cmd := &cobra.Command{
		Use:   "resize <file|glob>...",
		Short: "Resize an image file",
		Long:  "Resize one or more supported image files using ImageMagick. Accepts file paths and shell-style glob patterns.",
		Example: `  tin resize photo.jpg -w 1200
  tin resize photo.png -w 1200 -H 800 -o photo-small.png
  tin resize '*.jpg' -w 1200
  tin resize photo.jpg -w 1200 --in-place`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}
			if len(inputs) > 1 && output != "" {
				return fmt.Errorf("--output is only allowed with a single input")
			}
			if inPlace && output != "" {
				return fmt.Errorf("--in-place cannot be combined with --output")
			}

			out := cmd.OutOrStdout()
			for _, input := range inputs {
				result, err := resize.Resize(cmd.Context(), resize.Options{
					Input:   input,
					Output:  output,
					Width:   width,
					Height:  height,
					InPlace: inPlace,
				})
				if err != nil {
					return err
				}

				ui.PrintBox(out, "resized", []ui.Row{
					{Label: "before", Value: size.Format(result.BeforeSize)},
					{Label: "after", Value: size.Format(result.AfterSize)},
					{Label: "output", Value: result.Output},
				})
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&width, "width", "w", 0, "target width in pixels")
	cmd.Flags().IntVarP(&height, "height", "H", 0, "optional target height in pixels")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (single input only)")
	cmd.Flags().BoolVar(&inPlace, "in-place", false, "replace each input file with its resized output")

	return cmd
}
