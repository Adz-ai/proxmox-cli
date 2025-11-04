package vm

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a virtual machine",
	Long:  `Restart a running virtual machine on the specified node.`,
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
			fmt.Fprintln(out, "Force restart requested")
			task, err = vm.Reset(ctx)
		} else {
			task, err = vm.Reboot(ctx)
		}

		if err != nil {
			fmt.Fprintf(out, "Error restarting VM: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Restart task: %s\n", task.UPID)
		fmt.Fprintf(out, "VM %d restarted successfully\n", vmid)
	},
}

func init() {
	restartCmd.Flags().StringP("node", "n", "", "Node name")
	restartCmd.Flags().IntP("vmid", "i", 0, "VM ID")
	restartCmd.Flags().BoolP("force", "f", false, "Force restart (reset)")
	restartCmd.MarkFlagRequired("node")
	restartCmd.MarkFlagRequired("vmid")
}
