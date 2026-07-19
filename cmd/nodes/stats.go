package nodes

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show node resource usage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			nodeName, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("read node flag: %w", err)
			}
			nodeName = strings.TrimSpace(nodeName)
			if nodeName == "" {
				return fmt.Errorf("node cannot be empty")
			}
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}
			timeframeValue, err := cmd.Flags().GetString("timeframe")
			if err != nil {
				return fmt.Errorf("read timeframe flag: %w", err)
			}
			timeframe, err := utility.ParseTimeframe(timeframeValue)
			if err != nil {
				return err
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			samples, err := node.RRDData(ctx, timeframe, proxmox.AVERAGE)
			if err != nil {
				return fmt.Errorf("get stats for node %q: %w", nodeName, err)
			}

			summary := utility.SummarizeRRD(timeframeValue, samples)
			if format == "json" {
				return utility.PrintJSON(out, summary)
			}
			utility.PrintRRDSummary(out, fmt.Sprintf("node %s", nodeName), summary)
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node name")
	if err := cmd.MarkFlagRequired("node"); err != nil {
		panic(err)
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
	utility.AddTimeframeFlag(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}
