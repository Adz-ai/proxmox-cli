package vm

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
		Short: "Clone a virtual machine",
		Long:  `Create a copy of an existing virtual machine with a new VM ID.`,
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
			full, err := cmd.Flags().GetBool("full")
			if err != nil {
				return fmt.Errorf("read full flag: %w", err)
			}
			storage, err := cmd.Flags().GetString("storage")
			if err != nil {
				return fmt.Errorf("read storage flag: %w", err)
			}
			if strings.TrimSpace(nodeName) == "" {
				return fmt.Errorf("node cannot be empty")
			}
			if source <= 0 {
				return fmt.Errorf("source VM ID must be positive")
			}
			if target < 0 {
				return fmt.Errorf("target VM ID must be positive")
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
				return fmt.Errorf("target VM ID must differ from the source")
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			vm, err := node.VirtualMachine(ctx, source)
			if err != nil {
				return fmt.Errorf("get VM %d: %w", source, err)
			}

			options := &proxmox.VirtualMachineCloneOptions{
				NewID:   target,
				Name:    strings.TrimSpace(name),
				Full:    proxmox.IntOrBool(full),
				Storage: strings.TrimSpace(storage),
			}
			_, task, err := vm.Clone(ctx, options)
			if err != nil {
				return fmt.Errorf("clone VM %d to %d: %w", source, target, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("clone VM %d to %d: %w", source, target, err)
			}

			fmt.Fprintf(out, "VM %d cloned to %d successfully\n", source, target)
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node name")
	cmd.Flags().IntP("source", "s", 0, "Source VM ID")
	cmd.Flags().IntP("target", "t", 0, "New VM ID (omit to auto-assign the next free ID)")
	cmd.Flags().String("name", "", "Name for the new VM")
	cmd.Flags().Bool("full", false, "Create a full copy instead of a linked clone")
	cmd.Flags().String("storage", "", "Target storage for a full clone")
	for _, flag := range []string{"node", "source"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
	return cmd
}
