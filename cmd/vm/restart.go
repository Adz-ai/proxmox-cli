package vm

import (
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

func newRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart a virtual machine",
		Long:  `Reboot a running virtual machine on the specified node.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			node, id, err := vmTargetFromFlags(cmd)
			if err != nil {
				return err
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			retrievedNode, err := client.Node(ctx, node)
			if err != nil {
				return fmt.Errorf("get node %q: %w", node, err)
			}

			vm, err := retrievedNode.VirtualMachine(ctx, id)
			if err != nil {
				return fmt.Errorf("get VM %d: %w", id, err)
			}

			task, err := vm.Reboot(ctx)
			if err != nil {
				return fmt.Errorf("restart VM %d: %w", id, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("restart VM %d: %w", id, err)
			}

			fmt.Fprintf(out, "VM %d restarted successfully\n", id)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	return cmd
}
