package integration_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"

	"proxmox-cli/cmd"
	"proxmox-cli/cmd/utility"
	"proxmox-cli/internal/interfaces"
	"proxmox-cli/test/mocks"
)

// TestLXCGetCommand tests the lxc get command with mocked API responses
func TestLXCGetCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockClient := mocks.NewMockProxmoxClientInterface(ctrl)
	mockNode := mocks.NewMockNodeInterface(ctrl)

	// Set up test config
	setupTestConfig(t)
	defer cleanupTestConfig()

	// Set up expectations
	ctx := gomock.Any()
	nodes := proxmox.NodeStatuses{
		&proxmox.NodeStatus{Node: "pve1", Status: "online"},
		&proxmox.NodeStatus{Node: "pve2", Status: "online"},
	}
	mockClient.EXPECT().Nodes(ctx).Return(nodes, nil)

	// For each node, expect container queries
	containers1 := proxmox.Containers{
		&proxmox.Container{VMID: 100, Name: "web-server", Status: "running", Uptime: 86400},
		&proxmox.Container{VMID: 101, Name: "database", Status: "stopped", Uptime: 0},
	}
	containers2 := proxmox.Containers{
		&proxmox.Container{VMID: 200, Name: "app-server", Status: "running", Uptime: 172800},
	}

	mockClient.EXPECT().Node(ctx, "pve1").Return(mockNode, nil)
	mockNode.EXPECT().Containers(ctx).Return(containers1, nil)
	
	mockClient.EXPECT().Node(ctx, "pve2").Return(mockNode, nil)
	mockNode.EXPECT().Containers(ctx).Return(containers2, nil)

	// Set up factory
	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface {
		return mockClient
	})
	defer utility.ResetClientFactory()

	// Execute command
	output := executeCommand(t, []string{"lxc", "get"})

	// Verify output
	expectedStrings := []string{
		"LXC Containers:",
		"Node: pve1",
		"100",
		"web-server",
		"running",
		"101",
		"database",
		"stopped",
		"Node: pve2",
		"200",
		"app-server",
		"running",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains(output, []byte(expected)) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", expected, output)
		}
	}
}

// TestLXCStartCommand tests the lxc start command
func TestLXCStartCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockClient := mocks.NewMockProxmoxClientInterface(ctrl)
	mockNode := mocks.NewMockNodeInterface(ctrl)
	mockContainer := mocks.NewMockContainerInterface(ctrl)

	// Set up test config
	setupTestConfig(t)
	defer cleanupTestConfig()

	// Set up expectations
	ctx := gomock.Any()
	mockClient.EXPECT().Node(ctx, "pve1").Return(mockNode, nil)
	mockNode.EXPECT().Container(ctx, 100).Return(mockContainer, nil)
	
	task := &proxmox.Task{
		UPID: proxmox.UPID("UPID:pve1:00001234:00112233:65432100:start"),
	}
	mockContainer.EXPECT().Start(ctx).Return(task, nil)

	// Set up factory
	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface {
		return mockClient
	})
	defer utility.ResetClientFactory()

	// Execute command
	output := executeCommand(t, []string{"lxc", "start", "-n", "pve1", "-i", "100"})

	// Verify output
	if !bytes.Contains(output, []byte("Container 100 started successfully")) {
		t.Errorf("Expected success message in output, got:\n%s", output)
	}
	if !bytes.Contains(output, []byte("UPID:pve1:00001234:00112233:65432100:start")) {
		t.Errorf("Expected UPID in output, got:\n%s", output)
	}
}

// TestNodeGetCommand tests the nodes get command
func TestNodeGetCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockClient := mocks.NewMockProxmoxClientInterface(ctrl)

	// Set up test config
	setupTestConfig(t)
	defer cleanupTestConfig()

	// Set up expectations
	ctx := gomock.Any()
	nodes := proxmox.NodeStatuses{
		&proxmox.NodeStatus{
			Node:   "pve1",
			Status: "online",
			Type:   "node",
			Uptime: 432000, // 5 days
			CPU:    0.15,
			MaxCPU: 8,
			Mem:    4294967296,    // 4GB
			MaxMem: 17179869184,   // 16GB
		},
		&proxmox.NodeStatus{
			Node:   "pve2",
			Status: "offline",
			Type:   "node",
			Uptime: 0,
		},
	}
	mockClient.EXPECT().Nodes(ctx).Return(nodes, nil)

	// Set up factory
	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface {
		return mockClient
	})
	defer utility.ResetClientFactory()

	// Execute command
	output := executeCommand(t, []string{"nodes", "get"})

	// Verify output
	expectedStrings := []string{
		"Nodes in cluster:",
		"pve1",
		"online",
		"5d 0h",
		"pve2",
		"offline",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains(output, []byte(expected)) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", expected, output)
		}
	}
}

// TestAuthenticationCheckFailure tests command behavior when not authenticated
func TestAuthenticationCheckFailure(t *testing.T) {
	// Clear any config
	viper.Reset()

	// Execute a protected command
	output := executeCommand(t, []string{"lxc", "get"})

	// Should see auth error
	if !bytes.Contains(output, []byte("Not configured")) || !bytes.Contains(output, []byte("Please run 'proxmox-cli auth login")) {
		t.Errorf("Expected authentication error message, got:\n%s", output)
	}
}

// Helper functions

func setupTestConfig(t *testing.T) {
	t.Helper()
	
	// Create temp config directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".proxmox-cli")
	os.MkdirAll(configDir, 0755)
	
	// Set up viper
	viper.Reset()
	viper.SetConfigFile(filepath.Join(configDir, "config.json"))
	viper.Set("server_url", "https://192.168.1.100:8006")
	viper.Set("auth_ticket.ticket", "PVE:root@pam:1234567890::abcdef")
	viper.Set("auth_ticket.CSRFPreventionToken", "1234567890:abcdef")
	viper.WriteConfig()
}

func cleanupTestConfig() {
	viper.Reset()
}

func executeCommand(t *testing.T, args []string) []byte {
	t.Helper()
	
	// Create command
	rootCmd := cmd.NewRootCmd()
	
	// Capture output
	var output bytes.Buffer
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&output)
	rootCmd.SetArgs(args)
	
	// Execute
	err := rootCmd.Execute()
	if err != nil {
		// Some commands may return errors (like auth failures), which is expected
		// The error is already written to output
	}
	
	return output.Bytes()
}