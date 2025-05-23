package lxc

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop an LXC container",
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

		task, err := container.Stop(ctx)
		if err != nil {
			fmt.Fprintf(out, "Error stopping container: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Stop task: %s\n", task.UPID)
		fmt.Fprintf(out, "Container %d stopped successfully\n", vmid)
	},
}

func init() {
	stopCmd.Flags().StringP("node", "n", "", "Node name")
	stopCmd.Flags().IntP("vmid", "i", 0, "Container ID")
	stopCmd.MarkFlagRequired("node")
	stopCmd.MarkFlagRequired("vmid")
}