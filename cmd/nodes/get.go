package nodes

import (
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

type nodeSummary struct {
	Node   string `json:"node"`
	Status string `json:"status"`
	Type   string `json:"type"`
	Uptime uint64 `json:"uptime_seconds"`
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "List all nodes in the cluster",
		Long:  `Display a list of all nodes in the Proxmox cluster with their status and uptime.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			nodes, err := client.Nodes(cmd.Context())
			if err != nil {
				return fmt.Errorf("fetch cluster nodes: %w", err)
			}

			summaries := make([]nodeSummary, 0, len(nodes))
			for _, node := range nodes {
				summaries = append(summaries, nodeSummary{
					Node:   node.Node,
					Status: node.Status,
					Type:   node.Type,
					Uptime: node.Uptime,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintln(out, "Nodes in cluster:")
			fmt.Fprintln(out, "=================")
			fmt.Fprintf(out, "%-15s %-10s %-8s %-12s\n", "Node", "Status", "Type", "Uptime")
			fmt.Fprintf(out, "%-15s %-10s %-8s %-12s\n", "----", "------", "----", "------")
			for _, summary := range summaries {
				uptime := "N/A"
				if summary.Uptime > 0 {
					days := summary.Uptime / 86400
					hours := (summary.Uptime % 86400) / 3600
					uptime = fmt.Sprintf("%dd %dh", days, hours)
				}
				fmt.Fprintf(out, "%-15s %-10s %-8s %-12s\n",
					summary.Node,
					summary.Status,
					summary.Type,
					uptime)
			}
			return nil
		},
	}

	utility.AddOutputFlag(cmd)
	return cmd
}
