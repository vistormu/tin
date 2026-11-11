package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"tin/internal/archive"
	"tin/internal/shared/fsx"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewUnzipCommand() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "unzip <archive.zip|glob>...",
		Short: "Extract zip archives",
		Long:  "Extract one or more zip archives. Accepts paths and shell-style glob patterns. Each archive is extracted into its own directory named after the archive.",
		Example: `  tin unzip photos.zip
  tin unzip '*.zip'
  tin unzip '*.zip' -o extracted`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			failed := 0
			for _, input := range inputs {
				target := ""
				if outputDir != "" {
					target = filepath.Join(outputDir, archive.UnzipDirName(input))
				}

				result, err := archive.Unzip(archive.UnzipOptions{
					Input:     input,
					OutputDir: target,
				})
				if err != nil {
					// A single archive reports its own error verbatim; a batch
					// keeps going so one bad archive cannot strand the rest.
					if len(inputs) == 1 {
						return err
					}
					failed++
					ui.PrintWarning(cmd.ErrOrStderr(), fmt.Sprintf("skipped %s: %v", input, err))
					continue
				}

				ui.PrintBox(out, "unzipped", []ui.Row{
					{Label: "files", Value: ui.Plural(result.Files, "file")},
					{Label: "bytes", Value: size.Format(result.Bytes)},
					{Label: "output", Value: result.OutputDir},
				})
			}

			if failed > 0 {
				return fmt.Errorf("%s failed to extract, %d of %d extracted", ui.Plural(failed, "archive"), len(inputs)-failed, len(inputs))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "directory to place each archive's output directory in")
	return cmd
}
