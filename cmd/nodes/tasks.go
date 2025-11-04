package nodes

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks on a node",
	Long:  `List all tasks running or completed on a specific node in the Proxmox cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		limit, _ := cmd.Flags().GetInt("limit")

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

		// Note: Task listing requires additional API endpoints
		// that are not yet fully implemented in go-proxmox
		fmt.Fprintf(out, "Task listing for node %s:\n", nodeName)
		fmt.Fprintln(out, "This feature requires direct Proxmox API access.")
		fmt.Fprintln(out, "Use: pvesh get /nodes/<node>/tasks (on the Proxmox node)")
		fmt.Fprintln(out, "Or access via Proxmox web interface: <Node> > Task History")

		// Suppress unused variable warning
		_ = limit
	},
}

func init() {
	tasksCmd.Flags().StringP("node", "n", "", "Node name")
	tasksCmd.Flags().IntP("limit", "l", 20, "Limit number of tasks to display")
	tasksCmd.MarkFlagRequired("node")
}
