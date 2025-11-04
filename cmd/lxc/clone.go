package lxc

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone an LXC container",
	Long:  `Clone an existing LXC container to create a new container with a different ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		sourceID, _ := cmd.Flags().GetInt("source")
		targetID, _ := cmd.Flags().GetInt("target")
		newName, _ := cmd.Flags().GetString("name")
		fullClone, _ := cmd.Flags().GetBool("full")

		if nodeName == "" || sourceID == 0 || targetID == 0 {
			fmt.Fprintln(out, "Error: node, source, and target are required")
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		node, err := client.Node(ctx, nodeName)
		if err != nil {
			fmt.Fprintf(out, "Error getting node: %v\n", err)
			return
		}

		container, err := node.Container(ctx, sourceID)
		if err != nil {
			fmt.Fprintf(out, "Error getting source container: %v\n", err)
			return
		}

		// Build clone options
		cloneOptions := &struct {
			NewID    int
			Hostname string
			Full     int
		}{
			NewID:    targetID,
			Hostname: newName,
			Full:     0,
		}

		if fullClone {
			cloneOptions.Full = 1
		}

		task, err := container.Clone(ctx, cloneOptions)
		if err != nil {
			fmt.Fprintf(out, "Error cloning container: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Clone task started: %s\n", task.UPID)
		fmt.Fprintf(out, "Container %d cloned to %d successfully\n", sourceID, targetID)
		if newName != "" {
			fmt.Fprintf(out, "New container name: %s\n", newName)
		}
	},
}

func init() {
	cloneCmd.Flags().StringP("node", "n", "", "Node name")
	cloneCmd.Flags().IntP("source", "s", 0, "Source container ID")
	cloneCmd.Flags().IntP("target", "t", 0, "Target container ID")
	cloneCmd.Flags().String("name", "", "Name for the cloned container")
	cloneCmd.Flags().Bool("full", false, "Create a full clone instead of linked clone")
	cloneCmd.MarkFlagRequired("node")
	cloneCmd.MarkFlagRequired("source")
	cloneCmd.MarkFlagRequired("target")
}
