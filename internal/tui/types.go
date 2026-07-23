// Package tui implements the interactive terminal UI (k9s-style) for
// browsing and managing Proxmox cluster resources.
package tui

import "context"

// Kind identifies the type of a cluster resource row.
type Kind string

const (
	KindVM      Kind = "vm"
	KindLXC     Kind = "lxc"
	KindNode    Kind = "node"
	KindStorage Kind = "storage"
)

// Action is a lifecycle operation that can be applied to a guest.
type Action string

const (
	ActionStart    Action = "start"
	ActionShutdown Action = "shutdown"
	ActionStop     Action = "stop"
	ActionReboot   Action = "reboot"
)

// Resource is one row of cluster state, normalized from the Proxmox
// cluster/resources endpoint.
type Resource struct {
	Kind     Kind
	ID       string
	VMID     uint64
	Name     string
	Node     string
	Status   string
	CPU      float64 // fraction of one core-set, 0..1
	MaxCPU   uint64
	Mem      uint64
	MaxMem   uint64
	Disk     uint64
	MaxDisk  uint64
	Uptime   uint64
	Tags     string
	Template bool
	HAState  string
	Shared   bool
	Plugin   string
}

// DataSource provides cluster state and guest actions to the UI. The real
// implementation wraps the Proxmox client; tests substitute a fake.
type DataSource interface {
	Resources(ctx context.Context) ([]Resource, error)
	Guest(ctx context.Context, resource Resource, action Action) error
	Version(ctx context.Context) (string, error)
}
