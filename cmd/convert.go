package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tin/internal/convert"
	"tin/internal/shared/fsx"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewConvertCommand() *cobra.Command {
	var target string
	var output string

	cmd := &cobra.Command{
		Use:   "convert <file|glob>...",
		Short: "Convert an image file to another format",
		Long:  "Convert one or more supported image files between .jpg, .jpeg, .png, and .webp. Accepts file paths and shell-style glob patterns.",
		Example: `  tin convert photo.png -t webp
  tin convert photo.png -o photo.jpg
  tin convert '*.png' -t webp`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}
			if len(inputs) > 1 && output != "" {
				return fmt.Errorf("--output is only allowed with a single input")
			}

			out := cmd.OutOrStdout()
			for _, input := range inputs {
				result, err := convert.Convert(cmd.Context(), convert.Options{
					Input:  input,
					Output: output,
					Target: target,
				})
				if err != nil {
					return err
				}

				ui.PrintBox(out, "Converted", []ui.Row{
					{Label: "Before", Value: size.Format(result.BeforeSize)},
					{Label: "After", Value: size.Format(result.AfterSize)},
					{Label: "Output", Value: result.Output},
				})
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "target extension (jpg, png, webp)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (single input only)")

	return cmd
}
