package vm

import (
	"context"
	"fmt"
	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete virtual machine",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			node, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("get node flag: %w", err)
			}
			id, err := cmd.Flags().GetInt("id")
			if err != nil {
				return fmt.Errorf("get id flag: %w", err)
			}
			node = strings.TrimSpace(node)
			if node == "" {
				return fmt.Errorf("validate node: node cannot be empty")
			}
			if id <= 0 {
				return fmt.Errorf("validate id: id must be positive")
			}

			if err := utility.ConfirmAction(cmd, fmt.Sprintf("Delete VM %d on node %q? This cannot be undone.", id, node)); err != nil {
				return err
			}

			if err := deleteVM(cmd.Context(), node, id, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("delete VM %d from node %q: %w", id, node, err)
			}

			fmt.Fprintf(out, "VM %d deleted successfully\n", id)
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "node to delete VM from (required)")
	cmd.Flags().IntP("id", "i", 0, "id for VM to delete (required)")
	if err := cmd.MarkFlagRequired("node"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(err)
	}
	utility.AddYesFlag(cmd)

	return cmd
}

func deleteVM(ctx context.Context, node string, id int, timeout time.Duration, progress io.Writer) error {
	client, err := utility.AuthenticatedClient()
	if err != nil {
		return fmt.Errorf("authenticate Proxmox client: %w", err)
	}

	retrievedNode, err := client.Node(ctx, node)
	if err != nil {
		return fmt.Errorf("get node %q: %w", node, err)
	}

	vmToDelete, err := retrievedNode.VirtualMachine(ctx, id)
	if err != nil {
		return fmt.Errorf("get VM %d: %w", id, err)
	}

	task, err := vmToDelete.Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("start delete task: %w", err)
	}
	if err := utility.WaitForTask(ctx, task, timeout, progress); err != nil {
		return fmt.Errorf("wait for delete task: %w", err)
	}

	return nil
}
