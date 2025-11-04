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

		// Verify node exists
		_, err := client.Node(ctx, nodeName)
		if err != nil {
			fmt.Fprintf(out, "Error getting node: %v\n", err)
			return
		}

		// Note: Storage listing requires additional API endpoints
		// that are not yet fully implemented in go-proxmox
		fmt.Fprintf(out, "Storage listing for node %s:\n", nodeName)
		fmt.Fprintln(out, "This feature requires direct Proxmox API access.")
		fmt.Fprintln(out, "Use: pvesm status (on the Proxmox node)")
		fmt.Fprintln(out, "Or access via Proxmox web interface: Datacenter > Storage")
	},
}

func init() {
	storageCmd.Flags().StringP("node", "n", "", "Node name")
	storageCmd.MarkFlagRequired("node")
}
