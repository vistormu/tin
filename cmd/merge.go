package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tin/internal/merge"
	"tin/internal/shared/fsx"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewMergeCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "merge <file|glob>...",
		Short: "Merge PDF files into one",
		Long:  "Merge two or more PDF files into a single PDF using ghostscript. Accepts file paths and shell-style glob patterns.",
		Example: `  tin merge a.pdf b.pdf
  tin merge a.pdf b.pdf c.pdf -o combined.pdf
  tin merge '*.pdf' -o combined.pdf`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}
			if len(inputs) < 2 {
				return fmt.Errorf("at least two input files are required")
			}

			result, err := merge.Merge(cmd.Context(), merge.Options{
				Inputs: inputs,
				Output: output,
			})
			if err != nil {
				return err
			}

			ui.PrintBox(cmd.OutOrStdout(), "merged", []ui.Row{
				{Label: "files", Value: ui.Plural(len(result.Inputs), "file")},
				{Label: "before", Value: size.Format(result.BeforeSize)},
				{Label: "after", Value: size.Format(result.AfterSize)},
				{Label: "output", Value: result.Output},
			})
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (default: <first-input-dir>/merged.pdf)")

	return cmd
}
