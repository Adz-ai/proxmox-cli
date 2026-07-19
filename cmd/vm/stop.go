package vm

import (
	"fmt"
	"time"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

func newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a virtual machine",
		Long:  `Stop a running virtual machine on the specified node.`,
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

			task, err := vm.Stop(ctx)
			if err != nil {
				return fmt.Errorf("stop VM %d: %w", id, err)
			}
			if err := utility.WaitForTask(ctx, task, 10*time.Minute); err != nil {
				return fmt.Errorf("stop VM %d: %w", id, err)
			}

			fmt.Fprintf(out, "VM %d stopped successfully\n", id)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	return cmd
}
