package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"tin/internal/replace"
	"tin/internal/shared/confirm"
	"tin/internal/shared/ui"
)

func NewReplaceCommand() *cobra.Command {
	var dryRun bool
	var yes bool
	var extFlag string
	var nameFlag string

	cmd := &cobra.Command{
		Use:   "replace <old> <new> [dir]",
		Short: "Replace all occurrences of a string in files",
		Long:  "Recursively replace every occurrence of a literal string with another across files under a directory (current directory by default). Matching is literal, not regex. Binary files and the .git directory are skipped. Filter by extension or name to limit which files are touched.",
		Example: `  tin replace foo bar
  tin replace example.com/old example.com/new . -e go
  tin replace 'old text' 'new text' ./docs -e md --dry-run
  tin replace v1.2.3 v1.2.4 . -n '*.toml' --yes`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) > 2 {
				root = args[2]
			}

			plan, err := replace.BuildPlan(replace.Options{
				Root:       root,
				Old:        args[0],
				New:        args[1],
				Extensions: splitExts(extFlag),
				NameGlob:   nameFlag,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			printReplacePlan(out, plan)

			if dryRun {
				ui.PrintNote(out, "dry run: no files were changed")
				return nil
			}
			if plan.FileCount() == 0 {
				ui.PrintNote(out, "no matches found")
				return nil
			}

			prompt := fmt.Sprintf("replace %s in %s?",
				ui.Plural(plan.TotalMatches(), "occurrence"),
				ui.Plural(plan.FileCount(), "file"))
			ok, err := confirm.Ask(yes, prompt, cmd.InOrStdin(), out)
			if err != nil {
				return err
			}
			if !ok {
				ui.PrintWarning(out, "aborted")
				return nil
			}

			result, err := replace.Apply(plan)
			if err != nil {
				return err
			}
			ui.PrintBox(out, "replaced", []ui.Row{
				{Label: "files", Value: ui.Plural(result.Files, "file")},
				{Label: "occurrences", Value: ui.Plural(result.Matches, "occurrence")},
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned replacements without changing files")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	cmd.Flags().StringVarP(&extFlag, "ext", "e", "", "comma-separated extensions to limit replacement (e.g. go,md)")
	cmd.Flags().StringVarP(&nameFlag, "name", "n", "", "basename glob pattern to limit replacement")

	return cmd
}

func printReplacePlan(out io.Writer, plan replace.Plan) {
	ui.PrintBox(out, "replace plan", []ui.Row{
		{Label: "files", Value: ui.Plural(plan.FileCount(), "file")},
		{Label: "occurrences", Value: ui.Plural(plan.TotalMatches(), "occurrence")},
	})
	for _, c := range plan.Changes {
		fmt.Fprintf(out, "  %s (%s)\n", c.Path, ui.Plural(c.Count, "occurrence"))
	}
}
