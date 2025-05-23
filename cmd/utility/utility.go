package utility

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"

	"proxmox-cli/internal/interfaces"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
)

// RealProxmoxClient wraps the actual go-proxmox client
type RealProxmoxClient struct {
	client *proxmox.Client
}

// RealNode wraps the actual go-proxmox node
type RealNode struct {
	node *proxmox.Node
}

// RealContainer wraps the actual go-proxmox container
type RealContainer struct {
	container *proxmox.Container
}

// RealVirtualMachine wraps the actual go-proxmox VM
type RealVirtualMachine struct {
	vm *proxmox.VirtualMachine
}

// Implement the interfaces for real types
func (r *RealProxmoxClient) Nodes(ctx context.Context) (proxmox.NodeStatuses, error) {
	return r.client.Nodes(ctx)
}

func (r *RealProxmoxClient) Node(ctx context.Context, nodeName string) (interfaces.NodeInterface, error) {
	node, err := r.client.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	return &RealNode{node: node}, nil
}

func (r *RealProxmoxClient) Version(ctx context.Context) (*proxmox.Version, error) {
	return r.client.Version(ctx)
}

func (r *RealNode) VirtualMachines(ctx context.Context) (proxmox.VirtualMachines, error) {
	return r.node.VirtualMachines(ctx)
}

func (r *RealNode) Containers(ctx context.Context) (proxmox.Containers, error) {
	return r.node.Containers(ctx)
}

func (r *RealNode) Container(ctx context.Context, vmid int) (interfaces.ContainerInterface, error) {
	container, err := r.node.Container(ctx, vmid)
	if err != nil {
		return nil, err
	}
	return &RealContainer{container: container}, nil
}

func (r *RealNode) VirtualMachine(ctx context.Context, vmid int) (interfaces.VirtualMachineInterface, error) {
	vm, err := r.node.VirtualMachine(ctx, vmid)
	if err != nil {
		return nil, err
	}
	return &RealVirtualMachine{vm: vm}, nil
}

func (r *RealNode) NewVirtualMachine(ctx context.Context, vmid int, options ...proxmox.VirtualMachineOption) (*proxmox.Task, error) {
	return r.node.NewVirtualMachine(ctx, vmid, options...)
}

func (r *RealNode) NewContainer(ctx context.Context, vmid int, options ...proxmox.ContainerOption) (*proxmox.Task, error) {
	return r.node.NewContainer(ctx, vmid, options...)
}


func (r *RealContainer) Start(ctx context.Context) (*proxmox.Task, error) {
	return r.container.Start(ctx)
}

func (r *RealContainer) Stop(ctx context.Context) (*proxmox.Task, error) {
	return r.container.Stop(ctx)
}

func (r *RealContainer) Shutdown(ctx context.Context, force bool, timeout int) (*proxmox.Task, error) {
	return r.container.Shutdown(ctx, force, timeout)
}

func (r *RealContainer) Reboot(ctx context.Context) (*proxmox.Task, error) {
	return r.container.Reboot(ctx)
}

func (r *RealContainer) Delete(ctx context.Context) (*proxmox.Task, error) {
	return r.container.Delete(ctx)
}

func (r *RealContainer) Clone(ctx context.Context, options *proxmox.ContainerCloneOptions) (int, *proxmox.Task, error) {
	return r.container.Clone(ctx, options)
}

func (r *RealContainer) Snapshots(ctx context.Context) ([]*proxmox.ContainerSnapshot, error) {
	return r.container.Snapshots(ctx)
}

func (r *RealVirtualMachine) Start(ctx context.Context) (*proxmox.Task, error) {
	return r.vm.Start(ctx)
}

func (r *RealVirtualMachine) Stop(ctx context.Context) (*proxmox.Task, error) {
	return r.vm.Stop(ctx)
}

func (r *RealVirtualMachine) Shutdown(ctx context.Context) (*proxmox.Task, error) {
	return r.vm.Shutdown(ctx)
}

func (r *RealVirtualMachine) Reboot(ctx context.Context) (*proxmox.Task, error) {
	return r.vm.Reboot(ctx)
}

func (r *RealVirtualMachine) Delete(ctx context.Context) (*proxmox.Task, error) {
	return r.vm.Delete(ctx)
}

func (r *RealVirtualMachine) Clone(ctx context.Context, options *proxmox.VirtualMachineCloneOptions) (int, *proxmox.Task, error) {
	return r.vm.Clone(ctx, options)
}

// Global variable for dependency injection (for testing)
var clientFactory func() interfaces.ProxmoxClientInterface

func CheckIfAuthPresent() error {
	// First check if the server URL is configured
	serverURL := viper.GetString("server_url")
	if serverURL == "" {
		return errors.New("❌ Not configured. Please run 'proxmox-cli auth login -u <username>' to set up")
	}

	// Check if the client is authenticated
	authTicket := viper.Sub("auth_ticket")
	if authTicket == nil || authTicket.GetString("ticket") == "" || authTicket.GetString("CSRFPreventionToken") == "" {
		return errors.New("❌ Not authenticated. Please run 'proxmox-cli auth login -u <username>' to log in")
	}
	return nil
}

// GetClient returns a Proxmox client (real or mock depending on test context)
func GetClient() interfaces.ProxmoxClientInterface {
	// If we have a mock factory (from tests), use it
	if clientFactory != nil {
		return clientFactory()
	}

	// Otherwise, create real client
	endpoint := viper.GetString("server_url")
	if endpoint == "" {
		log.Fatal("❌ Proxmox server URL not configured. Please run 'proxmox-cli auth login -u <username>'")
	}

	// Create HTTP client with TLS config
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Create Proxmox client with session
	authTicket := viper.Sub("auth_ticket")
	var realClient *proxmox.Client
	if authTicket != nil {
		ticket := authTicket.GetString("ticket")
		csrfToken := authTicket.GetString("CSRFPreventionToken")
		if ticket != "" && csrfToken != "" {
			// Use WithSession option to set auth
			realClient = proxmox.NewClient(endpoint+"/api2/json",
				proxmox.WithHTTPClient(httpClient),
				proxmox.WithSession(ticket, csrfToken))
		}
	}

	if realClient == nil {
		// Return the client without auth if no session found
		realClient = proxmox.NewClient(endpoint+"/api2/json", proxmox.WithHTTPClient(httpClient))
	}

	return &RealProxmoxClient{client: realClient}
}

// SetClientFactory sets a factory function for creating clients (used by tests)
func SetClientFactory(factory func() interfaces.ProxmoxClientInterface) {
	clientFactory = factory
}

// ResetClientFactory resets the client factory to use real clients
func ResetClientFactory() {
	clientFactory = nil
}