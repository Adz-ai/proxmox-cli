package lxc

import (
	"fmt"
	"github.com/Adz-ai/proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "List all LXC containers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			nodes, err := client.Nodes(ctx)
			if err != nil {
				return fmt.Errorf("list nodes: %w", err)
			}

			fmt.Fprintln(out, "LXC Containers:")
			fmt.Fprintln(out, "================")

			totalContainers := 0
			for _, nodeStatus := range nodes {
				node, err := client.Node(ctx, nodeStatus.Node)
				if err != nil {
					return fmt.Errorf("get node %q: %w", nodeStatus.Node, err)
				}

				containers, err := node.Containers(ctx)
				if err != nil {
					return fmt.Errorf("list containers on node %q: %w", nodeStatus.Node, err)
				}

				if len(containers) > 0 {
					fmt.Fprintf(out, "\nNode: %s\n", nodeStatus.Node)
					fmt.Fprintf(out, "%-10s %-20s %-10s %-12s %-12s\n", "VMID", "Name", "Status", "Type", "Uptime")
					fmt.Fprintf(out, "%-10s %-20s %-10s %-12s %-12s\n", "----", "----", "------", "----", "------")

					for _, container := range containers {
						uptime := "N/A"
						if container.Uptime > 0 {
							days := container.Uptime / 86400
							hours := (container.Uptime % 86400) / 3600
							uptime = fmt.Sprintf("%dd %dh", days, hours)
						}
						fmt.Fprintf(out, "%-10v %-20s %-10s %-12s %-12s\n",
							container.VMID,
							container.Name,
							container.Status,
							"lxc",
							uptime)
						totalContainers++
					}
				}
			}

			if totalContainers == 0 {
				fmt.Fprintln(out, "No LXC containers found in the cluster")
			}
			return nil
		},
	}
}
