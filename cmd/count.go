package cmd

import (
	"github.com/spf13/cobra"

	"tin/internal/count"
	"tin/internal/shared/fsx"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewCountCommand() *cobra.Command {
	var extFlag string

	cmd := &cobra.Command{
		Use:   "count <path|glob>...",
		Short: "Count files, lines, and bytes in paths",
		Long:  "Walk one or more files or directories and report total file, line, and byte counts. Accepts paths and shell-style glob patterns.",
		Example: `  tin count .
  tin count ./src --ext go
  tin count '*.md'`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}

			result, err := count.Count(count.Options{
				Paths:      paths,
				Extensions: splitExts(extFlag),
			})
			if err != nil {
				return err
			}

			ui.PrintBox(cmd.OutOrStdout(), "Counted", []ui.Row{
				{Label: "Files", Value: ui.Plural(result.Files, "file")},
				{Label: "Lines", Value: ui.Plural(result.Lines, "line")},
				{Label: "Bytes", Value: size.Format(result.Bytes)},
			})
			return nil
		},
	}

	cmd.Flags().StringVarP(&extFlag, "ext", "e", "", "comma-separated extensions (e.g. go,md)")

	return cmd
}
