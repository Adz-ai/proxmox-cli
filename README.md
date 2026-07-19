# Proxmox CLI

A powerful command-line interface for managing Proxmox Virtual Environment (PVE) resources. Control your virtualization infrastructure from the terminal with ease!

## Features

- **Secure Authentication**: Password or API token auth with verified TLS by default
- **Scriptable**: JSON output (`-o json`) on all read commands
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
proxmox-cli auth login -u <user>    # Authenticate with username and password
proxmox-cli auth token -t 'user@realm!tokenname'  # Authenticate with an API token
proxmox-cli auth logout             # Clear stored credentials
proxmox-cli status                  # Check configuration and connection
proxmox-cli status --verbose        # Detailed status with server info
proxmox-cli --version               # Show build version
```

Password logins use Proxmox session tickets, which expire after about two
hours. For scripts and long-lived setups, use an API token instead: create one
in the Proxmox web interface under Datacenter > Permissions > API Tokens, then
run `proxmox-cli auth token -t 'user@realm!tokenname'` and paste the secret
when prompted. Tokens do not expire and take precedence over a stored ticket.

### Global Flags
```bash
-o, --output table|json   # Structured output on get/describe/list commands
    --timeout <duration>  # Maximum time to wait for Proxmox tasks (default 10m)
```

### Cluster Overview
```bash
proxmox-cli get                     # Every VM, container, and storage in one view
proxmox-cli get --type vm           # Only VMs (also: lxc, storage)
proxmox-cli get -n <node>           # Only resources on one node
proxmox-cli get --status running    # Only running guests
```

### Virtual Machine Management
```bash
# List and inspect VMs
proxmox-cli vm get                  # List all VMs across all nodes
proxmox-cli vm get -n <node> --status running  # Filter by node and status
proxmox-cli vm describe -n <node> -i <vmid>  # Show detailed VM information

# VM lifecycle (omit -i on create to auto-assign the next free ID)
proxmox-cli vm create -n <node> -s <spec.yaml>  # Create VM from YAML
proxmox-cli vm start -n <node> -i <vmid>     # Start a VM
proxmox-cli vm shutdown -n <node> -i <vmid>  # Clean, guest-initiated shutdown
proxmox-cli vm stop -n <node> -i <vmid>      # Hard-stop a VM
proxmox-cli vm restart -n <node> -i <vmid>   # Restart a VM
proxmox-cli vm suspend -n <node> -i <vmid>   # Pause, keeping state in memory
proxmox-cli vm resume -n <node> -i <vmid>    # Resume a suspended VM
proxmox-cli vm delete -n <node> -i <vmid>    # Delete a VM

# Snapshots
proxmox-cli vm snapshot create -n <node> -i <vmid> --name <snapshot>
proxmox-cli vm snapshot list -n <node> -i <vmid>
proxmox-cli vm snapshot rollback -n <node> -i <vmid> --name <snapshot>
proxmox-cli vm snapshot delete -n <node> -i <vmid> --name <snapshot>
```

### LXC Container Management
```bash
# List and inspect containers
proxmox-cli lxc get                 # List all containers across all nodes
proxmox-cli lxc describe -n <node> -i <ctid>  # Show container details

# Container lifecycle (omit -i on create and -t on clone to auto-assign IDs)
proxmox-cli lxc create -n <node> -s <spec.yaml>  # Create from YAML
proxmox-cli lxc start -n <node> -i <ctid>     # Start a container
proxmox-cli lxc shutdown -n <node> -i <ctid>  # Clean shutdown (--force, --grace-seconds)
proxmox-cli lxc stop -n <node> -i <ctid>      # Hard-stop a container
proxmox-cli lxc restart -n <node> -i <ctid>   # Restart a container
proxmox-cli lxc suspend -n <node> -i <ctid>   # Suspend a container
proxmox-cli lxc resume -n <node> -i <ctid>    # Resume a suspended container
proxmox-cli lxc delete -n <node> -i <ctid>    # Delete a container
proxmox-cli lxc delete -n <node> -i <ctid> --force --purge # Force deletion, removing related configuration

# Snapshots and cloning
proxmox-cli lxc snapshot create -n <node> -i <ctid> --name <snapshot>
proxmox-cli lxc snapshot list -n <node> -i <ctid>
proxmox-cli lxc snapshot rollback -n <node> -i <ctid> --name <snapshot> [--start]
proxmox-cli lxc snapshot delete -n <node> -i <ctid> --name <snapshot>
proxmox-cli lxc clone -n <node> -s <source> --name <name>
```

### Node Management
```bash
proxmox-cli nodes get               # List all cluster nodes with status
proxmox-cli nodes describe -n <node>  # Show detailed node information
proxmox-cli nodes storage -n <node>   # List storage with type and usage
proxmox-cli nodes tasks -n <node>     # List recent tasks (-r for running only)
```

### Shell Completion
```bash
proxmox-cli completion bash > /etc/bash_completion.d/proxmox-cli
proxmox-cli completion zsh > "${fpath[1]}/_proxmox-cli"
proxmox-cli completion fish > ~/.config/fish/completions/proxmox-cli.fish
```

Node-name flags (`-n`) tab-complete live from the cluster when you are
authenticated.

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
  },
  "api_token": {
    "token_id": "user@realm!tokenname",
    "secret": "..."
  }
}
```

Only one of `auth_ticket` or `api_token` is normally present; `api_token`
wins when both are stored.

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
- Session and API token authentication
- Cluster-wide resource overview with type, node, and status filters
- Node listing, details, storage, and task history
- VM operations (list, describe, create, start, shutdown, stop, restart, suspend, resume, delete)
- LXC operations (all of the above plus clone)
- VM and LXC snapshots (create, list, rollback, delete)
- Auto-assigned guest IDs on create and clone
- Shell completion with live node-name lookup
- JSON output for read commands and configurable task timeouts
- Nonzero exit statuses for operational failures
- TLS verification, custom CA support, and private config files

### In Development
- VM clone operations
- Migration between nodes

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
