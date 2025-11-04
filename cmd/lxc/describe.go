package lxc

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe an LXC container",
	Long:  `Get detailed information about a specific LXC container.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		vmid, _ := cmd.Flags().GetInt("vmid")

		if nodeName == "" || vmid == 0 {
			fmt.Fprintln(out, "Error: node and vmid are required")
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		node, err := client.Node(ctx, nodeName)
		if err != nil {
			fmt.Fprintf(out, "Error getting node: %v\n", err)
			return
		}

		container, err := node.Container(ctx, vmid)
		if err != nil {
			fmt.Fprintf(out, "Error getting container: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Container %d Details:\n", vmid)
		fmt.Fprintf(out, "  Name: %s\n", container.Name)
		fmt.Fprintf(out, "  Status: %s\n", container.Status)
		fmt.Fprintf(out, "  Node: %s\n", container.Node)
		fmt.Fprintf(out, "  Type: %s\n", container.Type)
		fmt.Fprintf(out, "  Uptime: %d seconds\n", container.Uptime)
		fmt.Fprintf(out, "  CPU: %.2f%%\n", container.CPU)
		fmt.Fprintf(out, "  Memory: %d MB / %d MB\n", container.Mem/(1024*1024), container.MaxMem/(1024*1024))
		fmt.Fprintf(out, "  Disk: %d GB / %d GB\n", container.Disk/(1024*1024*1024), container.MaxDisk/(1024*1024*1024))
		fmt.Fprintf(out, "  Swap: %d MB / %d MB\n", container.Swap/(1024*1024), container.MaxSwap/(1024*1024))

		// Get additional config details
		config, err := container.Config(ctx)
		if err == nil && config != nil {
			fmt.Fprintln(out, "\nConfiguration:")
			if config.Hostname != "" {
				fmt.Fprintf(out, "  Hostname: %s\n", config.Hostname)
			}
			if config.Cores > 0 {
				fmt.Fprintf(out, "  CPU Cores: %d\n", config.Cores)
			}
			if config.OSTemplate != "" {
				fmt.Fprintf(out, "  OS Template: %s\n", config.OSTemplate)
			}
		}
	},
}

func init() {
	describeCmd.Flags().StringP("node", "n", "", "Node name")
	describeCmd.Flags().IntP("vmid", "i", 0, "Container ID")
	describeCmd.MarkFlagRequired("node")
	describeCmd.MarkFlagRequired("vmid")
}
