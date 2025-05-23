package nodes

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "List all nodes in the cluster",
	Long:  `Display a list of all nodes in the Proxmox cluster with their status and uptime.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		
		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		client := utility.GetClient()

		nodes, err := client.Nodes(context.Background())
		if err != nil {
			fmt.Fprintf(out, "Error fetching nodes: %v\n", err)
			return
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
	},
}

func init() {
	Cmd.AddCommand(getCmd)
}