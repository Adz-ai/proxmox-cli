# Proxmox CLI

A powerful command-line interface for managing Proxmox Virtual Environment (PVE) resources. Control your virtualization infrastructure from the terminal with ease!

## Features

- **Secure Authentication**: Verified TLS by default with private session storage
- **Virtual Machine Management**: Create, list, describe, start, stop, restart, and delete VMs
- **LXC Container Support**: Full container lifecycle plus snapshot create and list
- **Node Operations**: Monitor and manage cluster nodes
- **User-Friendly**: Clear error messages, helpful prompts, and guided setup
- **Fast & Efficient**: Direct API communication with Proxmox VE
- **Automated Testing**: Unit, mocked CLI, and BDD coverage in CI

## Requirements

- Go 1.26+ (for building from source)
- Proxmox VE 6.0+ server
- Network access to Proxmox API (usually port 8006)

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/Adz-ai/proxmox-cli.git
cd proxmox-cli

# Build the CLI
go build -o proxmox-cli main.go

# Or install directly
go install github.com/Adz-ai/proxmox-cli@latest

# Optional: Move to PATH
sudo mv proxmox-cli /usr/local/bin/
```

### First Time Setup

```bash
# 1. Run the CLI to see available commands
proxmox-cli

# 2. Configure server connection (interactive)
proxmox-cli init

# 3. Login with your credentials
proxmox-cli auth login -u root@pam

# 4. Check your connection
proxmox-cli status --verbose

# 5. Start using it!
proxmox-cli nodes get
proxmox-cli vm get
proxmox-cli lxc get
```

## Commands Reference

### Authentication & Setup
```bash
proxmox-cli init                    # Configure server connection (interactive)
proxmox-cli init --force            # Reconfigure existing setup
proxmox-cli auth login -u <user>    # Authenticate with Proxmox
proxmox-cli auth logout             # Clear authentication tokens
proxmox-cli status                  # Check configuration and connection
proxmox-cli status --verbose        # Detailed status with server info
proxmox-cli --version               # Show build version
```

### Virtual Machine Management
```bash
# List and inspect VMs
proxmox-cli vm get                  # List all VMs across all nodes
proxmox-cli vm describe -n <node> -i <vmid>  # Show detailed VM information

# VM lifecycle
proxmox-cli vm create -n <node> -i <vmid> -s <spec.yaml>  # Create VM from YAML
proxmox-cli vm start -n <node> -i <vmid>     # Start a VM
proxmox-cli vm stop -n <node> -i <vmid>      # Stop a VM
proxmox-cli vm restart -n <node> -i <vmid>   # Restart a VM
proxmox-cli vm delete -n <node> -i <vmid>    # Delete a VM
```

### LXC Container Management
```bash
# List and inspect containers
proxmox-cli lxc get                 # List all containers across all nodes
proxmox-cli lxc describe -n <node> -i <ctid>  # Show container details

# Container lifecycle
proxmox-cli lxc create -n <node> -i <ctid> -s <spec.yaml>  # Create from YAML
proxmox-cli lxc start -n <node> -i <ctid>     # Start a container
proxmox-cli lxc stop -n <node> -i <ctid>      # Stop a container
proxmox-cli lxc restart -n <node> -i <ctid>   # Restart a container
proxmox-cli lxc delete -n <node> -i <ctid>    # Delete a container
proxmox-cli lxc delete -n <node> -i <ctid> --force --purge # Force deletion, removing related configuration

# Snapshots
proxmox-cli lxc snapshot create -n <node> -i <ctid> --name <snapshot>
proxmox-cli lxc snapshot list -n <node> -i <ctid>

# Advanced operations (coming soon)
proxmox-cli lxc clone -n <node> -s <source> -t <target> --name <name>
```

### Node Management
```bash
proxmox-cli nodes get               # List all cluster nodes with status
proxmox-cli nodes describe -n <node>  # Show detailed node information

