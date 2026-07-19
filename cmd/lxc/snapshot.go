package lxc

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
		Short: "Manage LXC container snapshots",
		Long:  `Create and list snapshots of LXC containers.`,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newSnapshotCreateCmd(), newSnapshotListCmd(), newSnapshotRollbackCmd(), newSnapshotDeleteCmd())
	return cmd
}

func snapshotNameFromFlags(cmd *cobra.Command) (string, error) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return "", fmt.Errorf("get name flag: %w", err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("snapshot name cannot be empty")
	}
	return name, nil
}

func newSnapshotRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Roll back an LXC container to a snapshot",
		Long:  `Restore a container to the state captured in the named snapshot. Changes made since the snapshot are lost.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			name, err := snapshotNameFromFlags(cmd)
			if err != nil {
				return err
			}
			start, err := cmd.Flags().GetBool("start")
			if err != nil {
				return fmt.Errorf("read start flag: %w", err)
			}
			_, targetID, err := containerTargetFromFlags(cmd)
			if err != nil {
				return err
			}
			if err := utility.ConfirmAction(cmd, fmt.Sprintf("Roll back container %d to snapshot %q? Changes made since the snapshot will be lost.", targetID, name)); err != nil {
				return err
			}

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := container.RollbackSnapshot(ctx, name, start)
			if err != nil {
				return fmt.Errorf("roll back container %d to snapshot %q: %w", vmid, name, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("roll back container %d to snapshot %q: %w", vmid, name, err)
			}

			fmt.Fprintf(out, "Container %d rolled back to snapshot %q successfully\n", vmid, name)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	cmd.Flags().String("name", "", "Snapshot name")
	cmd.Flags().Bool("start", false, "Start the container after the rollback")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	utility.AddYesFlag(cmd)
	return cmd
}

func newSnapshotDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a snapshot of an LXC container",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			name, err := snapshotNameFromFlags(cmd)
			if err != nil {
				return err
			}
			_, targetID, err := containerTargetFromFlags(cmd)
			if err != nil {
				return err
			}
			if err := utility.ConfirmAction(cmd, fmt.Sprintf("Delete snapshot %q of container %d?", name, targetID)); err != nil {
				return err
			}

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := container.DeleteSnapshot(ctx, name)
			if err != nil {
				return fmt.Errorf("delete snapshot %q of container %d: %w", name, vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("delete snapshot %q of container %d: %w", name, vmid, err)
			}

			fmt.Fprintf(out, "Snapshot %q of container %d deleted successfully\n", name, vmid)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	cmd.Flags().String("name", "", "Snapshot name")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	utility.AddYesFlag(cmd)
	return cmd
}

func newSnapshotCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a snapshot of an LXC container",
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

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := container.NewSnapshot(ctx, name)
			if err != nil {
				return fmt.Errorf("create snapshot %q for container %d: %w", name, vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("create snapshot %q for container %d: %w", name, vmid, err)
			}

			fmt.Fprintf(out, "Snapshot %q created successfully for container %d\n", name, vmid)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	cmd.Flags().String("name", "", "Snapshot name")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	return cmd
}

type snapshotSummary struct {
	Name        string `json:"name"`
	CreatedAt   string `json:"created_at,omitempty"`
	Description string `json:"description,omitempty"`
}

func newSnapshotListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List snapshots of an LXC container",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			snapshots, err := container.Snapshots(ctx)
			if err != nil {
				return fmt.Errorf("list snapshots for container %d: %w", vmid, err)
			}

			summaries := []snapshotSummary{}
			for _, snapshot := range snapshots {
				// The API reports the live state as the pseudo-snapshot "current".
				if snapshot.Name == "current" {
					continue
				}
				created := ""
				if snapshot.SnapshotCreationTime > 0 {
					created = time.Unix(snapshot.SnapshotCreationTime, 0).UTC().Format(time.RFC3339)
				}
				summaries = append(summaries, snapshotSummary{
					Name:        snapshot.Name,
					CreatedAt:   created,
					Description: snapshot.Description,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintf(out, "Snapshots for container %d:\n", vmid)
			fmt.Fprintln(out, "===========================")
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

	addContainerTargetFlags(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}

// containerFromFlags resolves the node/vmid flags to a container after
// authenticating, shared by the snapshot subcommands.
func containerFromFlags(cmd *cobra.Command) (interfaces.ContainerInterface, int, error) {
	nodeName, vmid, err := containerTargetFromFlags(cmd)
	if err != nil {
		return nil, 0, err
	}

	client, err := utility.AuthenticatedClient()
	if err != nil {
		return nil, 0, fmt.Errorf("authenticate Proxmox client: %w", err)
	}

	node, err := client.Node(cmd.Context(), nodeName)
	if err != nil {
		return nil, 0, fmt.Errorf("get node %q: %w", nodeName, err)
	}

	container, err := node.Container(cmd.Context(), vmid)
	if err != nil {
		return nil, 0, fmt.Errorf("get container %d: %w", vmid, err)
	}
	return container, vmid, nil
}
