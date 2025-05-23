package lxc

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"proxmox-cli/cmd/utility"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "List all LXC containers",
	Run: func(cmd *cobra.Command, args []string) {
		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Println(err)
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		nodes, err := client.Nodes(ctx)
		if err != nil {
			fmt.Printf("Error fetching nodes: %v\n", err)
			return
		}

		fmt.Println("LXC Containers:")
		fmt.Println("================")

		totalContainers := 0
		for _, nodeStatus := range nodes {
			// Get node object
			node, err := client.Node(ctx, nodeStatus.Node)
			if err != nil {
				fmt.Printf("Error getting node %s: %v\n", nodeStatus.Node, err)
				continue
			}

			// Get containers on this node
			containers, err := node.Containers(ctx)
			if err != nil {
				fmt.Printf("Error fetching containers from node %s: %v\n", nodeStatus.Node, err)
				continue
			}

			if len(containers) > 0 {
				fmt.Printf("\nNode: %s\n", nodeStatus.Node)
				fmt.Printf("%-10s %-20s %-10s %-12s %-12s\n", "VMID", "Name", "Status", "Type", "Uptime")
				fmt.Printf("%-10s %-20s %-10s %-12s %-12s\n", "----", "----", "------", "----", "------")
				
				for _, container := range containers {
					uptime := "N/A"
					if container.Uptime > 0 {
						days := container.Uptime / 86400
						hours := (container.Uptime % 86400) / 3600
						uptime = fmt.Sprintf("%dd %dh", days, hours)
					}
					fmt.Printf("%-10v %-20s %-10s %-12s %-12s\n",
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
			fmt.Println("No LXC containers found in the cluster")
		}
	},
}