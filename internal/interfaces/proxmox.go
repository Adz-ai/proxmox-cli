package interfaces

import (
	"context"

	"github.com/luthermonson/go-proxmox"
)

//go:generate mockgen -destination=../../test/mocks/proxmox_client.go -package=mocks proxmox-cli/internal/interfaces ProxmoxClientInterface,NodeInterface,ContainerInterface,VirtualMachineInterface,VirtualMachineCreator,ContainerCreator

// ProxmoxClientInterface defines the interface that both real and mock clients must implement
type ProxmoxClientInterface interface {
	Nodes(ctx context.Context) (proxmox.NodeStatuses, error)
	Node(ctx context.Context, nodeName string) (NodeInterface, error)
	Version(ctx context.Context) (*proxmox.Version, error)
}

// NodeInterface defines the interface for node operations
type NodeInterface interface {
	VirtualMachines(ctx context.Context) (proxmox.VirtualMachines, error)
	Containers(ctx context.Context) (proxmox.Containers, error)
	Container(ctx context.Context, vmid int) (ContainerInterface, error)
	VirtualMachine(ctx context.Context, vmid int) (VirtualMachineInterface, error)
	NewVirtualMachine(ctx context.Context, vmid int, options ...proxmox.VirtualMachineOption) (*proxmox.Task, error)
	NewContainer(ctx context.Context, vmid int, options ...proxmox.ContainerOption) (*proxmox.Task, error)
}

// ContainerInterface defines the interface for container operations
type ContainerInterface interface {
	Start(ctx context.Context) (*proxmox.Task, error)
	Stop(ctx context.Context) (*proxmox.Task, error)
	Shutdown(ctx context.Context, force bool, timeout int) (*proxmox.Task, error)
	Reboot(ctx context.Context) (*proxmox.Task, error)
	Delete(ctx context.Context) (*proxmox.Task, error)
	Clone(ctx context.Context, options *proxmox.ContainerCloneOptions) (int, *proxmox.Task, error)
	Snapshots(ctx context.Context) ([]*proxmox.ContainerSnapshot, error)
}

// VirtualMachineInterface defines the interface for VM operations
type VirtualMachineInterface interface {
	Start(ctx context.Context) (*proxmox.Task, error)
	Stop(ctx context.Context) (*proxmox.Task, error)
	Shutdown(ctx context.Context) (*proxmox.Task, error)
	Reboot(ctx context.Context) (*proxmox.Task, error)
	Delete(ctx context.Context) (*proxmox.Task, error)
	Clone(ctx context.Context, options *proxmox.VirtualMachineCloneOptions) (int, *proxmox.Task, error)
}

// VirtualMachineCreator interface for creating VMs
type VirtualMachineCreator interface {
	Create(ctx context.Context, node NodeInterface, vmid int, options ...proxmox.VirtualMachineOption) (*proxmox.Task, error)
}

// ContainerCreator interface for creating containers
type ContainerCreator interface {
	Create(ctx context.Context, node NodeInterface, vmid int, options ...proxmox.ContainerOption) (*proxmox.Task, error)
}

