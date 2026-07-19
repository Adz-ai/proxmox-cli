package lxc

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

func newCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone an LXC container",
		Long:  `Create a copy of an existing LXC container with a new container ID.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			nodeName, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("read node flag: %w", err)
			}
			source, err := cmd.Flags().GetInt("source")
			if err != nil {
				return fmt.Errorf("read source flag: %w", err)
			}
			target, err := cmd.Flags().GetInt("target")
			if err != nil {
				return fmt.Errorf("read target flag: %w", err)
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return fmt.Errorf("read name flag: %w", err)
			}
			if err := validateContainerTarget(nodeName, source); err != nil {
				return err
			}
			if target < 0 {
				return fmt.Errorf("target container ID must be positive")
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			target, err = utility.ResolveVMID(ctx, client, target)
			if err != nil {
				return err
			}
			if target == source {
				return fmt.Errorf("target container ID must differ from the source")
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			container, err := node.Container(ctx, source)
			if err != nil {
				return fmt.Errorf("get container %d: %w", source, err)
			}

			options := &proxmox.ContainerCloneOptions{
				NewID:    target,
				Hostname: strings.TrimSpace(name),
			}
			_, task, err := container.Clone(ctx, options)
			if err != nil {
				return fmt.Errorf("clone container %d to %d: %w", source, target, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("clone container %d to %d: %w", source, target, err)
			}

			fmt.Fprintf(out, "Container %d cloned to %d successfully\n", source, target)
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node name")
	cmd.Flags().IntP("source", "s", 0, "Source container ID")
	cmd.Flags().IntP("target", "t", 0, "New container ID (omit to auto-assign the next free ID)")
	cmd.Flags().String("name", "", "Hostname for the new container")
	for _, flag := range []string{"node", "source"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
	return cmd
}
