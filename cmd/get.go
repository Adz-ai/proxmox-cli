package cmd

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

type resourceSummary struct {
	Type   string `json:"type"`
	VMID   uint64 `json:"vmid,omitempty"`
	Name   string `json:"name"`
	Node   string `json:"node"`
	Status string `json:"status"`
	Uptime uint64 `json:"uptime_seconds,omitempty"`
}

func newResourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "List cluster resources (VMs, containers, and storage)",
		Long: `Display every VM, LXC container, and storage across the cluster in a
single view, using one API call instead of querying each node.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}
			typeFilter, err := cmd.Flags().GetString("type")
			if err != nil {
				return fmt.Errorf("get type flag: %w", err)
			}
			typeFilter = strings.ToLower(strings.TrimSpace(typeFilter))
			switch typeFilter {
			case "", "vm", "lxc", "storage":
			default:
				return fmt.Errorf("unsupported type %q; use vm, lxc, or storage", typeFilter)
			}
			nodeFilter, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("get node flag: %w", err)
			}
			statusFilter, err := cmd.Flags().GetString("status")
			if err != nil {
				return fmt.Errorf("get status flag: %w", err)
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}
			ctx := cmd.Context()

			cluster, err := client.Cluster(ctx)
			if err != nil {
				return fmt.Errorf("get cluster: %w", err)
			}

			// The API-side filter narrows the payload; "vm" covers both QEMU
			// and LXC guests.
			var apiFilters []string
			switch typeFilter {
			case "vm", "lxc":
				apiFilters = []string{"vm"}
			case "storage":
				apiFilters = []string{"storage"}
			}
			resources, err := cluster.Resources(ctx, apiFilters...)
			if err != nil {
				return fmt.Errorf("list cluster resources: %w", err)
			}

			summaries := []resourceSummary{}
			for _, resource := range resources {
				kind := resource.Type
				if kind == "qemu" {
					kind = "vm"
				}
				if kind != "vm" && kind != "lxc" && kind != "storage" {
					continue
				}
				if typeFilter != "" && kind != typeFilter {
					continue
				}
				if nodeFilter != "" && resource.Node != nodeFilter {
					continue
				}
				if statusFilter != "" && resource.Status != statusFilter {
					continue
				}
				name := resource.Name
				if kind == "storage" {
					name = resource.Storage
				}
				summaries = append(summaries, resourceSummary{
					Type:   kind,
					VMID:   resource.VMID,
					Name:   name,
					Node:   resource.Node,
					Status: resource.Status,
					Uptime: resource.Uptime,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintln(out, "Cluster resources:")
			fmt.Fprintln(out, "==================")
			fmt.Fprintf(out, "%-8s %-8s %-24s %-15s %-10s %s\n", "Type", "VMID", "Name", "Node", "Status", "Uptime")
			fmt.Fprintf(out, "%-8s %-8s %-24s %-15s %-10s %s\n", "----", "----", "----", "----", "------", "------")
			for _, summary := range summaries {
				vmid := "-"
				if summary.VMID > 0 {
					vmid = fmt.Sprintf("%d", summary.VMID)
				}
				uptime := "-"
				if summary.Uptime > 0 {
					days := summary.Uptime / 86400
					hours := (summary.Uptime % 86400) / 3600
					uptime = fmt.Sprintf("%dd %dh", days, hours)
				}
				fmt.Fprintf(out, "%-8s %-8s %-24s %-15s %-10s %s\n",
					summary.Type, vmid, summary.Name, summary.Node, summary.Status, uptime)
			}
			if len(summaries) == 0 {
				fmt.Fprintln(out, "No resources found")
			}
			return nil
		},
	}

	cmd.Flags().String("type", "", "Only list resources of this type: vm, lxc, or storage")
	cmd.Flags().StringP("node", "n", "", "Only list resources on this node")
	cmd.Flags().String("status", "", "Only list resources with this status")
	_ = cmd.RegisterFlagCompletionFunc("type", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"vm", "lxc", "storage"}, cobra.ShellCompDirectiveNoFileComp
	})
	utility.RegisterNodeFlagCompletion(cmd, "node")
	utility.AddOutputFlag(cmd)
	return cmd
}
