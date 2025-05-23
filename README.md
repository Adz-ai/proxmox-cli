# Proxmox CLI 🚀

A powerful command-line interface for managing Proxmox Virtual Environment (PVE) resources. Control your virtualization infrastructure from the terminal with ease!

## ✨ Features

- **🔐 Secure Authentication**: Login with username/password, stores session tokens
- **🖥️ Virtual Machine Management**: Create, delete, list, and manage VMs
- **📦 LXC Container Support**: List and manage Linux containers
- **🌐 Node Operations**: Monitor and manage cluster nodes
- **🎯 User-Friendly**: Clear error messages, helpful prompts, and guided setup
- **⚡ Fast & Efficient**: Direct API communication with Proxmox VE

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

# 2. Login (automatically configures server URL if needed)
proxmox-cli auth login -u root@pam

# 3. Check your connection
proxmox-cli status --verbose

# 4. Start using it!
proxmox-cli nodes get
proxmox-cli vm get
proxmox-cli lxc get
```

## 📚 Commands Overview

### Authentication & Setup
```bash
proxmox-cli init           # Configure server connection
proxmox-cli auth login     # Authenticate with Proxmox
proxmox-cli auth logout    # Clear authentication
proxmox-cli status         # Check configuration and connection
```

### Virtual Machine Management
```bash
proxmox-cli vm get         # List all VMs
proxmox-cli vm create      # Create VM from YAML spec
proxmox-cli vm describe    # Show VM details
proxmox-cli vm delete      # Delete a VM
```

### LXC Container Management
```bash
proxmox-cli lxc get        # List all containers
# More LXC commands coming soon!
```

### Node Management
```bash
proxmox-cli nodes get      # List cluster nodes
proxmox-cli nodes describe # Show node details
```

## 🔧 Configuration

Configuration is stored in `~/.proxmox-cli/config.json`

### Manual Configuration
```bash
proxmox-cli init
```

### Reconfigure
```bash
proxmox-cli init --force
```

## 📝 Creating VMs from YAML

Create a VM specification file:

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

Create the VM:
```bash
proxmox-cli vm create -n node1 -i 100 -s vm-spec.yaml
```

## 🧪 Development

### Running Tests
```bash
# Run all tests (mock mode by default)
go test ./...

# Run tests against real Proxmox (be careful!)
export PROXMOX_TEST_MODE=real
export PROXMOX_TEST_URL=https://your-proxmox:8006
export PROXMOX_TEST_USER=root@pam
export PROXMOX_TEST_PASS=your-password
go test ./...
```

### Building
```bash
go build -o proxmox-cli main.go
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [go-proxmox](https://github.com/luthermonson/go-proxmox) for Proxmox API
- Inspired by the need for better Proxmox CLI tools

