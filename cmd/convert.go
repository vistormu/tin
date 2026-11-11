package cmd

import (
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
		Short: "Convert images, pdfs, videos, or audio between formats",
		Long:  "Convert one or more files between supported formats. Routes through ImageMagick (images, pdf↔image), poppler (pdf→svg), or ffmpeg (video, audio, video→audio) based on input and target extensions. Accepts file paths and shell-style glob patterns.",
		Example: `  tin convert photo.heic -t jpg
  tin convert photo.png -t webp
  tin convert '*.png' -t webp
  tin convert scan.tiff -t png
  tin convert doc.pdf -t png
  tin convert diagram.pdf -t svg
  tin convert page-*.jpg -t pdf -o album.pdf
  tin convert clip.mp4 -t mkv
  tin convert clip.mp4 -t mp3
  tin convert song.wav -t mp3`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}

			results, err := convert.Convert(cmd.Context(), convert.Options{
				Inputs: inputs,
				Output: output,
				Target: target,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			for _, result := range results {
				rows := []ui.Row{
					{Label: "before", Value: size.Format(result.BeforeSize)},
					{Label: "after", Value: size.Format(result.AfterSize)},
				}
				if len(result.Inputs) > 1 {
					rows = append(rows, ui.Row{Label: "inputs", Value: ui.Plural(len(result.Inputs), "file")})
				}
				switch len(result.Outputs) {
				case 0:
				case 1:
					rows = append(rows, ui.Row{Label: "output", Value: result.Outputs[0]})
				default:
					rows = append(rows,
						ui.Row{Label: "pages", Value: ui.Plural(len(result.Outputs), "page")},
						ui.Row{Label: "first", Value: result.Outputs[0]},
						ui.Row{Label: "last", Value: result.Outputs[len(result.Outputs)-1]},
					)
				}
				ui.PrintBox(out, "converted", rows)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "target extension (e.g. jpg, heic, tiff, pdf, svg, mkv, mp3)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (single input only, or combined image→pdf)")

	return cmd
}
