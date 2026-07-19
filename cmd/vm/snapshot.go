package vm

import (
	"fmt"
	"strings"
	"time"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/spf13/cobra"
)

func newSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage virtual machine snapshots",
		Long:  `Create and list snapshots of virtual machines.`,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newSnapshotCreateCmd(), newSnapshotListCmd())
	return cmd
}

func newSnapshotCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a snapshot of a virtual machine",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return fmt.Errorf("get name flag: %w", err)
			}
			name = strings.TrimSpace(name)
			if name == "" {
				return fmt.Errorf("snapshot name cannot be empty")
			}

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := vm.NewSnapshot(ctx, name)
			if err != nil {
				return fmt.Errorf("create snapshot %q for VM %d: %w", name, id, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("create snapshot %q for VM %d: %w", name, id, err)
			}

			fmt.Fprintf(out, "Snapshot %q created successfully for VM %d\n", name, id)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	cmd.Flags().String("name", "", "Snapshot name")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	return cmd
}

type vmSnapshotSummary struct {
	Name        string `json:"name"`
	CreatedAt   string `json:"created_at,omitempty"`
	Description string `json:"description,omitempty"`
}

func newSnapshotListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List snapshots of a virtual machine",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			snapshots, err := vm.Snapshots(ctx)
			if err != nil {
				return fmt.Errorf("list snapshots for VM %d: %w", id, err)
			}

			summaries := []vmSnapshotSummary{}
			for _, snapshot := range snapshots {
				// The API reports the live state as the pseudo-snapshot "current".
				if snapshot.Name == "current" {
					continue
				}
				created := ""
				if snapshot.Snaptime > 0 {
					created = time.Unix(snapshot.Snaptime, 0).UTC().Format(time.RFC3339)
				}
				summaries = append(summaries, vmSnapshotSummary{
					Name:        snapshot.Name,
					CreatedAt:   created,
					Description: snapshot.Description,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintf(out, "Snapshots for VM %d:\n", id)
			fmt.Fprintln(out, "====================")
			for _, summary := range summaries {
				created := summary.CreatedAt
				if created == "" {
					created = "N/A"
				}
				fmt.Fprintf(out, "%-20s %-22s %s\n", summary.Name, created, summary.Description)
			}
			if len(summaries) == 0 {
				fmt.Fprintln(out, "No snapshots found")
			}
			return nil
		},
	}

	addVMTargetFlags(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}

// vmFromFlags resolves the node/id flags to a virtual machine after
// authenticating, shared by the snapshot subcommands.
func vmFromFlags(cmd *cobra.Command) (interfaces.VirtualMachineInterface, int, error) {
	node, id, err := vmTargetFromFlags(cmd)
	if err != nil {
		return nil, 0, err
	}

	client, err := utility.AuthenticatedClient()
	if err != nil {
		return nil, 0, fmt.Errorf("authenticate Proxmox client: %w", err)
	}

	retrievedNode, err := client.Node(cmd.Context(), node)
	if err != nil {
		return nil, 0, fmt.Errorf("get node %q: %w", node, err)
	}

	vm, err := retrievedNode.VirtualMachine(cmd.Context(), id)
	if err != nil {
		return nil, 0, fmt.Errorf("get VM %d: %w", id, err)
	}
	return vm, id, nil
}
