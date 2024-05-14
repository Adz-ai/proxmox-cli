# Proxmox CLI

Proxmox CLI is a command-line interface for managing Proxmox Virtual Environment (PVE) resources. It allows you to interact with your Proxmox server, manage nodes, virtual machines, and storage through various subcommands.

## Features

- **Authentication**: Login to your Proxmox server and save authentication tokens for subsequent API requests.
- **Nodes Management**: List and manage nodes in your Proxmox cluster.
- **VMs Management**: List and manage virtual machines.

## Installation

1. **Clone the repository**:

    ```sh
    git clone https://github.com/yourusername/proxmox-cli.git
    cd proxmox-cli
    ```

2. **Install dependencies**:

   Ensure you have Go installed, then run:

    ```sh
    go mod tidy
    ```

3. **Build the CLI**:

    ```sh
    go build -o proxmox-cli main.go
    ```

