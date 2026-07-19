package nodes

import (
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "List all nodes in the cluster",
		Long:  `Display a list of all nodes in the Proxmox cluster with their status and uptime.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			nodes, err := client.Nodes(cmd.Context())
			if err != nil {
				return fmt.Errorf("fetch cluster nodes: %w", err)
			}

			fmt.Fprintln(out, "Nodes in cluster:")
			fmt.Fprintln(out, "=================")
			fmt.Fprintf(out, "%-15s %-10s %-8s %-12s\n", "Node", "Status", "Type", "Uptime")
			fmt.Fprintf(out, "%-15s %-10s %-8s %-12s\n", "----", "------", "----", "------")
			for _, node := range nodes {
				uptime := "N/A"
				if node.Uptime > 0 {
					days := node.Uptime / 86400
					hours := (node.Uptime % 86400) / 3600
					uptime = fmt.Sprintf("%dd %dh", days, hours)
				}
				fmt.Fprintf(out, "%-15s %-10s %-8s %-12s\n",
					node.Node,
					node.Status,
					node.Type,
					uptime)
			}
			return nil
		},
	}
}
