package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"tin/internal/find"
)

func NewFindCommand() *cobra.Command {
	var extFlag string
	var nameFlag string
	var typeFlag string

	cmd := &cobra.Command{
		Use:   "find [dir]",
		Short: "Find files by extension or name pattern",
		Long:  "Recursively walk a directory and print files matching the given filters. Defaults to the current directory.",
		Example: `  tin find . -e go
  tin find ./src -e go,md
  tin find . -n '*test*'
  tin find . --type d -n '*tmp*'`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) > 0 {
				root = args[0]
			}

			ft, err := find.ParseType(typeFlag)
			if err != nil {
				return err
			}

			matches, err := find.Find(find.Options{
				Root:       root,
				Extensions: splitExts(extFlag),
				NameGlob:   nameFlag,
				Type:       ft,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			for _, m := range matches {
				fmt.Fprintln(out, m.Path)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&extFlag, "ext", "e", "", "comma-separated extensions (e.g. go,md)")
	cmd.Flags().StringVarP(&nameFlag, "name", "n", "", "basename glob pattern")
	cmd.Flags().StringVarP(&typeFlag, "type", "t", "f", "type filter: f (files), d (directories), or all")

	return cmd
}

func splitExts(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
