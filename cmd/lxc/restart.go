package lxc

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart an LXC container",
	Long:  `Restart a running LXC container on the specified node.`,
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

		task, err := container.Reboot(ctx)
		if err != nil {
			fmt.Fprintf(out, "Error restarting container: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Restart task: %s\n", task.UPID)
		fmt.Fprintf(out, "Container %d restarted successfully\n", vmid)
	},
}

func init() {
	restartCmd.Flags().StringP("node", "n", "", "Node name")
	restartCmd.Flags().IntP("vmid", "i", 0, "Container ID")
	restartCmd.MarkFlagRequired("node")
	restartCmd.MarkFlagRequired("vmid")
}
