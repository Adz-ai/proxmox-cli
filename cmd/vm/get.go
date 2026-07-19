package vm

import (
	"errors"
	"fmt"
	"github.com/Adz-ai/proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "List all virtual machines",
		Long:  `Display a list of all virtual machines across all nodes in the Proxmox cluster.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}
			ctx := cmd.Context()

			nodes, err := client.Nodes(ctx)
			if err != nil {
				return fmt.Errorf("fetch cluster nodes: %w", err)
			}

			fmt.Fprintln(out, "Virtual Machines:")
			fmt.Fprintln(out, "=================")

			totalVMs := 0
			var nodeErrors []error
			for _, nodeStatus := range nodes {
				node, err := client.Node(ctx, nodeStatus.Node)
				if err != nil {
					nodeErrors = append(nodeErrors, fmt.Errorf("get node %q: %w", nodeStatus.Node, err))
					continue
				}

				vms, err := node.VirtualMachines(ctx)
				if err != nil {
					nodeErrors = append(nodeErrors, fmt.Errorf("fetch VMs from node %q: %w", nodeStatus.Node, err))
					continue
				}

				if len(vms) > 0 {
					fmt.Fprintf(out, "\nNode: %s\n", nodeStatus.Node)
					fmt.Fprintf(out, "%-10s %-20s %-10s %-12s\n", "VMID", "Name", "Status", "Uptime")
					fmt.Fprintf(out, "%-10s %-20s %-10s %-12s\n", "----", "----", "------", "------")

					for _, vm := range vms {
						uptime := "N/A"
						if vm.Uptime > 0 {
							days := vm.Uptime / 86400
							hours := (vm.Uptime % 86400) / 3600
							uptime = fmt.Sprintf("%dd %dh", days, hours)
						}
						fmt.Fprintf(out, "%-10v %-20s %-10s %-12s\n",
							vm.VMID,
							vm.Name,
							vm.Status,
							uptime)
						totalVMs++
					}
				}
			}

			if totalVMs == 0 {
				fmt.Fprintln(out, "No virtual machines found in the cluster")
			}
			if err := errors.Join(nodeErrors...); err != nil {
				return fmt.Errorf("list virtual machines: %w", err)
			}
			return nil
		},
	}
}
