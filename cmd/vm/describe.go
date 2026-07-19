package vm

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/spf13/cobra"
)

func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a virtual machine",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			node, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("get node flag: %w", err)
			}
			vmID, err := cmd.Flags().GetInt("id")
			if err != nil {
				return fmt.Errorf("get id flag: %w", err)
			}
			node = strings.TrimSpace(node)
			if node == "" {
				return fmt.Errorf("validate node: node cannot be empty")
			}
			if vmID <= 0 {
				return fmt.Errorf("validate id: id must be positive")
			}
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			if err := describeVirtualMachine(cmd.Context(), out, node, vmID, format); err != nil {
				return fmt.Errorf("describe virtual machine %d on node %q: %w", vmID, node, err)
			}
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node name")
	cmd.Flags().IntP("id", "i", 0, "VM ID")
	if err := cmd.MarkFlagRequired("node"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(err)
	}
	utility.AddOutputFlag(cmd)

	return cmd
}

func describeVirtualMachine(ctx context.Context, out io.Writer, node string, vmID int, format string) error {
	client, err := utility.AuthenticatedClient()
	if err != nil {
		return fmt.Errorf("authenticate Proxmox client: %w", err)
	}

	retrievedNode, err := client.Node(ctx, node)
	if err != nil {
		return fmt.Errorf("get node %q: %w", node, err)
	}

	vm, err := retrievedNode.VirtualMachine(ctx, vmID)
	if err != nil {
		return fmt.Errorf("get VM %d: %w", vmID, err)
	}
	details := vm.Details()

	if format == "json" {
		return utility.PrintJSON(out, struct {
			VMID int `json:"vmid"`
			interfaces.VirtualMachineDetails
		}{vmID, details})
	}

	fmt.Fprintln(out, "VM Details")
	fmt.Fprintln(out, "==========")
	fmt.Fprintf(out, "ID: %d\n", vmID)
	fmt.Fprintf(out, "Name: %s\n", details.Name)
	fmt.Fprintf(out, "Node: %s\n", details.Node)
	fmt.Fprintf(out, "Status: %s\n", details.Status)
	if details.Tags != "" {
		fmt.Fprintf(out, "Tags: %s\n", details.Tags)
	}
	fmt.Fprintf(out, "Uptime: %s\n", formatUptime(details.Uptime))
	fmt.Fprintf(out, "CPUs: %d\n", details.CPUs)
	fmt.Fprintf(out, "CPU Usage: %.2f%%\n", details.CPU*100)
	fmt.Fprintf(out, "Memory: %s / %s\n", formatBytes(details.Memory), formatBytes(details.MaxMemory))
	fmt.Fprintf(out, "Disk: %s / %s\n", formatBytes(details.Disk), formatBytes(details.MaxDisk))

	return nil
}

func formatBytes(bytes uint64) string {
	const gibibyte = 1024 * 1024 * 1024
	return fmt.Sprintf("%.2f GiB", float64(bytes)/gibibyte)
}

func formatUptime(uptime uint64) string {
	duration := time.Duration(uptime) * time.Second
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%d days, %d hours, %d minutes, %d seconds", days, hours, minutes, seconds)
}
