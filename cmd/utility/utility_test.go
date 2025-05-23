package utility

import (
	"testing"

	"github.com/spf13/viper"
)

func TestCheckIfAuthPresent(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func()
		expectedError  bool
		errorContains  string
	}{
		{
			name: "no configuration",
			setupConfig: func() {
				viper.Reset()
				viper.Set("server_url", "")
			},
			expectedError: true,
			errorContains: "Not configured",
		},
		{
			name: "only server configured",
			setupConfig: func() {
				viper.Reset()
				viper.Set("server_url", "https://192.168.1.100:8006")
			},
			expectedError: true,
			errorContains: "Not authenticated",
		},
		{
			name: "server configured with empty auth ticket",
			setupConfig: func() {
				viper.Reset()
				viper.Set("server_url", "https://192.168.1.100:8006")
				viper.Set("auth_ticket.ticket", "")
				viper.Set("auth_ticket.CSRFPreventionToken", "")
			},
			expectedError: true,
			errorContains: "Not authenticated",
		},
		{
			name: "fully configured",
			setupConfig: func() {
				viper.Reset()
				viper.Set("server_url", "https://192.168.1.100:8006")
				viper.Set("auth_ticket.ticket", "test-ticket")
				viper.Set("auth_ticket.CSRFPreventionToken", "test-csrf")
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupConfig()

			err := CheckIfAuthPresent()

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGetClient(t *testing.T) {
	// Only test cases that don't call log.Fatal to avoid process exit
	tests := []struct {
		name        string
		setupConfig func()
	}{
		{
			name: "server URL configured without auth",
			setupConfig: func() {
				viper.Reset()
				viper.Set("server_url", "https://192.168.1.100:8006")
			},
		},
		{
			name: "server URL and auth configured",
			setupConfig: func() {
				viper.Reset()
				viper.Set("server_url", "https://192.168.1.100:8006")
				viper.Set("auth_ticket.ticket", "test-ticket")
				viper.Set("auth_ticket.CSRFPreventionToken", "test-csrf")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupConfig()

			client := GetClient()
			if client == nil {
				t.Errorf("Expected client but got nil")
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		indexSubstring(s, substr) >= 0)))
}

func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}