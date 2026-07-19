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

	cmd.AddCommand(newSnapshotCreateCmd(), newSnapshotListCmd())
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
			if err := utility.WaitForTask(ctx, task, 10*time.Minute); err != nil {
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

func newSnapshotListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List snapshots of an LXC container",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			snapshots, err := container.Snapshots(ctx)
			if err != nil {
				return fmt.Errorf("list snapshots for container %d: %w", vmid, err)
			}

			fmt.Fprintf(out, "Snapshots for container %d:\n", vmid)
			fmt.Fprintln(out, "===========================")
			listed := 0
			for _, snapshot := range snapshots {
				// The API reports the live state as the pseudo-snapshot "current".
				if snapshot.Name == "current" {
					continue
				}
				created := "N/A"
				if snapshot.SnapshotCreationTime > 0 {
					created = time.Unix(snapshot.SnapshotCreationTime, 0).UTC().Format(time.RFC3339)
				}
				fmt.Fprintf(out, "%-20s %-22s %s\n", snapshot.Name, created, snapshot.Description)
				listed++
			}
			if listed == 0 {
				fmt.Fprintln(out, "No snapshots found")
			}
			return nil
		},
	}

	addContainerTargetFlags(cmd)
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
