package cobrahelp

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"tin/internal/shared/helptext"
)

const qualityHelpAnnotation = "tin.help.quality"

var addTemplateFuncs sync.Once

func Apply(root *cobra.Command) {
	addTemplateFuncs.Do(func() {
		cobra.AddTemplateFunc("qualityHelp", qualityHelp)
		cobra.AddTemplateFunc("commandHelp", commandHelp)
	})
	apply(root)
}

func SetQualityHelp(cmd *cobra.Command, sections []helptext.Section) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations[qualityHelpAnnotation] = helptext.RenderSections(sections)
}

func apply(cmd *cobra.Command) {
	cmd.SetHelpTemplate(helpTemplate)
	for _, child := range cmd.Commands() {
		apply(child)
	}
}

func qualityHelp(cmd *cobra.Command) string {
	if cmd == nil || cmd.Annotations == nil {
		return ""
	}
	return cmd.Annotations[qualityHelpAnnotation]
}

func commandHelp(cmd *cobra.Command) string {
	if cmd == nil || !cmd.HasAvailableSubCommands() {
		return ""
	}

	var children []*cobra.Command
	padding := 0
	for _, child := range cmd.Commands() {
		if !child.IsAvailableCommand() && child.Name() != "help" {
			continue
		}
		children = append(children, child)
		if len(child.Name()) > padding {
			padding = len(child.Name())
		}
	}

	var out strings.Builder
	for _, child := range children {
		fmt.Fprintf(&out, "  %-*s  %s\n", padding, child.Name(), child.Short)
	}
	return strings.TrimRight(out.String(), "\n")
}

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{with qualityHelp .}}

Quality Values:
{{.}}{{end}}{{with commandHelp .}}

Commands:
{{.}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
