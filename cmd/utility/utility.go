package utility

import (
	"github.com/spf13/viper"
	"log"
)

func CheckIfAuthPresent() {
	// Check if the client is authenticated
	if viper.Sub("auth_ticket").GetString("ticket") == "" || viper.Sub("auth_ticket").GetString("CSRFPreventionToken") == "" {
		log.Fatalf("No authentication token found. Please run 'proxmox-cli auth' first.")
	}
}
