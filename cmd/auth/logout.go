package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from Proxmox and clear authentication",
	Long:  `Clear the stored authentication ticket and CSRF token.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		
		// Check if authenticated
		authTicket := viper.Sub("auth_ticket")
		if authTicket == nil || authTicket.GetString("ticket") == "" {
			fmt.Fprintln(out, "ℹ️  Not currently logged in")
			return
		}

		// Clear auth ticket
		viper.Set("auth_ticket", nil)
		err := viper.WriteConfig()
		if err != nil {
			fmt.Fprintf(out, "❌ Failed to clear authentication: %s\n", err)
			return
		}

		fmt.Fprintln(out, "✅ Logged out successfully")
		fmt.Fprintln(out, "👋 Your authentication has been cleared")
	},
}

func init() {
	Cmd.AddCommand(logoutCmd)
}