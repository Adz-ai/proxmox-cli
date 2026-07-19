package vm

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec -n <node> -i <vmid> -- command [args...]",
		Short: "Run a command inside a virtual machine via the guest agent",
		Long: `Execute a command in the guest through the QEMU guest agent and print
its output, e.g.:

  proxmox-cli vm exec -n pve -i 100 -- uname -a

The guest agent must be installed and running inside the VM.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()
			ctx := cmd.Context()
			inputData, err := cmd.Flags().GetString("input-data")
			if err != nil {
				return fmt.Errorf("read input-data flag: %w", err)
			}

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			if err := vm.WaitForAgent(ctx, 15); err != nil {
				return fmt.Errorf("guest agent not responding in VM %d (is qemu-guest-agent installed and running?): %w", id, err)
			}

			pid, err := vm.AgentExec(ctx, args, inputData)
			if err != nil {
				return fmt.Errorf("execute command in VM %d: %w", id, err)
			}

			seconds := int(utility.TaskTimeout(cmd).Seconds())
			if seconds < 1 {
				seconds = 1
			}
			status, err := vm.WaitForAgentExecExit(ctx, pid, seconds)
			if err != nil {
				return fmt.Errorf("wait for command in VM %d: %w", id, err)
			}

			if status.OutData != "" {
				fmt.Fprint(out, status.OutData)
				if !strings.HasSuffix(status.OutData, "\n") {
					fmt.Fprintln(out)
				}
			}
			if status.ErrData != "" {
				fmt.Fprint(errOut, status.ErrData)
				if !strings.HasSuffix(status.ErrData, "\n") {
					fmt.Fprintln(errOut)
				}
			}
			if bool(status.OutTruncated) || status.ErrTruncated {
				fmt.Fprintln(errOut, "warning: output was truncated by the guest agent")
			}
			if status.ExitCode != 0 {
				return fmt.Errorf("command exited with code %d", status.ExitCode)
			}
			return nil
		},
	}

	addVMTargetFlags(cmd)
	cmd.Flags().String("input-data", "", "Data to pass to the command on stdin")
	return cmd
}

type vmInterfaceSummary struct {
	Name      string   `json:"name"`
	MAC       string   `json:"mac,omitempty"`
	Addresses []string `json:"addresses"`
}

func newIPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip",
		Short: "Show virtual machine IP addresses via the guest agent",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			interfaces, err := vm.AgentGetNetworkIFaces(ctx)
			if err != nil {
				return fmt.Errorf("get network interfaces of VM %d (is qemu-guest-agent running?): %w", id, err)
			}

			summaries := make([]vmInterfaceSummary, 0, len(interfaces))
			for _, iface := range interfaces {
				addresses := make([]string, 0, len(iface.IPAddresses))
				for _, address := range iface.IPAddresses {
					addresses = append(addresses, fmt.Sprintf("%s/%d", address.IPAddress, address.Prefix))
				}
				summaries = append(summaries, vmInterfaceSummary{
					Name:      iface.Name,
					MAC:       iface.HardwareAddress,
					Addresses: addresses,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintf(out, "Network interfaces of VM %d:\n", id)
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

	addVMTargetFlags(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}

func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show virtual machine resource usage",
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

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			samples, err := vm.RRDData(ctx, timeframe)
			if err != nil {
				return fmt.Errorf("get stats for VM %d: %w", id, err)
			}

			summary := utility.SummarizeRRD(timeframeValue, samples)
			if format == "json" {
				return utility.PrintJSON(out, summary)
			}
			utility.PrintRRDSummary(out, fmt.Sprintf("VM %d", id), summary)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	utility.AddTimeframeFlag(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}
