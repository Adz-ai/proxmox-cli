package auth

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Commands related with Authorization",
		Long:  "Authorization in the Proxmox cluster",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newLoginCmd(), newLogoutCmd())
	return cmd
}
