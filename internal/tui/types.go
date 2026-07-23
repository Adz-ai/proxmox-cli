// Package tui implements the interactive terminal UI (k9s-style) for
// browsing and managing Proxmox cluster resources.
package tui

import (
	"context"
	"io"
)

// Kind identifies the type of a cluster resource row.
type Kind string

const (
	KindVM      Kind = "vm"
	KindLXC     Kind = "lxc"
	KindNode    Kind = "node"
	KindStorage Kind = "storage"
	KindTask    Kind = "task"
)

// Action is a lifecycle operation that can be applied to a guest.
type Action string

const (
	ActionStart    Action = "start"
	ActionShutdown Action = "shutdown"
	ActionStop     Action = "stop"
	ActionReboot   Action = "reboot"
	ActionDelete   Action = "delete"
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

	// Task rows only.
	Target string
	User   string
	Start  int64
	End    int64
}

// Snapshot is one guest snapshot, shown in the snapshots overlay.
type Snapshot struct {
	Name        string
	Parent      string
	Description string
	Created     int64
}

// ShellSession runs an interactive console attached to the given streams,
// blocking until the session ends. stdin must be an interactive terminal.
type ShellSession func(stdin io.Reader, stdout, stderr io.Writer) error

// DataSource provides cluster state and guest actions to the UI. The real
// implementation wraps the Proxmox client; tests substitute a fake.
type DataSource interface {
	Resources(ctx context.Context) ([]Resource, error)
	Guest(ctx context.Context, resource Resource, action Action) error
	Version(ctx context.Context) (string, error)
	Tasks(ctx context.Context) ([]Resource, error)
	Snapshots(ctx context.Context, resource Resource) ([]Snapshot, error)
	Shell(resource Resource) (ShellSession, error)
}
