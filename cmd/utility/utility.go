package utility

import (
	"crypto/tls"
	"errors"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func CheckIfAuthPresent() error {
	// First check if server URL is configured
	serverURL := viper.GetString("server_url")
	if serverURL == "" {
		return errors.New("❌ Not configured. Please run 'proxmox-cli auth login -u <username>' to set up.")
	}

	// Check if the client is authenticated
	authTicket := viper.Sub("auth_ticket")
	if authTicket == nil || authTicket.GetString("ticket") == "" || authTicket.GetString("CSRFPreventionToken") == "" {
		return errors.New("❌ Not authenticated. Please run 'proxmox-cli auth login -u <username>' to log in.")
	}
	return nil
}

func GetClient() *proxmox.Client {
	endpoint := viper.GetString("server_url")
	if endpoint == "" {
		log.Fatal("❌ Proxmox server URL not configured. Please run 'proxmox-cli auth login -u <username>'")
	}

	// Create HTTP client with TLS config
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Create Proxmox client with session
	authTicket := viper.Sub("auth_ticket")
	if authTicket != nil {
		ticket := authTicket.GetString("ticket")
		csrfToken := authTicket.GetString("CSRFPreventionToken")
		if ticket != "" && csrfToken != "" {
			// Use WithSession option to set auth
			client := proxmox.NewClient(endpoint+"/api2/json", 
				proxmox.WithHTTPClient(httpClient),
				proxmox.WithSession(ticket, csrfToken))
			return client
		}
	}

	// Return client without auth if no session found
	return proxmox.NewClient(endpoint+"/api2/json", proxmox.WithHTTPClient(httpClient))
}
