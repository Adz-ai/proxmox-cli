package utility

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	
	"proxmox-cli/internal/interfaces"
	"proxmox-cli/test/mocks"
)

func TestCheckIfAuthPresent(t *testing.T) {
	tests := []struct {
		name          string
		setupConfig   func()
		expectedError bool
		errorContains string
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

func TestGetClient_WithMockFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock client
	mockClient := mocks.NewMockProxmoxClientInterface(ctrl)
	
	// Set up factory to return our mock
	SetClientFactory(func() interfaces.ProxmoxClientInterface {
		return mockClient
	})
	defer ResetClientFactory()

	// Get client should return our mock
	client := GetClient()
	if client != mockClient {
		t.Error("GetClient should return the mock client")
	}
}

func TestNodesCommand_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock client
	mockClient := mocks.NewMockProxmoxClientInterface(ctrl)
	
	// Set up expectations
	expectedNodes := proxmox.NodeStatuses{
		&proxmox.NodeStatus{
			Node:   "pve1",
			Status: "online",
			Type:   "node",
			Uptime: 86400,
		},
		&proxmox.NodeStatus{
			Node:   "pve2", 
			Status: "online",
			Type:   "node",
			Uptime: 172800,
		},
	}
	
	ctx := context.Background()
	mockClient.EXPECT().Nodes(ctx).Return(expectedNodes, nil)

	// Set up factory
	SetClientFactory(func() interfaces.ProxmoxClientInterface {
		return mockClient
	})
	defer ResetClientFactory()

	// Get client and call Nodes
	client := GetClient()
	nodes, err := client.Nodes(ctx)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}
	
	if nodes[0].Node != "pve1" {
		t.Errorf("Expected first node to be pve1, got %s", nodes[0].Node)
	}
}

func TestContainerOperations_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockClient := mocks.NewMockProxmoxClientInterface(ctrl)
	mockNode := mocks.NewMockNodeInterface(ctrl)
	mockContainer := mocks.NewMockContainerInterface(ctrl)
	
	ctx := context.Background()
	
	// Set up expectations
	mockClient.EXPECT().Node(ctx, "pve").Return(mockNode, nil)
	mockNode.EXPECT().Container(ctx, 100).Return(mockContainer, nil)
	
	expectedTask := &proxmox.Task{
		UPID: proxmox.UPID("UPID:pve:00001234:00112233:65432100:start"),
	}
	mockContainer.EXPECT().Start(ctx).Return(expectedTask, nil)

	// Set up factory
	SetClientFactory(func() interfaces.ProxmoxClientInterface {
		return mockClient
	})
	defer ResetClientFactory()

	// Test the flow
	client := GetClient()
	node, err := client.Node(ctx, "pve")
	if err != nil {
		t.Fatalf("Unexpected error getting node: %v", err)
	}
	
	container, err := node.Container(ctx, 100)
	if err != nil {
		t.Fatalf("Unexpected error getting container: %v", err)
	}
	
	task, err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Unexpected error starting container: %v", err)
	}
	
	if task.UPID != expectedTask.UPID {
		t.Errorf("Expected UPID %s, got %s", expectedTask.UPID, task.UPID)
	}
}
