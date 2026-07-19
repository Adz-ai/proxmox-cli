package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

			authTicket := viper.Sub("auth_ticket")
			wasAuthenticated := authTicket != nil && authTicket.GetString("ticket") != ""

			utility.ClearAuthTicket()
			if err := utility.WriteConfig(); err != nil {
				return fmt.Errorf("clear authentication: %w", err)
			}

			if !wasAuthenticated {
				fmt.Fprintln(out, "Not currently logged in")
				return nil
			}

			fmt.Fprintln(out, "Logged out successfully")
			fmt.Fprintln(out, "Your authentication has been cleared")
			return nil
		},
	}
}
