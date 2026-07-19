package nodes

import (
	"fmt"
	"strings"
	"time"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

type taskSummary struct {
	UPID      string `json:"upid"`
	Type      string `json:"type"`
	ID        string `json:"id,omitempty"`
	User      string `json:"user"`
	Status    string `json:"status"`
	Running   bool   `json:"running"`
	StartedAt string `json:"started_at,omitempty"`
	EndedAt   string `json:"ended_at,omitempty"`
}

func newTasksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "List tasks on a node",
		Long:  `List running and recently completed tasks on a specific node in the Proxmox cluster.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			nodeName, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("get node flag: %w", err)
			}
			nodeName = strings.TrimSpace(nodeName)
			if nodeName == "" {
				return fmt.Errorf("node cannot be empty")
			}
			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return fmt.Errorf("get limit flag: %w", err)
			}
			if limit <= 0 {
				return fmt.Errorf("limit must be positive")
			}
			runningOnly, err := cmd.Flags().GetBool("running")
			if err != nil {
				return fmt.Errorf("get running flag: %w", err)
			}
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}
			ctx := cmd.Context()

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			options := &proxmox.NodeTasksOptions{Limit: limit, Source: "all"}
			if runningOnly {
				options.Source = "active"
			}
			tasks, err := node.Tasks(ctx, options)
			if err != nil {
				return fmt.Errorf("list tasks on node %q: %w", nodeName, err)
			}

			summaries := make([]taskSummary, 0, len(tasks))
			for _, task := range tasks {
				summary := taskSummary{
					UPID:    string(task.UPID),
					Type:    task.Type,
					ID:      task.ID,
					User:    task.User,
					Status:  task.Status,
					Running: task.IsRunning,
				}
				if !task.StartTime.IsZero() {
					summary.StartedAt = task.StartTime.UTC().Format(time.RFC3339)
				}
				if !task.EndTime.IsZero() {
					summary.EndedAt = task.EndTime.UTC().Format(time.RFC3339)
				}
				summaries = append(summaries, summary)
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintf(out, "Tasks on node %s:\n", nodeName)
			fmt.Fprintln(out, "==================")
			fmt.Fprintf(out, "%-16s %-10s %-18s %-10s %-22s %s\n", "Type", "ID", "User", "Status", "Started", "Ended")
			fmt.Fprintf(out, "%-16s %-10s %-18s %-10s %-22s %s\n", "----", "--", "----", "------", "-------", "-----")
			for _, summary := range summaries {
				status := summary.Status
				if summary.Running {
					status = "running"
				}
				started := summary.StartedAt
				if started == "" {
					started = "N/A"
				}
				ended := summary.EndedAt
				if ended == "" {
					ended = "-"
				}
				fmt.Fprintf(out, "%-16s %-10s %-18s %-10s %-22s %s\n",
					summary.Type, summary.ID, summary.User, status, started, ended)
			}
			if len(summaries) == 0 {
				fmt.Fprintln(out, "No tasks found")
			}
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node name")
	cmd.Flags().IntP("limit", "l", 20, "Maximum number of tasks to list")
	cmd.Flags().BoolP("running", "r", false, "Show only running tasks")
	if err := cmd.MarkFlagRequired("node"); err != nil {
		panic(err)
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
	utility.AddOutputFlag(cmd)
	return cmd
}
