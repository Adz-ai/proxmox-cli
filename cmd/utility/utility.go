package utility

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/Adz-ai/proxmox-cli/internal/interfaces"

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

func (r *RealContainer) Delete(ctx context.Context, options *proxmox.ContainerDeleteOptions) (*proxmox.Task, error) {
	return r.container.Delete(ctx, options)
}

func (r *RealContainer) Clone(ctx context.Context, options *proxmox.ContainerCloneOptions) (int, *proxmox.Task, error) {
	return r.container.Clone(ctx, options)
}

func (r *RealContainer) Snapshots(ctx context.Context) ([]*proxmox.ContainerSnapshot, error) {
	return r.container.Snapshots(ctx)
}

func (r *RealContainer) NewSnapshot(ctx context.Context, name string) (*proxmox.Task, error) {
	return r.container.NewSnapshot(ctx, name)
}

func (r *RealContainer) Details() interfaces.ContainerDetails {
	return interfaces.ContainerDetails{
		Name:      r.container.Name,
		Node:      r.container.Node,
		Status:    r.container.Status,
		Tags:      r.container.Tags,
		CPUs:      r.container.CPUs,
		MaxMemory: r.container.MaxMem,
		MaxSwap:   r.container.MaxSwap,
		MaxDisk:   r.container.MaxDisk,
		Uptime:    r.container.Uptime,
	}
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

func (r *RealVirtualMachine) Details() interfaces.VirtualMachineDetails {
	return interfaces.VirtualMachineDetails{
		Name:      r.vm.Name,
		Node:      r.vm.Node,
		Status:    r.vm.Status,
		Tags:      r.vm.Tags,
		CPUs:      r.vm.CPUs,
		CPU:       r.vm.CPU,
		Memory:    r.vm.Mem,
		MaxMemory: r.vm.MaxMem,
		Disk:      r.vm.Disk,
		MaxDisk:   r.vm.MaxDisk,
		Uptime:    r.vm.Uptime,
	}
}

func (r *RealVirtualMachine) Delete(ctx context.Context, options *proxmox.VirtualMachineDeleteOptions) (*proxmox.Task, error) {
	return r.vm.Delete(ctx, options)
}

func (r *RealVirtualMachine) Clone(ctx context.Context, options *proxmox.VirtualMachineCloneOptions) (int, *proxmox.Task, error) {
	return r.vm.Clone(ctx, options)
}

func (r *RealVirtualMachine) Snapshots(ctx context.Context) ([]*proxmox.VirtualMachineSnapshot, error) {
	return r.vm.Snapshots(ctx)
}

func (r *RealVirtualMachine) NewSnapshot(ctx context.Context, name string) (*proxmox.Task, error) {
	return r.vm.NewSnapshot(ctx, name)
}

// Global variable for dependency injection (for testing)
var (
	clientFactory   func() interfaces.ProxmoxClientInterface
	clientFactoryMu sync.RWMutex
)

func CheckIfAuthPresent() error {
	// First check if the server URL is configured
	serverURL := viper.GetString("server_url")
	if serverURL == "" {
		return errors.New("not configured; run 'proxmox-cli init'")
	}

	if HasAPIToken() || HasSessionTicket() {
		return nil
	}
	return errors.New("not authenticated; run 'proxmox-cli auth login -u <username>' or 'proxmox-cli auth token -t <token-id>'")
}

// HasAPIToken reports whether an API token is stored in the configuration.
func HasAPIToken() bool {
	token := viper.Sub("api_token")
	return token != nil && token.GetString("token_id") != "" && token.GetString("secret") != ""
}

// HasSessionTicket reports whether a session ticket is stored in the configuration.
func HasSessionTicket() bool {
	authTicket := viper.Sub("auth_ticket")
	return authTicket != nil && authTicket.GetString("ticket") != "" && authTicket.GetString("CSRFPreventionToken") != ""
}

// ClearAuthTicket blanks stored session credentials in viper's state. Viper
// cannot delete keys, and an empty replacement map does not shadow values
// already loaded from the config file, so every existing subkey must be
// overwritten individually for WriteConfig to persist the cleared state.
func ClearAuthTicket() {
	clearCredentialSection("auth_ticket")
}

// ClearAPIToken blanks any stored API token; see ClearAuthTicket for the
// viper key-shadowing details.
func ClearAPIToken() {
	clearCredentialSection("api_token")
}

func clearCredentialSection(section string) {
	viper.Set(section, map[string]any{})
	for _, key := range viper.AllKeys() {
		if strings.HasPrefix(key, section+".") {
			viper.Set(key, "")
		}
	}
}

// GetClient returns a Proxmox client (real or mock depending on test context)
func GetClient() (interfaces.ProxmoxClientInterface, error) {
	clientFactoryMu.RLock()
	factory := clientFactory
	clientFactoryMu.RUnlock()
	if factory != nil {
		return factory(), nil
	}

	endpoint := viper.GetString("server_url")
	if endpoint == "" {
		return nil, errors.New("server URL is not configured")
	}

	normalizedEndpoint, err := NormalizeServerURL(endpoint)
	if err != nil {
		return nil, err
	}
	httpClient, err := NewHTTPClient(viper.GetBool("insecure"), viper.GetString("ca_cert"))
	if err != nil {
		return nil, err
	}
	apiEndpoint := normalizedEndpoint + "/api2/json"

	var realClient *proxmox.Client
	if HasAPIToken() {
		token := viper.Sub("api_token")
		realClient = proxmox.NewClient(apiEndpoint,
			proxmox.WithHTTPClient(httpClient),
			proxmox.WithAPIToken(token.GetString("token_id"), token.GetString("secret")))
	} else if HasSessionTicket() {
		authTicket := viper.Sub("auth_ticket")
		realClient = proxmox.NewClient(apiEndpoint,
			proxmox.WithHTTPClient(httpClient),
			proxmox.WithSession(authTicket.GetString("ticket"), authTicket.GetString("CSRFPreventionToken")))
	}

	if realClient == nil {
		realClient = proxmox.NewClient(apiEndpoint, proxmox.WithHTTPClient(httpClient))
	}

	return &RealProxmoxClient{client: realClient}, nil
}

func AuthenticatedClient() (interfaces.ProxmoxClientInterface, error) {
	if err := CheckIfAuthPresent(); err != nil {
		return nil, err
	}
	return GetClient()
}

// SetClientFactory sets a factory function for creating clients (used by tests)
func SetClientFactory(factory func() interfaces.ProxmoxClientInterface) {
	clientFactoryMu.Lock()
	defer clientFactoryMu.Unlock()
	clientFactory = factory
}

// ResetClientFactory resets the client factory to use real clients
func ResetClientFactory() {
	clientFactoryMu.Lock()
	defer clientFactoryMu.Unlock()
	clientFactory = nil
}

func NormalizeServerURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("server URL cannot be empty")
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported server URL scheme %q; HTTPS is required", parsed.Scheme)
	}
	if parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("server URL must contain only a scheme, host, port, and optional path")
	}
	parsed.Path = strings.TrimSuffix(strings.TrimSuffix(parsed.Path, "/"), "/api2/json")
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	return parsed.String(), nil
}

