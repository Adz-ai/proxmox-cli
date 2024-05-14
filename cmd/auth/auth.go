package auth

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "Commands related with Authorization",
	Long:  "Authorization in the Proxmox cluster",
}

func init() {
}
