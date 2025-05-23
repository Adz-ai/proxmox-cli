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
		// Check if authenticated
		authTicket := viper.Sub("auth_ticket")
		if authTicket == nil || authTicket.GetString("ticket") == "" {
			fmt.Println("ℹ️  Not currently logged in")
			return
		}

		// Clear auth ticket
		viper.Set("auth_ticket", nil)
		err := viper.WriteConfig()
		if err != nil {
			fmt.Printf("❌ Failed to clear authentication: %s\n", err)
			return
		}

		fmt.Println("✅ Logged out successfully")
		fmt.Println("👋 Your authentication has been cleared")
	},
}

func init() {
	AuthCmd.AddCommand(logoutCmd)
}