package vm

import (
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

	return cmd
}
