package lxc

import (
	"fmt"
	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an LXC container",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			nodeName, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("read node flag: %w", err)
			}
			vmid, err := cmd.Flags().GetInt("vmid")
			if err != nil {
				return fmt.Errorf("read vmid flag: %w", err)
			}
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return fmt.Errorf("read force flag: %w", err)
			}
			if err := validateContainerTarget(nodeName, vmid); err != nil {
				return err
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			container, err := node.Container(ctx, vmid)
			if err != nil {
				return fmt.Errorf("get container %d: %w", vmid, err)
			}

			task, err := container.Delete(ctx, &proxmox.ContainerDeleteOptions{Force: proxmox.IntOrBool(force)})
			if err != nil {
				return fmt.Errorf("delete container %d: %w", vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, 10*time.Minute); err != nil {
				return fmt.Errorf("delete container %d: %w", vmid, err)
			}

			fmt.Fprintf(out, "Container %d deleted successfully\n", vmid)
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node name")
	cmd.Flags().IntP("vmid", "i", 0, "Container ID")
	cmd.Flags().BoolP("force", "f", false, "Force deletion")
	for _, flag := range []string{"node", "vmid"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	return cmd
}
