package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out from Proxmox and clear authentication",
		Long:  `Clear the stored authentication ticket and CSRF token.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if !utility.HasSessionTicket() && !utility.HasAPIToken() {
				fmt.Fprintln(out, "Not currently logged in")
				return nil
			}

			utility.ClearAuthTicket()
			utility.ClearAPIToken()
			if err := utility.WriteConfig(); err != nil {
				return fmt.Errorf("clear authentication: %w", err)
			}

			fmt.Fprintln(out, "Logged out successfully")
			fmt.Fprintln(out, "Your authentication has been cleared")
			return nil
		},
	}
}
