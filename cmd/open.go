package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"tin/internal/opener"
	"tin/internal/shared/fsx"
)

func NewOpenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open <target>...",
		Short: "Open a file, directory, or URL with the system default handler",
		Long:  "Open one or more targets with the platform's default handler: open on macOS, xdg-open on Linux. URL-looking arguments (containing ://) are passed through; everything else must point to an existing local path.",
		Example: `  tin open report.pdf
  tin open ./photos
  tin open https://example.com`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, target := range args {
				if !isURL(target) {
					exists, err := fsx.Exists(target)
					if err != nil {
						return err
					}
					if !exists {
						return fmt.Errorf("target %q does not exist", target)
					}
				}
			}

			for _, target := range args {
				if err := opener.Open(cmd.Context(), target); err != nil {
					return err
				}
			}
			return nil
		},
	}

	return cmd
}

func isURL(s string) bool {
	return strings.Contains(s, "://")
}
