/*
Copyright © 2024 Adarssh Athithan
*/
package nodes

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "nodes",
	Short: "Manage nodes",
	Long:  "Manage nodes in the Proxmox cluster",
}

func init() {}
