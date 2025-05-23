package lxc

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an LXC container",
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		
		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		vmid, _ := cmd.Flags().GetInt("vmid")
		// force, _ := cmd.Flags().GetBool("force") // TODO: use force flag

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

		task, err := container.Delete(ctx)
		if err != nil {
			fmt.Fprintf(out, "Error deleting container: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Delete task: %s\n", task.UPID)
		fmt.Fprintf(out, "Container %d deleted successfully\n", vmid)
	},
}

func init() {
	deleteCmd.Flags().StringP("node", "n", "", "Node name")
	deleteCmd.Flags().IntP("vmid", "i", 0, "Container ID")
	deleteCmd.Flags().BoolP("force", "f", false, "Force deletion")
	deleteCmd.MarkFlagRequired("node")
	deleteCmd.MarkFlagRequired("vmid")
}