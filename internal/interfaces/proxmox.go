package interfaces

import (
	"context"

	"github.com/luthermonson/go-proxmox"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=../../test/mocks/proxmox_client.go -package=mocks github.com/Adz-ai/proxmox-cli/internal/interfaces ProxmoxClientInterface,NodeInterface,ContainerInterface,VirtualMachineInterface,ClusterInterface

// ProxmoxClientInterface defines the interface that both real and mock clients must implement
type ProxmoxClientInterface interface {
	Nodes(ctx context.Context) (proxmox.NodeStatuses, error)
	Node(ctx context.Context, nodeName string) (NodeInterface, error)
	Version(ctx context.Context) (*proxmox.Version, error)
	Cluster(ctx context.Context) (ClusterInterface, error)
}

// ClusterInterface defines the interface for cluster-level operations
type ClusterInterface interface {
	Resources(ctx context.Context, filters ...string) (proxmox.ClusterResources, error)
	NextID(ctx context.Context) (int, error)
}

// NodeInterface defines the interface for node operations
type NodeInterface interface {
	VirtualMachines(ctx context.Context) (proxmox.VirtualMachines, error)
	Containers(ctx context.Context) (proxmox.Containers, error)
	Container(ctx context.Context, vmid int) (ContainerInterface, error)
	VirtualMachine(ctx context.Context, vmid int) (VirtualMachineInterface, error)
	NewVirtualMachine(ctx context.Context, vmid int, options ...proxmox.VirtualMachineOption) (*proxmox.Task, error)
	NewContainer(ctx context.Context, vmid int, options ...proxmox.ContainerOption) (*proxmox.Task, error)
	Storages(ctx context.Context) (proxmox.Storages, error)
	Tasks(ctx context.Context, options *proxmox.NodeTasksOptions) ([]*proxmox.Task, error)
}

// ContainerInterface defines the interface for container operations
type ContainerInterface interface {
	Details() ContainerDetails
	Start(ctx context.Context) (*proxmox.Task, error)
	Stop(ctx context.Context) (*proxmox.Task, error)
	Shutdown(ctx context.Context, force bool, timeout int) (*proxmox.Task, error)
	Reboot(ctx context.Context) (*proxmox.Task, error)
	Delete(ctx context.Context, options *proxmox.ContainerDeleteOptions) (*proxmox.Task, error)
	Clone(ctx context.Context, options *proxmox.ContainerCloneOptions) (int, *proxmox.Task, error)
	Snapshots(ctx context.Context) ([]*proxmox.ContainerSnapshot, error)
	NewSnapshot(ctx context.Context, name string) (*proxmox.Task, error)
	RollbackSnapshot(ctx context.Context, name string, start bool) (*proxmox.Task, error)
	DeleteSnapshot(ctx context.Context, name string) (*proxmox.Task, error)
	Suspend(ctx context.Context) (*proxmox.Task, error)
	Resume(ctx context.Context) (*proxmox.Task, error)
}

type ContainerDetails struct {
	Name      string `json:"name"`
	Node      string `json:"node"`
	Status    string `json:"status"`
	Tags      string `json:"tags,omitempty"`
	CPUs      int    `json:"cpus"`
	MaxMemory uint64 `json:"max_memory_bytes"`
	MaxSwap   uint64 `json:"max_swap_bytes"`
	MaxDisk   uint64 `json:"max_disk_bytes"`
	Uptime    uint64 `json:"uptime_seconds"`
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
	Snapshots(ctx context.Context) ([]*proxmox.VirtualMachineSnapshot, error)
	NewSnapshot(ctx context.Context, name string) (*proxmox.Task, error)
	RollbackSnapshot(ctx context.Context, name string) (*proxmox.Task, error)
	DeleteSnapshot(ctx context.Context, name string) (*proxmox.Task, error)
	Pause(ctx context.Context) (*proxmox.Task, error)
	Resume(ctx context.Context) (*proxmox.Task, error)
}

type VirtualMachineDetails struct {
	Name      string  `json:"name"`
	Node      string  `json:"node"`
	Status    string  `json:"status"`
	Tags      string  `json:"tags,omitempty"`
	CPUs      int     `json:"cpus"`
	CPU       float64 `json:"cpu_usage"`
	Memory    uint64  `json:"memory_bytes"`
	MaxMemory uint64  `json:"max_memory_bytes"`
	Disk      uint64  `json:"disk_bytes"`
	MaxDisk   uint64  `json:"max_disk_bytes"`
	Uptime    uint64  `json:"uptime_seconds"`
}
