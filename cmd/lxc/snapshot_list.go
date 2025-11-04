package lxc

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List snapshots of an LXC container",
	Long:  `List all snapshots for a specific LXC container.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		vmid, _ := cmd.Flags().GetInt("vmid")

		if nodeName == "" || vmid == 0 {
			fmt.Fprintln(out, "Error: node and vmid are required")
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		node, err := client.Node(ctx, nodeName)
		if err != nil {
			fmt.Fprintf(out, "Error getting node: %v\n", err)
			return
		}

		container, err := node.Container(ctx, vmid)
		if err != nil {
			fmt.Fprintf(out, "Error getting container: %v\n", err)
			return
		}

		// List snapshots
		snapshots, err := container.Snapshots(ctx)
		if err != nil {
			fmt.Fprintf(out, "Error listing snapshots: %v\n", err)
			return
		}

		if len(snapshots) == 0 {
			fmt.Fprintf(out, "No snapshots found for container %d\n", vmid)
			return
		}

		fmt.Fprintf(out, "Snapshots for container %d:\n", vmid)
		fmt.Fprintf(out, "%-20s %-15s %-s\n", "NAME", "SNAPTIME", "DESCRIPTION")
		fmt.Fprintln(out, "--------------------------------------------------------------------------------")
		for _, snapshot := range snapshots {
			snapTime := ""
			if snapshot.SnapTime > 0 {
				snapTime = fmt.Sprintf("%d", snapshot.SnapTime)
			}
			description := snapshot.Description
			if description == "" {
				description = "-"
			}
			fmt.Fprintf(out, "%-20s %-15s %-s\n", snapshot.Name, snapTime, description)
		}
	},
}

func init() {
	snapshotListCmd.Flags().StringP("node", "n", "", "Node name")
	snapshotListCmd.Flags().IntP("vmid", "i", 0, "Container ID")
	snapshotListCmd.MarkFlagRequired("node")
	snapshotListCmd.MarkFlagRequired("vmid")
}
