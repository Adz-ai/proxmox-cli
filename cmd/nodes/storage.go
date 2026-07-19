package nodes

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

type storageSummary struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Content   string  `json:"content"`
	Active    bool    `json:"active"`
	Enabled   bool    `json:"enabled"`
	Shared    bool    `json:"shared"`
	Used      uint64  `json:"used_bytes"`
	Available uint64  `json:"available_bytes"`
	Total     uint64  `json:"total_bytes"`
	UsedPct   float64 `json:"used_percent"`
}

func newStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "List storage on a node",
		Long:  `List all storage available on a specific node in the Proxmox cluster.`,
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

			storages, err := node.Storages(ctx)
			if err != nil {
				return fmt.Errorf("list storage on node %q: %w", nodeName, err)
			}

			summaries := make([]storageSummary, 0, len(storages))
			for _, storage := range storages {
				summaries = append(summaries, storageSummary{
					Name:      storage.Name,
					Type:      storage.Type,
					Content:   storage.Content,
					Active:    storage.Active == 1,
					Enabled:   storage.Enabled == 1,
					Shared:    storage.Shared == 1,
					Used:      storage.Used,
					Available: storage.Avail,
					Total:     storage.Total,
					UsedPct:   storage.UsedFraction * 100,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintf(out, "Storage on node %s:\n", nodeName)
			fmt.Fprintln(out, "====================")
			fmt.Fprintf(out, "%-15s %-10s %-8s %-24s %s\n", "Name", "Type", "Active", "Usage", "Content")
			fmt.Fprintf(out, "%-15s %-10s %-8s %-24s %s\n", "----", "----", "------", "-----", "-------")
			for _, summary := range summaries {
				active := "no"
				if summary.Active {
					active = "yes"
				}
				usage := fmt.Sprintf("%s / %s (%.1f%%)",
					formatStorageBytes(summary.Used), formatStorageBytes(summary.Total), summary.UsedPct)
				fmt.Fprintf(out, "%-15s %-10s %-8s %-24s %s\n",
					summary.Name, summary.Type, active, usage, summary.Content)
			}
			if len(summaries) == 0 {
				fmt.Fprintln(out, "No storage found on this node")
			}
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node name")
	if err := cmd.MarkFlagRequired("node"); err != nil {
		panic(err)
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
	utility.AddOutputFlag(cmd)
	return cmd
}

func formatStorageBytes(bytes uint64) string {
	const gibibyte = 1024 * 1024 * 1024
	return fmt.Sprintf("%.1f GiB", float64(bytes)/gibibyte)
}
