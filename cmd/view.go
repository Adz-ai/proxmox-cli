package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// viewCmd represents the view subcommand
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the current authentication ticket",
	Long: `View details of the current authentication ticket stored in the configuration.
This command retrieves and displays the ticket and related information if available.`,
	Run: func(cmd *cobra.Command, args []string) {
		viewTicketDetails()
	},
}

func init() {
	authCmd.AddCommand(viewCmd)
}

func viewTicketDetails() {
	username := viper.Sub("auth_ticket").GetString("username")
	ticket := viper.Sub("auth_ticket").GetString("ticket")

	if ticket == "" {
		fmt.Println("No authentication ticket found. Please authenticate first.")
		return
	}

	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Authentication Ticket: %s\n", ticket)
}
