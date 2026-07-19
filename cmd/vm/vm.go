package vm

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Commands related with Virtual Machines",
		Long:  "Manage virtual machines in the Proxmox cluster",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newCreateVMCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newDescribeCmd())
	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newRestartCmd())
	cmd.AddCommand(newSnapshotCmd())

	return cmd
}

func addVMTargetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("node", "n", "", "Node name")
	cmd.Flags().IntP("id", "i", 0, "VM ID")
	for _, flag := range []string{"node", "id"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
}

func vmTargetFromFlags(cmd *cobra.Command) (string, int, error) {
	node, err := cmd.Flags().GetString("node")
	if err != nil {
		return "", 0, fmt.Errorf("get node flag: %w", err)
	}
	id, err := cmd.Flags().GetInt("id")
	if err != nil {
		return "", 0, fmt.Errorf("get id flag: %w", err)
	}
	node = strings.TrimSpace(node)
	if node == "" {
		return "", 0, fmt.Errorf("validate node: node cannot be empty")
	}
	if id <= 0 {
		return "", 0, fmt.Errorf("validate id: id must be positive")
	}
	return node, id, nil
}
