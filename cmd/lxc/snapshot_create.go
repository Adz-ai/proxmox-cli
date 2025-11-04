package lxc

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var snapshotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a snapshot of an LXC container",
	Long:  `Create a snapshot of an LXC container with a specified name.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		vmid, _ := cmd.Flags().GetInt("vmid")
		snapshotName, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		if nodeName == "" || vmid == 0 || snapshotName == "" {
			fmt.Fprintln(out, "Error: node, vmid, and name are required")
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

		// Create snapshot
		task, err := container.Snapshot(ctx, snapshotName)
		if err != nil {
			fmt.Fprintf(out, "Error creating snapshot: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Snapshot task started: %s\n", task.UPID)
		fmt.Fprintf(out, "Snapshot '%s' created successfully for container %d\n", snapshotName, vmid)
		if description != "" {
			fmt.Fprintf(out, "Note: Description '%s' may need to be set manually\n", description)
		}
	},
}

func init() {
	snapshotCreateCmd.Flags().StringP("node", "n", "", "Node name")
	snapshotCreateCmd.Flags().IntP("vmid", "i", 0, "Container ID")
	snapshotCreateCmd.Flags().String("name", "", "Snapshot name")
	snapshotCreateCmd.Flags().String("description", "", "Snapshot description")
	snapshotCreateCmd.MarkFlagRequired("node")
	snapshotCreateCmd.MarkFlagRequired("vmid")
	snapshotCreateCmd.MarkFlagRequired("name")
}
