package nodes

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Manage nodes",
		Long:  "Manage nodes in the Proxmox cluster",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newDescribeCmd())
	// TODO: Add these commands when we understand the API better
	// cmd.AddCommand(newStorageCmd())
	// cmd.AddCommand(newTasksCmd())
	// cmd.AddCommand(newServicesCmd())

	return cmd
}
