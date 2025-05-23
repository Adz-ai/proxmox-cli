package test

import (
	"os"
	"strings"

	"github.com/golang/mock/gomock"
	"proxmox-cli/cmd/utility"
	"proxmox-cli/internal/interfaces"
	"proxmox-cli/test/mocks"
)

// TestConfig holds test configuration
type TestConfig struct {
	UseMock     bool
	ProxmoxURL  string
	ProxmoxUser string
	ProxmoxPass string
	TestNode    string
	TestVMID    int
	TestLXCID   int
}

// GetTestConfig returns test configuration based on environment variables
func GetTestConfig() *TestConfig {
	config := &TestConfig{
		UseMock:   true,
		TestNode:  "pve",
		TestVMID:  999,
		TestLXCID: 998,
	}

	// Check if we should use real Proxmox server
	if os.Getenv("PROXMOX_TEST_MODE") == "real" {
		config.UseMock = false
		config.ProxmoxURL = os.Getenv("PROXMOX_TEST_URL")
		config.ProxmoxUser = os.Getenv("PROXMOX_TEST_USER")
		config.ProxmoxPass = os.Getenv("PROXMOX_TEST_PASS")

		if config.ProxmoxURL == "" || config.ProxmoxUser == "" || config.ProxmoxPass == "" {
			panic("Real mode requires PROXMOX_TEST_URL, PROXMOX_TEST_USER, and PROXMOX_TEST_PASS")
		}
	}

	if testNode := os.Getenv("PROXMOX_TEST_NODE"); testNode != "" {
		config.TestNode = testNode
	}

	return config
}

// ShouldSkipTest checks if a test should be skipped based on tags
func ShouldSkipTest(tags string) bool {
	if os.Getenv("PROXMOX_TEST_MODE") != "real" && strings.Contains(tags, "@real-only") {
		return true
	}
	if os.Getenv("PROXMOX_TEST_MODE") == "real" && strings.Contains(tags, "@mock-only") {
		return true
	}
	return false
}

// SetupMockClient configures the mock client for testing
func SetupMockClient(ctrl *gomock.Controller) *mocks.MockProxmoxClientInterface {
	mockClient := mocks.NewMockProxmoxClientInterface(ctrl)
	
	// Set up the dependency injection
	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface {
		return mockClient
	})
	
	return mockClient
}

// CleanupMockClient resets the client factory
func CleanupMockClient() {
	utility.ResetClientFactory()
}