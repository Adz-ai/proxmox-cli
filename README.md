# Proxmox CLI 🚀

A powerful command-line interface for managing Proxmox Virtual Environment (PVE) resources. Control your virtualization infrastructure from the terminal with ease!

## ✨ Features

- **🔐 Secure Authentication**: Login with username/password, stores session tokens
- **🖥️ Virtual Machine Management**: Create, delete, list, and describe VMs
- **📦 LXC Container Support**: Complete container lifecycle management
- **🌐 Node Operations**: Monitor and manage cluster nodes
- **🎯 User-Friendly**: Clear error messages, helpful prompts, and guided setup
- **⚡ Fast & Efficient**: Direct API communication with Proxmox VE
- **🧪 Fully Tested**: Comprehensive BDD test suite with mocking

## 📋 Requirements

- Go 1.23+ (for building from source)
- Proxmox VE 6.0+ server
- Network access to Proxmox API (usually port 8006)

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/Adz-ai/proxmox-cli.git
cd proxmox-cli

# Build the CLI
go build -o proxmox-cli main.go

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

## 📚 Commands Reference

### Authentication & Setup
```bash
proxmox-cli init                    # Configure server connection (interactive)
proxmox-cli init --force            # Reconfigure existing setup
proxmox-cli auth login -u <user>    # Authenticate with Proxmox
proxmox-cli auth logout             # Clear authentication tokens
proxmox-cli status                  # Check configuration and connection
proxmox-cli status --verbose        # Detailed status with server info
```

### Virtual Machine Management
```bash
# List and inspect VMs
proxmox-cli vm get                  # List all VMs across all nodes
proxmox-cli vm describe -n <node> -i <vmid>  # Show detailed VM information
proxmox-cli vm status -n <node> -i <vmid>    # Show VM status

# VM lifecycle
proxmox-cli vm create -n <node> -i <vmid> -s <spec.yaml>  # Create VM from YAML
proxmox-cli vm start -n <node> -i <vmid>     # Start a VM
proxmox-cli vm stop -n <node> -i <vmid>      # Stop a VM (graceful shutdown)
proxmox-cli vm stop -n <node> -i <vmid> -f   # Force stop (power off)
proxmox-cli vm restart -n <node> -i <vmid>   # Restart a VM (graceful reboot)
proxmox-cli vm restart -n <node> -i <vmid> -f  # Force restart (reset)
proxmox-cli vm delete -n <node> -i <vmid>    # Delete a VM
```

### LXC Container Management
```bash
# List and inspect containers
proxmox-cli lxc get                 # List all containers across all nodes
proxmox-cli lxc describe -n <node> -i <ctid>  # Show detailed container information
proxmox-cli lxc status -n <node> -i <ctid>    # Show container status

# Container lifecycle
proxmox-cli lxc create -n <node> -i <ctid> -s <spec.yaml>  # Create from YAML
proxmox-cli lxc start -n <node> -i <ctid>     # Start a container
proxmox-cli lxc stop -n <node> -i <ctid>      # Stop a container
proxmox-cli lxc restart -n <node> -i <ctid>   # Restart a container
proxmox-cli lxc delete -n <node> -i <ctid>    # Delete a container
proxmox-cli lxc delete -n <node> -i <ctid> -f # Force delete a container

# Advanced operations
proxmox-cli lxc clone -n <node> -s <source> -t <target> --name <name>  # Clone container
proxmox-cli lxc clone -n <node> -s <source> -t <target> --full        # Full clone
proxmox-cli lxc snapshot create -n <node> -i <ctid> --name <snapshot>  # Create snapshot
proxmox-cli lxc snapshot list -n <node> -i <ctid>                      # List snapshots
```

### Node Management
```bash
# Node information
proxmox-cli nodes get               # List all cluster nodes with status
proxmox-cli nodes describe -n <node>  # Show detailed node information

# Node operations
proxmox-cli nodes storage -n <node>     # List storage on node
proxmox-cli nodes tasks -n <node>       # List recent tasks on node
proxmox-cli nodes tasks -n <node> -l 50 # List last 50 tasks

# Future node operations (planned)
# proxmox-cli nodes services list -n <node>  # List services on node
```

## 🔧 Configuration

Configuration is stored in `~/.proxmox-cli/config.json`

### Sample Configuration
```json
{
  "server_url": "https://your-proxmox:8006",
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

## 📝 Creating Resources from YAML

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

## 🧪 Development & Testing

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
The project includes comprehensive BDD tests covering:
- ✅ Authentication workflows
- ✅ Node management operations  
- ✅ LXC container lifecycle
- ✅ VM management operations
- ✅ Error handling scenarios

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

## 🚦 Status & Roadmap

### ✅ Implemented Features
- Authentication and session management
- Node listing, detailed information, storage, and tasks
- VM operations (list, describe, create, delete, start, stop, restart, status)
- LXC operations (list, describe, create, start, stop, restart, delete, status)
- LXC advanced operations (clone with full/linked support, snapshots)
- Comprehensive error handling
- Full BDD test coverage

### 📋 Planned Features
- VM snapshots and cloning
- Template management
- Network configuration management
- Storage pool management
- Backup and restore operations
- Bulk operations
- Configuration profiles
- Services management on nodes

## 🤝 Contributing

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

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [go-proxmox](https://github.com/luthermonson/go-proxmox) for Proxmox API
- Testing with [Godog](https://github.com/cucumber/godog) for BDD
- Mocking with [GoMock](https://github.com/golang/mock) for reliable tests
- Inspired by the need for better Proxmox CLI tools

## 📞 Support

If you encounter issues or have questions:
1. Check the [Issues](https://github.com/Adz-ai/proxmox-cli/issues) page
2. Create a new issue with detailed information
3. Include logs and configuration (redact sensitive information)

---

**Happy Proxmox management!** 🎉