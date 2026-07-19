package nodes

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

type nodeDetails struct {
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Type      string  `json:"type"`
	Uptime    uint64  `json:"uptime_seconds"`
	CPUUsage  float64 `json:"cpu_usage"`
	CPUCores  int     `json:"cpu_cores"`
	Memory    uint64  `json:"memory_bytes"`
	MaxMemory uint64  `json:"max_memory_bytes"`
	Disk      uint64  `json:"disk_bytes"`
	MaxDisk   uint64  `json:"max_disk_bytes"`
}

func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show detailed information about a node",
		Long:  `Display detailed information about a specific node in the Proxmox cluster.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			nodeName, err := cmd.Flags().GetString("name")
			if err != nil {
				return fmt.Errorf("get name flag: %w", err)
			}
			nodeName = strings.TrimSpace(nodeName)
			if nodeName == "" {
				return fmt.Errorf("validate node name: name cannot be empty")
			}
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}
			ctx := cmd.Context()

			// Get the specific node (we'll use it later if needed for more operations)
			_, err = client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			// Get node status from the nodes list
			nodes, err := client.Nodes(ctx)
			if err != nil {
				return fmt.Errorf("fetch status for node %q: %w", nodeName, err)
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
				return fmt.Errorf("node %q not found in cluster", nodeName)
			}

			if format == "json" {
				return utility.PrintJSON(out, nodeDetails{
					Name:      nodeStatus.Node,
					Status:    nodeStatus.Status,
					Type:      nodeStatus.Type,
					Uptime:    nodeStatus.Uptime,
					CPUUsage:  nodeStatus.CPU,
					CPUCores:  nodeStatus.MaxCPU,
					Memory:    nodeStatus.Mem,
					MaxMemory: nodeStatus.MaxMem,
					Disk:      nodeStatus.Disk,
					MaxDisk:   nodeStatus.MaxDisk,
				})
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
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "Name of the node to describe (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	utility.AddOutputFlag(cmd)

	return cmd
}
