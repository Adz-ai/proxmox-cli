package lxc

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

type containerInterfaceSummary struct {
	Name      string   `json:"name"`
	MAC       string   `json:"mac,omitempty"`
	Addresses []string `json:"addresses"`
}

func newIPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip",
		Short: "Show LXC container IP addresses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			interfaces, err := container.Interfaces(ctx)
			if err != nil {
				return fmt.Errorf("get network interfaces of container %d: %w", vmid, err)
			}

			summaries := make([]containerInterfaceSummary, 0, len(interfaces))
			for _, iface := range interfaces {
				addresses := []string{}
				if iface.Inet != "" {
					addresses = append(addresses, iface.Inet)
				}
				if iface.Inet6 != "" {
					addresses = append(addresses, iface.Inet6)
				}
				summaries = append(summaries, containerInterfaceSummary{
					Name:      iface.Name,
					MAC:       iface.HWAddr,
					Addresses: addresses,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintf(out, "Network interfaces of container %d:\n", vmid)
			fmt.Fprintf(out, "%-12s %-18s %s\n", "Name", "MAC", "Addresses")
			fmt.Fprintf(out, "%-12s %-18s %s\n", "----", "---", "---------")
			for _, summary := range summaries {
				fmt.Fprintf(out, "%-12s %-18s %s\n", summary.Name, summary.MAC, strings.Join(summary.Addresses, ", "))
			}
			if len(summaries) == 0 {
				fmt.Fprintln(out, "No interfaces reported")
			}
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}

func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show LXC container resource usage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
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

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			samples, err := container.RRDData(ctx, timeframe)
			if err != nil {
				return fmt.Errorf("get stats for container %d: %w", vmid, err)
			}

			summary := utility.SummarizeRRD(timeframeValue, samples)
			if format == "json" {
				return utility.PrintJSON(out, summary)
			}
			utility.PrintRRDSummary(out, fmt.Sprintf("container %d", vmid), summary)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	utility.AddTimeframeFlag(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}
