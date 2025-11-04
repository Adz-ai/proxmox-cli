package vm

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a virtual machine",
	Long:  `Stop a running virtual machine on the specified node.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		vmid, _ := cmd.Flags().GetInt("vmid")
		force, _ := cmd.Flags().GetBool("force")

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

		vm, err := node.VirtualMachine(ctx, vmid)
		if err != nil {
			fmt.Fprintf(out, "Error getting VM: %v\n", err)
			return
		}

		var task *proxmox.Task
		if force {
			fmt.Fprintln(out, "Force shutdown requested")
			task, err = vm.Stop(ctx)
		} else {
			task, err = vm.Shutdown(ctx)
		}

		if err != nil {
			fmt.Fprintf(out, "Error stopping VM: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Stop task: %s\n", task.UPID)
		fmt.Fprintf(out, "VM %d stopped successfully\n", vmid)
	},
}

func init() {
	stopCmd.Flags().StringP("node", "n", "", "Node name")
	stopCmd.Flags().IntP("vmid", "i", 0, "VM ID")
	stopCmd.Flags().BoolP("force", "f", false, "Force stop (power off)")
	stopCmd.MarkFlagRequired("node")
	stopCmd.MarkFlagRequired("vmid")
}
