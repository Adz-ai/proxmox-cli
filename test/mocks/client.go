package mocks

// MockClient implements a mock Proxmox client for testing
// Currently simplified until we understand the exact struct fields
type MockClient struct {
	// Will be implemented when we have proper struct definitions
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

// AddMockNode adds a mock node with sample data
func (m *MockClient) AddMockNode(name string) {
	// To be implemented
}

// AddMockVM adds a mock VM to a specific node
func (m *MockClient) AddMockVM(nodeName string, vmid int, name string, status string) {
	// To be implemented
}

// AddMockContainer adds a mock container to a specific node
func (m *MockClient) AddMockContainer(nodeName string, vmid int, name string, status string) {
	// To be implemented
}