# Future node operations (planned)
# proxmox-cli nodes storage -n <node>     # List storage on node
# proxmox-cli nodes tasks -n <node>       # List tasks on node
# proxmox-cli nodes services list -n <node>  # List services on node
```

## Configuration

Configuration is stored in `~/.proxmox-cli/config.json`

TLS certificates are verified by default. For a private CA, configure its PEM file with `proxmox-cli init --force --ca-cert /path/to/ca.pem`. For isolated lab environments only, `--insecure` disables certificate verification.

### Sample Configuration
```json
{
  "server_url": "https://your-proxmox:8006",
  "insecure": false,
  "ca_cert": "",
  "auth_ticket": {
    "ticket": "PVE:user@realm:...",
    "CSRFPreventionToken": "..."
  }
}
```

### Manual Configuration
```bash
proxmox-cli init              # Interactive setup
proxmox-cli init --force      # Reconfigure existing
```

## Creating Resources from YAML

### VM Specification
```yaml
# vm-spec.yaml
name: "my-ubuntu-vm"
memory: 2048
cores: 2
sockets: 1
scsi0: "local-lvm:32"
ostype: "l26"
net0: "virtio,bridge=vmbr0"
ide2: "local:iso/ubuntu-22.04.iso,media=cdrom"
boot: "order=ide2;scsi0"
```

```bash
proxmox-cli vm create -n node1 -i 100 -s vm-spec.yaml
```

### LXC Container Specification
```yaml
# lxc-spec.yaml
ostemplate: "local:vztmpl/ubuntu-20.04-standard_20.04-1_amd64.tar.gz"
hostname: "my-container"
memory: 1024
swap: 512
cores: 2
rootfs: "local-lvm:8"
net0: "name=eth0,bridge=vmbr0,ip=dhcp"
password: "secure-password"
start: 1
unprivileged: 1
```

```bash
proxmox-cli lxc create -n node1 -i 200 -s lxc-spec.yaml
```

## Development & Testing

### Running Tests
```bash
# Run BDD tests (uses mocks)
go test ./test/bdd -v

# Run all tests
go test ./...

# Run specific test scenario
go test ./test/bdd -v -run "TestFeatures/List_all_nodes"

# Run with coverage
go test -cover ./...
```

### BDD Test Features
The active BDD suite covers configuration, authentication state, node inspection, and the implemented LXC lifecycle. Planned commands are tagged `@skip` until implemented.

Test features are located in `test/bdd/features/` and can be run directly from IDEs that support Cucumber/Gherkin.

### Building
```bash
# Standard build
go build -o proxmox-cli main.go

# Build with version info
go build -ldflags "-X main.version=v1.0.0" -o proxmox-cli main.go

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o proxmox-cli-linux main.go
GOOS=windows GOARCH=amd64 go build -o proxmox-cli.exe main.go
GOOS=darwin GOARCH=amd64 go build -o proxmox-cli-macos main.go
```

### Project Structure
```
proxmox-cli/
├── cmd/                    # CLI command implementations
│   ├── auth/              # Authentication commands
│   ├── lxc/               # LXC container commands
│   ├── nodes/             # Node management commands
│   ├── vm/                # Virtual machine commands
│   └── utility/           # Shared utilities
├── internal/              # Internal packages
│   └── interfaces/        # API interfaces for dependency injection
├── test/                  # Test suite
│   ├── bdd/               # BDD tests with Cucumber/Gherkin
│   │   └── features/      # Feature specifications
│   ├── integration/       # Integration tests
│   └── mocks/             # Generated mocks for testing
└── main.go               # CLI entry point
```

## Status & Roadmap

### Implemented Features
- Authentication and session management
- Node listing and detailed information
- VM operations (list, describe, create, start, stop, restart, delete)
- LXC operations (list, describe, create, start, stop, restart, delete)
- LXC snapshots (create, list)
- Nonzero exit statuses for operational failures
- TLS verification, custom CA support, and private config files

### In Development
- VM and LXC clone operations
- Node storage and task management
- VM snapshots

### Planned Features
- Template management
- Network configuration
- Storage management
- Backup operations
- Bulk operations
- Configuration profiles

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run the test suite (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines
- All new features should include BDD tests
- Follow existing code patterns and conventions
- Update documentation for new functionality
- Ensure all tests pass before submitting

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [go-proxmox](https://github.com/luthermonson/go-proxmox) for Proxmox API
- Testing with [Godog](https://github.com/cucumber/godog) for BDD
- Mocking with [GoMock](https://github.com/uber-go/mock) for reliable tests
- Inspired by the need for better Proxmox CLI tools

## Support

If you encounter issues or have questions:
1. Check the [Issues](https://github.com/Adz-ai/proxmox-cli/issues) page
2. Create a new issue with detailed information
3. Include logs and configuration (redact sensitive information)

---

**Happy Proxmox management!**
