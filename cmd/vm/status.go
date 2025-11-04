package vm

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of a virtual machine",
	Long:  `Get the current status of a virtual machine on the specified node.`,
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

		vm, err := node.VirtualMachine(ctx, vmid)
		if err != nil {
			fmt.Fprintf(out, "Error getting VM: %v\n", err)
			return
		}

		fmt.Fprintf(out, "VM %d Status:\n", vmid)
		fmt.Fprintf(out, "  Name: %s\n", vm.Name)
		fmt.Fprintf(out, "  Status: %s\n", vm.Status)
		fmt.Fprintf(out, "  Node: %s\n", vm.Node)
		fmt.Fprintf(out, "  Uptime: %d seconds\n", vm.Uptime)
		fmt.Fprintf(out, "  CPU: %.2f%%\n", vm.CPU)
		fmt.Fprintf(out, "  Memory: %d MB / %d MB\n", vm.Mem/(1024*1024), vm.MaxMem/(1024*1024))
		fmt.Fprintf(out, "  Disk: %d GB / %d GB\n", vm.Disk/(1024*1024*1024), vm.MaxDisk/(1024*1024*1024))
	},
}

func init() {
	statusCmd.Flags().StringP("node", "n", "", "Node name")
	statusCmd.Flags().IntP("vmid", "i", 0, "VM ID")
	statusCmd.MarkFlagRequired("node")
	statusCmd.MarkFlagRequired("vmid")
}
