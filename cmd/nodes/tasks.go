package nodes

import (
	"context"
	"fmt"
	"proxmox-cli/cmd/utility"
	"time"

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

		node, err := client.Node(ctx, nodeName)
		if err != nil {
			fmt.Fprintf(out, "Error getting node: %v\n", err)
			return
		}

		tasks, err := node.Tasks(ctx)
		if err != nil {
			fmt.Fprintf(out, "Error getting tasks: %v\n", err)
			return
		}

		if len(tasks) == 0 {
			fmt.Fprintf(out, "No tasks found on node %s\n", nodeName)
			return
		}

		// Limit the number of tasks displayed
		displayCount := len(tasks)
		if limit > 0 && limit < displayCount {
			displayCount = limit
		}

		fmt.Fprintf(out, "Recent tasks on node %s (showing %d):\n", nodeName, displayCount)
		fmt.Fprintf(out, "%-12s %-20s %-10s %-15s %-s\n",
			"UPID", "TYPE", "STATUS", "USER", "START TIME")
		fmt.Fprintln(out, "--------------------------------------------------------------------------------")

		for i := 0; i < displayCount; i++ {
			task := tasks[i]

			// Format start time
			startTime := time.Unix(int64(task.StartTime), 0).Format("2006-01-02 15:04")

			// Truncate UPID for display
			upid := task.UPID
			if len(upid) > 12 {
				upid = upid[:12] + "..."
			}

			fmt.Fprintf(out, "%-15s %-20s %-10s %-15s %-s\n",
				upid, task.Type, task.Status, task.User, startTime)
		}

		if limit > 0 && len(tasks) > limit {
			fmt.Fprintf(out, "\n(%d more tasks not shown, increase --limit to see more)\n", len(tasks)-limit)
		}
	},
}

func init() {
	tasksCmd.Flags().StringP("node", "n", "", "Node name")
	tasksCmd.Flags().IntP("limit", "l", 20, "Limit number of tasks to display")
	tasksCmd.MarkFlagRequired("node")
}
