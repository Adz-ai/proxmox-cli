package vm

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "vm",
	Short: "Commands related with Virtual Machines",
	Long:  "Manage virtual machines in the Proxmox cluster",
}

func init() {
}
