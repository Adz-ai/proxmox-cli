package vm

import (
	"github.com/spf13/cobra"
)

var VMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Commands related with Virtual Machines",
	Long:  "Manage virtual machines in the Proxmox cluster",
}

func init() {
	VMCmd.AddCommand(getCmd)
	VMCmd.AddCommand(createVMCmd)
	VMCmd.AddCommand(deleteCmd)
	VMCmd.AddCommand(describeCmd)
}
