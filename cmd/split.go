package cmd

import (
	"github.com/spf13/cobra"

	"tin/internal/shared/ui"
	"tin/internal/split"
)

func NewSplitCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "split <file.pdf>",
		Short: "Split a PDF into one file per page",
		Long:  "Split a PDF file into separate single-page PDFs using ghostscript.",
		Example: `  tin split report.pdf
  tin split report.pdf -o report-pages`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := split.Split(cmd.Context(), split.Options{
				Input:     args[0],
				OutputDir: output,
			})
			if err != nil {
				return err
			}

			ui.PrintBox(cmd.OutOrStdout(), "Split", []ui.Row{
				{Label: "Pages", Value: ui.Plural(len(result.Files), "page")},
				{Label: "Output", Value: result.OutputDir},
			})
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output directory (default: <name>-pages)")

	return cmd
}
