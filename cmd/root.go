package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"tin/internal/platform"
	"tin/internal/shared/cobrahelp"
	"tin/internal/shared/ui"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "tin",
		Short:         "Tiny personal CLI tools",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return platform.EnsureSupported()
		},
	}

	root.AddCommand(
		NewCompressCommand(),
		NewFlattenCommand(),
		NewResizeCommand(),
		NewSizeCommand(),
		NewTrashCommand(),
		NewUnzipCommand(),
		NewZipCommand(),
	)

	cobrahelp.Apply(root)

	return root
}

func Execute() {
	if code := execute(NewRootCommand()); code != 0 {
		os.Exit(code)
	}
}

func execute(root *cobra.Command) int {
	if err := root.Execute(); err != nil {
		ui.PrintError(root.ErrOrStderr(), err)
		return 1
	}
	return 0
}
