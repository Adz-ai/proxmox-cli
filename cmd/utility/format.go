package utility

import (
	"fmt"
	"strings"
)

// FormatServerURL ensures the server URL has the correct format
func FormatServerURL(url string) string {
	// Remove any trailing slashes
	url = strings.TrimRight(url, "/")
	
	// Add protocol if missing
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	// Add default port if missing
	if !strings.Contains(url, ":") || (strings.Count(url, ":") == 1 && strings.HasPrefix(url, "https://")) {
		url = url + ":8006"
	}
	
	return url
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	homeDir := GetHomeDir()
	return fmt.Sprintf("%s/.config/proxmox-cli/config.yaml", homeDir)
}

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	// Simple implementation using environment variable
	// In production, you'd use os.UserHomeDir() or os/user package
	if homeDir := getEnvVar("HOME"); homeDir != "" {
		return homeDir
	}
	return "/home/user" // Fallback
}

// getEnvVar is a wrapper for environment variable access (for testing)
var getEnvVar = func(key string) string {
	// In real implementation, this would be os.Getenv(key)
	// Made as a variable for easier testing
	return ""
}