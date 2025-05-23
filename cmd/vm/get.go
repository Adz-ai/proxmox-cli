package vm

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "List all virtual machines",
	Long:  `Display a list of all virtual machines across all nodes in the Proxmox cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		
		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		nodes, err := client.Nodes(ctx)
		if err != nil {
			fmt.Fprintf(out, "Error fetching nodes: %v\n", err)
			return
		}

		fmt.Fprintln(out, "Virtual Machines:")
		fmt.Fprintln(out, "=================")

		totalVMs := 0
		for _, nodeStatus := range nodes {
			node, err := client.Node(ctx, nodeStatus.Node)
			if err != nil {
				fmt.Fprintf(out, "Error getting node %s: %v\n", nodeStatus.Node, err)
				continue
			}

			vms, err := node.VirtualMachines(ctx)
			if err != nil {
				fmt.Fprintf(out, "Error fetching VMs from node %s: %v\n", nodeStatus.Node, err)
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
	},
}