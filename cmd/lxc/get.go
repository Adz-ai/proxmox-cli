package lxc

import (
	"errors"
	"fmt"
	"io"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

type containerSummary struct {
	Node   string `json:"node"`
	VMID   uint64 `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Uptime uint64 `json:"uptime_seconds"`
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "List all LXC containers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}
			nodeFilter, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("get node flag: %w", err)
			}
			statusFilter, err := cmd.Flags().GetString("status")
			if err != nil {
				return fmt.Errorf("get status flag: %w", err)
			}
			ctx := cmd.Context()
			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			nodes, err := client.Nodes(ctx)
			if err != nil {
				return fmt.Errorf("list nodes: %w", err)
			}

			summaries := []containerSummary{}
			var nodeErrors []error
			for _, nodeStatus := range nodes {
				if nodeFilter != "" && nodeStatus.Node != nodeFilter {
					continue
				}
				node, err := client.Node(ctx, nodeStatus.Node)
				if err != nil {
					nodeErrors = append(nodeErrors, fmt.Errorf("get node %q: %w", nodeStatus.Node, err))
					continue
				}

				containers, err := node.Containers(ctx)
				if err != nil {
					nodeErrors = append(nodeErrors, fmt.Errorf("list containers on node %q: %w", nodeStatus.Node, err))
					continue
				}

				for _, container := range containers {
					if statusFilter != "" && container.Status != statusFilter {
						continue
					}
					summaries = append(summaries, containerSummary{
						Node:   nodeStatus.Node,
						VMID:   uint64(container.VMID),
						Name:   container.Name,
						Status: container.Status,
						Uptime: container.Uptime,
					})
				}
			}

			if format == "json" {
				if err := utility.PrintJSON(out, summaries); err != nil {
					return err
				}
			} else {
				printContainerTable(out, summaries)
			}
			if err := errors.Join(nodeErrors...); err != nil {
				return fmt.Errorf("list LXC containers: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Only list containers on this node")
	cmd.Flags().String("status", "", "Only list containers with this status")
	utility.RegisterNodeFlagCompletion(cmd, "node")
	utility.AddOutputFlag(cmd)
	return cmd
}

func printContainerTable(out io.Writer, summaries []containerSummary) {
	fmt.Fprintln(out, "LXC Containers:")
	fmt.Fprintln(out, "================")

	currentNode := ""
	for _, summary := range summaries {
		if summary.Node != currentNode {
			currentNode = summary.Node
			fmt.Fprintf(out, "\nNode: %s\n", summary.Node)
			fmt.Fprintf(out, "%-10s %-20s %-10s %-12s %-12s\n", "VMID", "Name", "Status", "Type", "Uptime")
			fmt.Fprintf(out, "%-10s %-20s %-10s %-12s %-12s\n", "----", "----", "------", "----", "------")
		}
		uptime := "N/A"
		if summary.Uptime > 0 {
			days := summary.Uptime / 86400
			hours := (summary.Uptime % 86400) / 3600
			uptime = fmt.Sprintf("%dd %dh", days, hours)
		}
		fmt.Fprintf(out, "%-10v %-20s %-10s %-12s %-12s\n",
			summary.VMID,
			summary.Name,
			summary.Status,
			"lxc",
			uptime)
	}

	if len(summaries) == 0 {
		fmt.Fprintln(out, "No LXC containers found in the cluster")
	}
}
