package cmd

import (
	"github.com/spf13/cobra"

	"tin/internal/resize"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewResizeCommand() *cobra.Command {
	var width int
	var height int
	var output string

	cmd := &cobra.Command{
		Use:   "resize <file>",
		Short: "Resize an image file",
		Long:  "Resize a supported image file using ImageMagick.",
		Example: `  tin resize photo.jpg -w 1200
  tin resize photo.png -w 1200 -H 800 -o photo-small.png`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := resize.Resize(cmd.Context(), resize.Options{
				Input:  args[0],
				Output: output,
				Width:  width,
				Height: height,
			})
			if err != nil {
				return err
			}

			ui.PrintBox(cmd.OutOrStdout(), "Resized", []ui.Row{
				{Label: "Before", Value: size.Format(result.BeforeSize)},
				{Label: "After", Value: size.Format(result.AfterSize)},
				{Label: "Output", Value: result.Output},
			})
			return nil
		},
	}

	cmd.Flags().IntVarP(&width, "width", "w", 0, "target width in pixels")
	cmd.Flags().IntVarP(&height, "height", "H", 0, "optional target height in pixels")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output path")

	return cmd
}
