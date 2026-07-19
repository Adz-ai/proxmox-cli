package lxc

import (
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe an LXC container",
		Long:  `Display detailed information about an LXC container on the specified node.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			nodeName, vmid, err := containerTargetFromFlags(cmd)
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

			container, err := node.Container(ctx, vmid)
			if err != nil {
				return fmt.Errorf("get container %d: %w", vmid, err)
			}
			details := container.Details()

			fmt.Fprintln(out, "Container Details")
			fmt.Fprintln(out, "=================")
			fmt.Fprintf(out, "ID: %d\n", vmid)
			fmt.Fprintf(out, "Name: %s\n", details.Name)
			fmt.Fprintf(out, "Node: %s\n", details.Node)
			fmt.Fprintf(out, "Status: %s\n", details.Status)
			if details.Tags != "" {
				fmt.Fprintf(out, "Tags: %s\n", details.Tags)
			}
			fmt.Fprintf(out, "Uptime: %s\n", formatUptime(details.Uptime))
			fmt.Fprintf(out, "CPUs: %d\n", details.CPUs)
			fmt.Fprintf(out, "Memory: %s\n", formatBytes(details.MaxMemory))
			fmt.Fprintf(out, "Swap: %s\n", formatBytes(details.MaxSwap))
			fmt.Fprintf(out, "Disk: %s\n", formatBytes(details.MaxDisk))
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	return cmd
}

func formatBytes(bytes uint64) string {
	const gibibyte = 1024 * 1024 * 1024
	return fmt.Sprintf("%.2f GiB", float64(bytes)/gibibyte)
}

func formatUptime(uptime uint64) string {
	if uptime == 0 {
		return "N/A"
	}
	days := uptime / 86400
	hours := (uptime % 86400) / 3600
	minutes := (uptime % 3600) / 60
	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}