func NewHTTPClient(insecure bool, caCertPath string) (*http.Client, error) {
	if insecure && caCertPath != "" {
		return nil, errors.New("insecure TLS and a custom CA certificate are mutually exclusive")
	}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("default HTTP transport has an unexpected type")
	}
	transport = transport.Clone()
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecure}
	if caCertPath != "" {
		pem, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("read CA certificate: %w", err)
		}
		pool, err := x509.SystemCertPool()
		if err != nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, errors.New("CA certificate file contains no valid PEM certificates")
		}
		tlsConfig.RootCAs = pool
	}
	transport.TLSClientConfig = tlsConfig
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}, nil
}

func ConfigFile() (string, error) {
	if configured := os.Getenv("PROXMOX_CLI_CONFIG"); configured != "" {
		return configured, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".proxmox-cli", "config.json"), nil
}

func LoadConfig() error {
	path, err := ConfigFile()
	if err != nil {
		return err
	}
	viper.SetConfigFile(path)
	viper.SetConfigType("json")
	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("read configuration: %w", err)
		}
	} else if err == nil {
		if err := os.Chmod(path, 0o600); err != nil {
			return fmt.Errorf("secure configuration: %w", err)
		}
	}
	return nil
}

// normalizeCredentials tidies the credential sections before they are
// persisted: entries blanked by ClearAuthTicket/ClearAPIToken are dropped
// (removing a section entirely once no credentials remain), and the CSRF
// token key viper lowercases internally is restored to its documented casing.
func normalizeCredentials(settings map[string]any) {
	for _, section := range []string{"auth_ticket", "api_token"} {
		credentials, ok := settings[section].(map[string]any)
		if !ok {
			continue
		}
		for key, value := range credentials {
			if text, ok := value.(string); ok && text == "" {
				delete(credentials, key)
			}
		}
		if value, ok := credentials["csrfpreventiontoken"]; ok {
			delete(credentials, "csrfpreventiontoken")
			credentials["CSRFPreventionToken"] = value
		}
		if len(credentials) == 0 {
			delete(settings, section)
		}
	}
}

func WriteConfig() error {
	path, err := ConfigFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create configuration directory: %w", err)
	}
	settings := viper.AllSettings()
	normalizeCredentials(settings)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode configuration: %w", err)
	}
	data = append(data, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary configuration: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure temporary configuration: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write configuration: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync configuration: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close configuration: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace configuration: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("secure configuration: %w", err)
	}
	viper.SetConfigFile(path)
	return nil
}

// AddOutputFlag registers the shared --output flag on a read command.
func AddOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "Output format: table or json")
}

// OutputFormat validates and returns the shared --output flag.
func OutputFormat(cmd *cobra.Command) (string, error) {
	format, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", fmt.Errorf("read output flag: %w", err)
	}
	switch format {
	case "table", "json":
		return format, nil
	default:
		return "", fmt.Errorf("unsupported output format %q; use table or json", format)
	}
}

// PrintJSON writes v to out as indented JSON.
func PrintJSON(out io.Writer, v any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("encode JSON output: %w", err)
	}
	return nil
}

// DefaultTaskTimeout bounds how long commands wait for a Proxmox task when
// the user does not override it with --timeout.
const DefaultTaskTimeout = 10 * time.Minute

// TaskTimeout returns the value of the root --timeout flag, falling back to
// the default when the flag is missing or not positive.
func TaskTimeout(cmd *cobra.Command) time.Duration {
	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil || timeout <= 0 {
		return DefaultTaskTimeout
	}
	return timeout
}

func WaitForTask(ctx context.Context, task *proxmox.Task, timeout time.Duration) error {
	if task == nil {
		return errors.New("no task returned by Proxmox")
	}
	if !task.IsSuccessful && !task.IsFailed {
		seconds := int(math.Ceil(timeout.Seconds()))
		if seconds < 1 {
			seconds = 1
		}
		if err := task.WaitFor(ctx, seconds); err != nil {
			return fmt.Errorf("wait for task %s: %w", task.UPID, err)
		}
	}
	if !task.IsSuccessful {
		return fmt.Errorf("task %s failed: %s", task.UPID, task.ExitStatus)
	}
	return nil
}
