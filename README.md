# Proxmox CLI

[![CI](https://github.com/Adz-ai/proxmox-cli/actions/workflows/go.yml/badge.svg)](https://github.com/Adz-ai/proxmox-cli/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/v/release/Adz-ai/proxmox-cli)](https://github.com/Adz-ai/proxmox-cli/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/Adz-ai/proxmox-cli)](https://goreportcard.com/report/github.com/Adz-ai/proxmox-cli)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A command-line interface and k9s-style terminal UI for managing Proxmox Virtual Environment (PVE). Manage VMs, containers, snapshots, backups, and consoles from the terminal, across multiple clusters.

![proxmox-cli TUI](docs/tui.svg)

## Features

- **Secure Authentication**: Password or API token auth with verified TLS by default
- **Interactive TUI**: A k9s-style terminal UI (`proxmox-cli tui`) with live views and guest actions
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

Homebrew (macOS and Linux):

```bash
brew install adz-ai/tap/proxmox-cli
```

Released binaries for Linux, macOS, and Windows (plus `.deb`, `.rpm`, and
`.apk` packages) are on the [releases page](https://github.com/Adz-ai/proxmox-cli/releases/latest):

```bash
# Example: Linux amd64
curl -LO https://github.com/Adz-ai/proxmox-cli/releases/latest/download/checksums.txt
curl -LO "https://github.com/Adz-ai/proxmox-cli/releases/latest/download/proxmox-cli_<version>_linux_amd64.tar.gz"
tar -xzf proxmox-cli_*_linux_amd64.tar.gz proxmox-cli
sudo install proxmox-cli /usr/local/bin/
```

With Go:

```bash
go install github.com/Adz-ai/proxmox-cli@latest
```

Or build from source:

```bash
git clone https://github.com/Adz-ai/proxmox-cli.git
cd proxmox-cli
make build   # builds to build/proxmox-cli
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
    --context <name>      # Target a specific cluster context for this command
-y, --yes                 # Skip confirmation prompts on destructive commands
```

Destructive commands (`vm delete`, `lxc delete`, `snapshot rollback`,
`snapshot delete`, `backup restore --force`, `context delete`) ask for
confirmation before acting; pass `--yes` in scripts. Long-running
operations (backups, migrations, restores) stream the Proxmox task log
while they wait, so you can watch progress instead of a silent cursor.

### Multiple Clusters (Contexts)
```bash
proxmox-cli --context work init                   # Configure a new context
proxmox-cli --context work auth login -u root@pam # Authenticate it
proxmox-cli context list                          # List contexts (* marks current)
proxmox-cli context use work                      # Switch the current context
proxmox-cli context delete old-lab                # Remove a context and its credentials
proxmox-cli --context homelab vm get              # One-off command against another cluster
```

Each context holds its own server URL, TLS settings, and credentials.
Existing single-cluster configurations are migrated into a context named
`default` automatically the next time the CLI writes its config.

### Cluster Overview
```bash
proxmox-cli get                     # Every VM, container, and storage in one view
proxmox-cli get --type vm           # Only VMs (also: lxc, storage)
proxmox-cli get -n <node>           # Only resources on one node
proxmox-cli get --status running    # Only running guests
```

### Interactive TUI
```bash
proxmox-cli tui                     # k9s-style terminal UI (or: proxmox-cli --tui)
proxmox-cli tui --refresh 10s       # Slower auto-refresh (default 5s)
```

The TUI is styled after k9s: a header with cluster info (context, server,
user, PVE version, aggregate CPU/MEM), keyboard hints, and live-refreshing
views of guests, nodes, storage, and tasks in a bordered table with
breadcrumbs.

| Key | Action |
|-----|--------|
| `1` / `2` / `3` / `4`, `tab` | Switch between guests, nodes, storage, and tasks views |
| `:` | Command mode (`guests`, `vm`, `lxc`, `nodes`, `storage`, `tasks`, `help`, `quit`) |
| `j`/`k`, arrows, `g`/`G`, page keys | Move the selection |
| `/` | Filter rows (`esc` clears) |
| shift-key | Sort by column (`I` id, `N` name, `O` node, `S` status, `C` cpu, `M` mem, `A` age, `U` used, `T` total); same key inverts |
| `enter` | Describe the selected row |
| `s` / `d` / `x` / `r` | Start / shutdown / stop / reboot the selected guest |
| `c` | Interactive console on the selected guest (`Ctrl+]` to exit; requires session login) |
| `t` | Browse the selected guest's snapshots |
| `ctrl+d` | Delete the selected guest (stopped guests only, confirms) |
| `R` | Refresh immediately |
| `?` | Keyboard reference |
| `q` | Quit |

Destructive actions (shutdown, stop, reboot) ask for confirmation inside the
UI before any API call is made. Templates and non-guest rows are protected
from lifecycle actions.

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

# Cloning and migration
proxmox-cli vm clone -n <node> -s <source> --name <name> [--full] [--storage <storage>]
proxmox-cli vm migrate -n <node> -i <vmid> --target <node> [--online] [--with-local-disks]

# Configuration
proxmox-cli vm config set -n <node> -i <vmid> memory=4096 cores=4
proxmox-cli vm resize -n <node> -i <vmid> --disk scsi0 --size +10G
proxmox-cli vm tags -n <node> -i <vmid> --add web --remove old

# Guest agent, stats, and console
proxmox-cli vm exec -n <node> -i <vmid> -- uname -a   # Run a command in the guest
proxmox-cli vm ip -n <node> -i <vmid>                 # Show guest IP addresses
proxmox-cli vm stats -n <node> -i <vmid> [--timeframe hour|day|week|month|year]
proxmox-cli vm console -n <node> -i <vmid>            # Interactive console (Ctrl+] to exit)
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

# Migration and configuration
proxmox-cli lxc migrate -n <node> -i <ctid> --target <node> [--restart]
proxmox-cli lxc config set -n <node> -i <ctid> memory=2048 swap=512
proxmox-cli lxc resize -n <node> -i <ctid> --disk rootfs --size +2G
proxmox-cli lxc tags -n <node> -i <ctid> --add web --remove old

# Networking, stats, and console
proxmox-cli lxc ip -n <node> -i <ctid>                # Show container IP addresses
proxmox-cli lxc stats -n <node> -i <ctid> [--timeframe hour|day|week|month|year]
proxmox-cli lxc console -n <node> -i <ctid>           # Interactive console (Ctrl+] to exit)
```

### Templates & ISO Images
```bash
proxmox-cli template available -n <node>              # Templates downloadable from the appliance index
proxmox-cli template list -n <node> --storage <s>     # Downloaded templates
proxmox-cli template download -n <node> --storage <s> --template <name>
proxmox-cli iso list -n <node> --storage <s>
proxmox-cli iso download -n <node> --storage <s> --url <url> --filename <name>
```

Console access requires a password login (`auth login`); Proxmox does not
allow API tokens to open console websockets. `vm exec` and `vm ip` need the
QEMU guest agent installed and running inside the VM.

### Backup & Restore
```bash
proxmox-cli backup create -n <node> -i <vmid> --storage <storage> [--mode snapshot|suspend|stop]
proxmox-cli backup list -n <node> --storage <storage> [-i <vmid>]
proxmox-cli backup restore -n <node> -i <vmid> --archive <volid> [--storage <storage>] [--force]
```

The restore command detects whether the archive is a VM or container backup
from its name. `vm migrate` checks migration preconditions first, so blocking
problems (local resources, running without --online, local disks) are
reported before anything moves.

### Node Management
```bash
proxmox-cli nodes get               # List all cluster nodes with status
proxmox-cli nodes describe -n <node>  # Show detailed node information
proxmox-cli nodes storage -n <node>   # List storage with type and usage
proxmox-cli nodes tasks -n <node>     # List recent tasks (-r for running only)
proxmox-cli nodes stats -n <node>     # Node resource usage over a timeframe
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
  "current_context": "default",
  "contexts": {
    "default": {
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
  }
}
```

Only one of `auth_ticket` or `api_token` is normally present per context;
`api_token` wins when both are stored.

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
- Multi-cluster contexts with per-context credentials and a --context flag
- Session and API token authentication
- Cluster-wide resource overview with type, node, and status filters
- Node listing, details, storage, and task history
- Full VM and LXC lifecycle (create, start, shutdown, stop, restart, suspend, resume, delete)
- Cloning and migration with preflight checks for both guest types
- VM and LXC snapshots (create, list, rollback, delete)
- Backups: vzdump create, list, and restore with guest-type detection
- Configuration editing, disk resize, and tag management
- Guest agent integration: vm exec and IP discovery
- Interactive consoles for VMs and containers
- Resource stats for nodes, VMs, and containers (RRD-based)
- LXC template and ISO image management with server-side downloads
- Auto-assigned guest IDs on create and clone
- Shell completion with live node-name lookup
- JSON output for read commands and configurable task timeouts
- Nonzero exit statuses for operational failures
- TLS verification, custom CA support, and private config files

### Planned Features
- Firewall rule management
- User, group, and ACL administration
- HA resource management
- Bulk operations
- Configuration profiles

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for the
development setup, the checks a change must pass, and how releases work.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [go-proxmox](https://github.com/luthermonson/go-proxmox) for Proxmox API
- Testing with [Godog](https://github.com/cucumber/godog) for BDD
- Mocking with [GoMock](https://github.com/uber-go/mock) for reliable tests
- TUI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss), styled after [k9s](https://github.com/derailed/k9s)

## Support

If you encounter issues or have questions:
1. Check the [Issues](https://github.com/Adz-ai/proxmox-cli/issues) page
2. Create a new issue with detailed information
3. Include logs and configuration (redact sensitive information)

---

**Happy Proxmox management!**
