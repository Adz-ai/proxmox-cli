package utility

import (
	"context"
	"crypto/tls"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/Adz-ai/proxmox-cli/test/mocks"
)

func TestCheckIfAuthPresent(t *testing.T) {
	t.Cleanup(viper.Reset)
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
			errorContains: "not configured",
		},
		{
			name: "only server configured",
			setupConfig: func() {
				viper.Reset()
				viper.Set("server_url", "https://192.168.1.100:8006")
			},
			expectedError: true,
			errorContains: "not authenticated",
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
			errorContains: "not authenticated",
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
	t.Cleanup(viper.Reset)
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

			client, err := GetClient()
			if err != nil {
				t.Fatalf("GetClient() error = %v", err)
			}
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
	client, err := GetClient()
	if err != nil {
		t.Fatalf("GetClient() error = %v", err)
	}
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
	client, err := GetClient()
	if err != nil {
		t.Fatalf("GetClient() error = %v", err)
	}
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
	client, err := GetClient()
	if err != nil {
		t.Fatalf("GetClient() error = %v", err)
	}
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

func TestNormalizeServerURL(t *testing.T) {
	got, err := NormalizeServerURL("pve.example.com:8006/api2/json/")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://pve.example.com:8006" {
		t.Fatalf("NormalizeServerURL() = %q", got)
	}
	if _, err := NormalizeServerURL("ftp://pve.example.com"); err == nil {
		t.Fatal("expected unsupported scheme error")
	}
	if _, err := NormalizeServerURL("http://pve.example.com"); err == nil {
		t.Fatal("expected plaintext HTTP to be rejected")
	}
}

func TestWaitForCompletedTask(t *testing.T) {
	if err := WaitForTask(context.Background(), &proxmox.Task{IsSuccessful: true}, time.Second); err != nil {
		t.Fatal(err)
	}
	if err := WaitForTask(context.Background(), &proxmox.Task{IsFailed: true, ExitStatus: "failed"}, time.Second); err == nil {
		t.Fatal("expected failed task error")
	}
	if err := WaitForTask(context.Background(), nil, time.Second); err == nil {
		t.Fatal("expected nil task error")
	}
	// A task without a UPID comes from a synchronous API response; there is
	// nothing to poll and the operation already succeeded.
	if err := WaitForTask(context.Background(), &proxmox.Task{}, time.Second); err != nil {
		t.Fatalf("synchronous response should succeed, got %v", err)
	}
}

func TestSummarizeRRDSkipsNaNSamples(t *testing.T) {
	nan := math.NaN()
	summary := SummarizeRRD("hour", []*proxmox.RRDData{
		{Time: 1, CPU: 0.20, Mem: 2 << 30, MaxMem: 4 << 30, NetIn: 1024},
		{Time: 2, CPU: nan, Mem: nan},
		{Time: 3, CPU: 0.40, Mem: 3 << 30, MaxMem: 4 << 30, NetIn: 3072},
		nil,
	})
	if summary.Samples != 2 {
		t.Fatalf("samples = %d, want 2 (NaN and nil skipped)", summary.Samples)
	}
	if summary.PeakCPU != 40 || summary.AverageCPU != 30 {
		t.Fatalf("cpu peak/avg = %.1f/%.1f, want 40/30", summary.PeakCPU, summary.AverageCPU)
	}
	if summary.LatestMemory != 3<<30 || summary.MaxMemory != 4<<30 {
		t.Fatalf("memory latest/max = %d/%d", summary.LatestMemory, summary.MaxMemory)
	}
	if summary.AverageNetIn != 2048 {
		t.Fatalf("avg net in = %.0f, want 2048", summary.AverageNetIn)
	}
}

func TestParseTimeframe(t *testing.T) {
	if _, err := ParseTimeframe("day"); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseTimeframe("fortnight"); err == nil {
		t.Fatal("expected error for unsupported timeframe")
	}
}

func TestNewHTTPClientTLSDefaults(t *testing.T) {
	client, err := NewHTTPClient(false, "")
	if err != nil {
		t.Fatal(err)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.Transport)
	}
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("TLS verification must be enabled by default")
	}
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Fatal("minimum TLS version must be TLS 1.2")
	}
}

func TestWriteConfigPermissions(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "config", "config.json")
	t.Setenv("PROXMOX_CLI_CONFIG", path)
	viper.Set("server_url", "https://pve.example.com:8006")
	if err := WriteConfig(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions = %o, want 600", info.Mode().Perm())
	}
}

func TestWriteConfigCanonicalizesAuthTicket(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("PROXMOX_CLI_CONFIG", path)
	viper.Set("server_url", "https://pve.example.com:8006")
	viper.Set("auth_ticket.ticket", "ticket-value")
	viper.Set("auth_ticket.CSRFPreventionToken", "token-value")
	if err := WriteConfig(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"CSRFPreventionToken": "token-value"`) {
		t.Fatalf("CSRF token key not written in documented casing:\n%s", data)
	}
	if strings.Contains(string(data), "csrfpreventiontoken") {
		t.Fatalf("lowercased CSRF token key leaked into config file:\n%s", data)
	}
}

func TestLoadConfigSecuresLegacyFile(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("PROXMOX_CLI_CONFIG", path)
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("legacy config permissions = %o, want 600", info.Mode().Perm())
	}
}
