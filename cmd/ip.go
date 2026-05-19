package cmd

import (
	"io"
	"strings"

	"github.com/spf13/cobra"

	"tin/internal/ip"
	"tin/internal/shared/ui"
)

func NewIpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip",
		Short: "Show local interface and public IP addresses",
		Long:  "List local network interfaces with their IPv4 and IPv6 addresses, and fetch the current public IP address.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := ip.Gather(cmd.Context())
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			printInterfaces(out, info.Interfaces)
			printPublic(out, info.Public, info.PublicErr)
			return nil
		},
	}

	return cmd
}

func printInterfaces(out io.Writer, ifaces []ip.Interface) {
	if len(ifaces) == 0 {
		ui.PrintNote(out, "no active interfaces")
		return
	}

	rows := make([]ui.Row, 0, len(ifaces)*2)
	for _, iface := range ifaces {
		if len(iface.V4) > 0 {
			rows = append(rows, ui.Row{Label: iface.Name + " v4", Value: strings.Join(iface.V4, ", ")})
		}
		if len(iface.V6) > 0 {
			rows = append(rows, ui.Row{Label: iface.Name + " v6", Value: strings.Join(iface.V6, ", ")})
		}
	}
	ui.PrintBox(out, "local", rows)
}

func printPublic(out io.Writer, addr string, err error) {
	if err != nil {
		ui.PrintBox(out, "public", []ui.Row{
			{Label: "error", Value: err.Error()},
		})
		return
	}
	ui.PrintBox(out, "public", []ui.Row{
		{Label: "v4", Value: addr},
	})
}
