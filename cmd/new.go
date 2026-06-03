package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tin/internal/shared/fsx"
	"tin/internal/shared/ui"
)

type newTarget struct {
	path  string
	isDir bool
}

func NewNewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <path>...",
		Short: "Create files or directories with parents auto-created",
		Long:  "Create one or more files or directories. Paths ending with / become directories; everything else becomes an empty file. Missing parent directories are created automatically. Aborts before creating anything if any target already exists.",
		Example: `  tin new notes.md
  tin new docs/
  tin new src/api/handler.go docs/guides/`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targets := make([]newTarget, 0, len(args))
			seen := make(map[string]struct{}, len(args))
			for _, raw := range args {
				isDir := strings.HasSuffix(raw, "/")
				clean := filepath.Clean(raw)
				if clean == "." || clean == "" {
					return fmt.Errorf("invalid path %q", raw)
				}
				if _, ok := seen[clean]; ok {
					return fmt.Errorf("duplicate target %q", raw)
				}
				seen[clean] = struct{}{}

				exists, err := fsx.Exists(clean)
				if err != nil {
					return err
				}
				if exists {
					return fmt.Errorf("target %q already exists", raw)
				}
				targets = append(targets, newTarget{path: clean, isDir: isDir})
			}

			var files, dirs int
			for _, t := range targets {
				if t.isDir {
					if err := os.MkdirAll(t.path, 0o755); err != nil {
						return err
					}
					dirs++
					continue
				}
				if parent := filepath.Dir(t.path); parent != "." && parent != "/" {
					if err := os.MkdirAll(parent, 0o755); err != nil {
						return err
					}
				}
				f, err := os.OpenFile(t.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
				if err != nil {
					return fmt.Errorf("create %q: %w", t.path, err)
				}
				if err := f.Close(); err != nil {
					return err
				}
				files++
			}

			ui.PrintBox(cmd.OutOrStdout(), "created", []ui.Row{
				{Label: "files", Value: ui.Plural(files, "file")},
				{Label: "dirs", Value: ui.Plural(dirs, "directory")},
			})
			return nil
		},
	}

	return cmd
}
