package nodes

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show detailed information about a node",
	Long:  `Display detailed information about a specific node in the Proxmox cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		nodeName, _ := cmd.Flags().GetString("name")

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		// Get the specific node (we'll use it later if needed for more operations)
		_, err = client.Node(ctx, nodeName)
		if err != nil {
			fmt.Fprintf(out, "Error getting node %s: %v\n", nodeName, err)
			return
		}

		// Get node status from the nodes list
		nodes, err := client.Nodes(ctx)
		if err != nil {
			fmt.Fprintf(out, "Error fetching node status: %v\n", err)
			return
		}

		// Find our node in the list
		var nodeStatus *proxmox.NodeStatus
		for _, n := range nodes {
			if n.Node == nodeName {
				nodeStatus = n
				break
			}
		}

		if nodeStatus == nil {
			fmt.Fprintf(out, "Node %s not found in cluster\n", nodeName)
			return
		}

		// Display node information
		fmt.Fprintln(out, "Node Information")
		fmt.Fprintln(out, "================")
		fmt.Fprintf(out, "Name: %s\n", nodeStatus.Node)
		fmt.Fprintf(out, "Status: %s\n", nodeStatus.Status)
		fmt.Fprintf(out, "Type: %s\n", nodeStatus.Type)
		
		if nodeStatus.Uptime > 0 {
			days := nodeStatus.Uptime / 86400
			hours := (nodeStatus.Uptime % 86400) / 3600
			fmt.Fprintf(out, "Uptime: %dd %dh\n", days, hours)
		}
		
		// CPU information
		if nodeStatus.MaxCPU > 0 {
			cpuUsage := nodeStatus.CPU * 100
			fmt.Fprintf(out, "\nCPU Usage: %.2f%%\n", cpuUsage)
			fmt.Fprintf(out, "CPU Cores: %d\n", nodeStatus.MaxCPU)
		}
		
		// Memory information
		if nodeStatus.MaxMem > 0 {
			memUsageGB := float64(nodeStatus.Mem) / (1024 * 1024 * 1024)
			maxMemGB := float64(nodeStatus.MaxMem) / (1024 * 1024 * 1024)
			memUsagePercent := (float64(nodeStatus.Mem) / float64(nodeStatus.MaxMem)) * 100
			fmt.Fprintf(out, "\nMemory: %.1f GB / %.1f GB (%.1f%%)\n", memUsageGB, maxMemGB, memUsagePercent)
		}
		
		// Disk information
		if nodeStatus.MaxDisk > 0 {
			diskUsageGB := float64(nodeStatus.Disk) / (1024 * 1024 * 1024)
			maxDiskGB := float64(nodeStatus.MaxDisk) / (1024 * 1024 * 1024)
			diskUsagePercent := (float64(nodeStatus.Disk) / float64(nodeStatus.MaxDisk)) * 100
			fmt.Fprintf(out, "Root Disk: %.1f GB / %.1f GB (%.1f%%)\n", diskUsageGB, maxDiskGB, diskUsagePercent)
		}
	},
}

func init() {
	describeCmd.Flags().StringP("name", "n", "", "Name of the node to describe (required)")
	describeCmd.MarkFlagRequired("name")
}