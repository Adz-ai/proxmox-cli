package lxc

import (
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

func newRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart an LXC container",
		Long:  `Reboot a running LXC container on the specified node.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			nodeName, vmid, err := containerTargetFromFlags(cmd)
			if err != nil {
				return err
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			container, err := node.Container(ctx, vmid)
			if err != nil {
				return fmt.Errorf("get container %d: %w", vmid, err)
			}

			task, err := container.Reboot(ctx)
			if err != nil {
				return fmt.Errorf("restart container %d: %w", vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("restart container %d: %w", vmid, err)
			}

			fmt.Fprintf(out, "Container %d restarted successfully\n", vmid)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	return cmd
}
