package nodes

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "List storage on a node",
	Long:  `List all storage available on a specific node in the Proxmox cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		if nodeName == "" {
			fmt.Fprintln(out, "Error: node name is required")
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		node, err := client.Node(ctx, nodeName)
		if err != nil {
			fmt.Fprintf(out, "Error getting node: %v\n", err)
			return
		}

		storages, err := node.Storages(ctx)
		if err != nil {
			fmt.Fprintf(out, "Error getting storage: %v\n", err)
			return
		}

		if len(storages) == 0 {
			fmt.Fprintf(out, "No storage found on node %s\n", nodeName)
			return
		}

		fmt.Fprintf(out, "Storage on node %s:\n", nodeName)
		fmt.Fprintf(out, "%-20s %-15s %-10s %-15s %-15s %-10s\n",
			"NAME", "TYPE", "ACTIVE", "TOTAL", "USED", "AVAIL")
		fmt.Fprintln(out, "--------------------------------------------------------------------------------")

		for _, storage := range storages {
			active := "no"
			if storage.Active == 1 {
				active = "yes"
			}

			totalGB := float64(storage.Total) / (1024 * 1024 * 1024)
			usedGB := float64(storage.Used) / (1024 * 1024 * 1024)
			availGB := float64(storage.Avail) / (1024 * 1024 * 1024)

			fmt.Fprintf(out, "%-20s %-15s %-10s %13.2f GB %13.2f GB %9.2f GB\n",
				storage.Storage, storage.Type, active, totalGB, usedGB, availGB)
		}
	},
}

func init() {
	storageCmd.Flags().StringP("node", "n", "", "Node name")
	storageCmd.MarkFlagRequired("node")
}
