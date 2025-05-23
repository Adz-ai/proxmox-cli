# Testing Guide for Proxmox CLI

## Overview

The Proxmox CLI uses BDD (Behavior Driven Development) testing with Gherkin feature files. Tests can run in two modes:

1. **Mock Mode** (default) - Uses mock data without requiring a real Proxmox server
2. **Real Mode** - Tests against an actual Proxmox server

## Running Tests

### Mock Mode (Default)

```bash
# Run all tests with mock data
go test ./...

# Run specific test file
go test -run TestFeaturesEnhanced
```

### Real Mode

To test against a real Proxmox server:

```bash
# Set environment variables
export PROXMOX_TEST_MODE=real
export PROXMOX_TEST_URL=https://your-proxmox-server:8006
export PROXMOX_TEST_USER=root@pam
export PROXMOX_TEST_PASS=your-password
export PROXMOX_TEST_NODE=pve  # Optional, defaults to 'pve'

# Run tests
go test ./...
```

## Test Structure

### Feature Files

Located in `/features/`:
- `vm.feature` - Virtual machine management tests
- `lxc.feature` - LXC container management tests  
- `nodes.feature` - Node management tests

### Test Implementation

- `cmd_test.go` - Original VM-focused tests (kept for compatibility)
- `cmd_test_enhanced.go` - Enhanced tests supporting both VMs and LXCs with mock/real modes
- `test/config.go` - Test configuration management
- `test/mocks/` - Mock implementations for testing

## Writing New Tests

1. Add scenarios to appropriate `.feature` file
2. Implement step definitions in `cmd_test_enhanced.go`
3. For mock-only tests, data setup goes in `setupMockData()`
4. For real-mode tests, ensure proper cleanup in step implementations

## Best Practices

1. **Resource Cleanup**: All created resources (VMs, LXCs, snapshots) are tracked and cleaned up automatically
2. **Idempotency**: Tests should be runnable multiple times without side effects
3. **Isolation**: Each test scenario should be independent
4. **Mock First**: Write tests to work in mock mode first, then adapt for real mode

## Safety Considerations

When running tests against a real Proxmox server:

1. Use a dedicated test node or cluster
2. Use high VM/LXC IDs (e.g., 999, 998) to avoid conflicts
3. Always verify cleanup completed successfully
4. Don't run tests on production systems

## Continuous Integration

For CI/CD pipelines, use mock mode by default:

```yaml
# Example GitHub Actions
- name: Run Tests
  run: go test -v ./...
```

Only use real mode in controlled environments with proper credentials management.