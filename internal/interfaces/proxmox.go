package interfaces

import (
	"context"

	"github.com/luthermonson/go-proxmox"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=../../test/mocks/proxmox_client.go -package=mocks github.com/Adz-ai/proxmox-cli/internal/interfaces ProxmoxClientInterface,NodeInterface,ContainerInterface,VirtualMachineInterface

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
	Delete(ctx context.Context, options *proxmox.ContainerDeleteOptions) (*proxmox.Task, error)
	Clone(ctx context.Context, options *proxmox.ContainerCloneOptions) (int, *proxmox.Task, error)
	Snapshots(ctx context.Context) ([]*proxmox.ContainerSnapshot, error)
}

// VirtualMachineInterface defines the interface for VM operations
type VirtualMachineInterface interface {
	Details() VirtualMachineDetails
	Start(ctx context.Context) (*proxmox.Task, error)
	Stop(ctx context.Context) (*proxmox.Task, error)
	Shutdown(ctx context.Context) (*proxmox.Task, error)
	Reboot(ctx context.Context) (*proxmox.Task, error)
	Delete(ctx context.Context, options *proxmox.VirtualMachineDeleteOptions) (*proxmox.Task, error)
	Clone(ctx context.Context, options *proxmox.VirtualMachineCloneOptions) (int, *proxmox.Task, error)
}

type VirtualMachineDetails struct {
	Name      string
	Node      string
	Status    string
	Tags      string
	CPUs      int
	CPU       float64
	Memory    uint64
	MaxMemory uint64
	Disk      uint64
	MaxDisk   uint64
	Uptime    uint64
}
