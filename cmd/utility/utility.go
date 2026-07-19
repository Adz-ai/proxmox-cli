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

// RealCluster wraps the actual go-proxmox cluster
type RealCluster struct {
	cluster *proxmox.Cluster
}

func (r *RealProxmoxClient) Cluster(ctx context.Context) (interfaces.ClusterInterface, error) {
	cluster, err := r.client.Cluster(ctx)
	if err != nil {
		return nil, err
	}
	return &RealCluster{cluster: cluster}, nil
}

func (r *RealCluster) Resources(ctx context.Context, filters ...string) (proxmox.ClusterResources, error) {
	return r.cluster.Resources(ctx, filters...)
}

func (r *RealCluster) NextID(ctx context.Context) (int, error) {
	return r.cluster.NextID(ctx)
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

func (r *RealNode) Storages(ctx context.Context) (proxmox.Storages, error) {
	return r.node.Storages(ctx)
}

// RealStorage wraps the actual go-proxmox storage
type RealStorage struct {
	storage *proxmox.Storage
}

func (r *RealNode) Storage(ctx context.Context, name string) (interfaces.StorageInterface, error) {
	storage, err := r.node.Storage(ctx, name)
	if err != nil {
		return nil, err
	}
	return &RealStorage{storage: storage}, nil
}

func (r *RealStorage) GetContent(ctx context.Context) ([]*proxmox.StorageContent, error) {
	return r.storage.GetContent(ctx)
}

func (r *RealNode) Tasks(ctx context.Context, options *proxmox.NodeTasksOptions) ([]*proxmox.Task, error) {
	return r.node.Tasks(ctx, options)
}

func (r *RealNode) Vzdump(ctx context.Context, options *proxmox.VirtualMachineBackupOptions) (*proxmox.Task, error) {
	return r.node.Vzdump(ctx, options)
}

func (r *RealNode) Appliances(ctx context.Context) (proxmox.Appliances, error) {
	return r.node.Appliances(ctx)
}

func (r *RealNode) DownloadAppliance(ctx context.Context, template, storage string) (string, error) {
	return r.node.DownloadAppliance(ctx, template, storage)
}

func (r *RealNode) VzTmpls(ctx context.Context, storage string) (proxmox.VzTmpls, error) {
	return r.node.VzTmpls(ctx, storage)
}

func (r *RealNode) StorageDownloadURL(ctx context.Context, options *proxmox.StorageDownloadURLOptions) (string, error) {
	return r.node.StorageDownloadURL(ctx, options)
}

func (r *RealNode) RRDData(ctx context.Context, timeframe proxmox.Timeframe, cf proxmox.ConsolidationFunction) ([]*proxmox.RRDData, error) {
	return r.node.RRDData(ctx, timeframe, cf)
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

func (r *RealContainer) RollbackSnapshot(ctx context.Context, name string, start bool) (*proxmox.Task, error) {
	return r.container.Snapshot(name).Rollback(ctx, start)
}

func (r *RealContainer) DeleteSnapshot(ctx context.Context, name string) (*proxmox.Task, error) {
	return r.container.Snapshot(name).Delete(ctx)
}

func (r *RealContainer) Suspend(ctx context.Context) (*proxmox.Task, error) {
	return r.container.Suspend(ctx)
}

func (r *RealContainer) Resume(ctx context.Context) (*proxmox.Task, error) {
	return r.container.Resume(ctx)
}

func (r *RealContainer) Migrate(ctx context.Context, options *proxmox.ContainerMigrateOptions) (*proxmox.Task, error) {
	return r.container.Migrate(ctx, options)
}

func (r *RealContainer) Config(ctx context.Context, options ...proxmox.ContainerOption) (*proxmox.Task, error) {
	return r.container.Config(ctx, options...)
}

func (r *RealContainer) Resize(ctx context.Context, disk, size string) (*proxmox.Task, error) {
	return r.container.Resize(ctx, disk, size)
}

func (r *RealContainer) AddTag(ctx context.Context, value string) (*proxmox.Task, error) {
	return r.container.AddTag(ctx, value)
}

func (r *RealContainer) RemoveTag(ctx context.Context, value string) (*proxmox.Task, error) {
	return r.container.RemoveTag(ctx, value)
}

func (r *RealContainer) Interfaces(ctx context.Context) (proxmox.ContainerInterfaces, error) {
	return r.container.Interfaces(ctx)
}

func (r *RealContainer) RRDData(ctx context.Context, timeframe proxmox.Timeframe, cf ...proxmox.ConsolidationFunction) ([]*proxmox.RRDData, error) {
	return r.container.RRDData(ctx, timeframe, cf...)
}

func (r *RealContainer) TermProxy(ctx context.Context) (*proxmox.Term, error) {
	return r.container.TermProxy(ctx)
}

func (r *RealContainer) TermWebSocket(term *proxmox.Term) (chan []byte, chan []byte, chan error, func() error, error) {
	return r.container.TermWebSocket(term)
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

func (r *RealVirtualMachine) RollbackSnapshot(ctx context.Context, name string) (*proxmox.Task, error) {
	return r.vm.Snapshot(name).Rollback(ctx)
}

func (r *RealVirtualMachine) DeleteSnapshot(ctx context.Context, name string) (*proxmox.Task, error) {
	return r.vm.Snapshot(name).Delete(ctx)
}

func (r *RealVirtualMachine) Pause(ctx context.Context) (*proxmox.Task, error) {
	return r.vm.Pause(ctx)
}

func (r *RealVirtualMachine) Resume(ctx context.Context) (*proxmox.Task, error) {
	return r.vm.Resume(ctx)
}

func (r *RealVirtualMachine) Migrate(ctx context.Context, options *proxmox.VirtualMachineMigrateOptions) (*proxmox.Task, error) {
	return r.vm.Migrate(ctx, options)
}

func (r *RealVirtualMachine) MigratePreconditions(ctx context.Context, target string) (*proxmox.VirtualMachineMigratePreconditions, error) {
	return r.vm.MigratePreconditions(ctx, target)
}

func (r *RealVirtualMachine) Config(ctx context.Context, options ...proxmox.VirtualMachineOption) (*proxmox.Task, error) {
	return r.vm.Config(ctx, options...)
}

func (r *RealVirtualMachine) ResizeDisk(ctx context.Context, disk, size string) (*proxmox.Task, error) {
	return r.vm.ResizeDisk(ctx, disk, size)
}

func (r *RealVirtualMachine) AddTag(ctx context.Context, value string) (*proxmox.Task, error) {
	return r.vm.AddTag(ctx, value)
}

func (r *RealVirtualMachine) RemoveTag(ctx context.Context, value string) (*proxmox.Task, error) {
	return r.vm.RemoveTag(ctx, value)
}

func (r *RealVirtualMachine) WaitForAgent(ctx context.Context, seconds int) error {
	return r.vm.WaitForAgent(ctx, seconds)
}

func (r *RealVirtualMachine) AgentExec(ctx context.Context, command []string, inputData string) (int, error) {
	return r.vm.AgentExec(ctx, command, inputData)
}

func (r *RealVirtualMachine) WaitForAgentExecExit(ctx context.Context, pid, seconds int) (*proxmox.AgentExecStatus, error) {
	return r.vm.WaitForAgentExecExit(ctx, pid, seconds)
}

func (r *RealVirtualMachine) AgentGetNetworkIFaces(ctx context.Context) ([]*proxmox.AgentNetworkIface, error) {
	return r.vm.AgentGetNetworkIFaces(ctx)
}

func (r *RealVirtualMachine) RRDData(ctx context.Context, timeframe proxmox.Timeframe, cf ...proxmox.ConsolidationFunction) ([]*proxmox.RRDData, error) {
	return r.vm.RRDData(ctx, timeframe, cf...)
}

func (r *RealVirtualMachine) TermProxy(ctx context.Context) (*proxmox.Term, error) {
	return r.vm.TermProxy(ctx)
}

func (r *RealVirtualMachine) TermWebSocket(term *proxmox.Term) (chan []byte, chan []byte, chan error, func() error, error) {
	return r.vm.TermWebSocket(term)
}

// Global variable for dependency injection (for testing)
var (
	clientFactory   func() interfaces.ProxmoxClientInterface
	clientFactoryMu sync.RWMutex
)

func CheckIfAuthPresent() error {
	// First check if the server URL is configured
	serverURL := ContextString("server_url")
	if serverURL == "" {
		return fmt.Errorf("context %q is not configured; run 'proxmox-cli init'", ActiveContext())
	}

	if HasAPIToken() || HasSessionTicket() {
		return nil
	}
	return errors.New("not authenticated; run 'proxmox-cli auth login -u <username>' or 'proxmox-cli auth token -t <token-id>'")
}

// HasAPIToken reports whether an API token is stored for the active context.
func HasAPIToken() bool {
	return ContextString("api_token.token_id") != "" && ContextString("api_token.secret") != ""
}

// HasSessionTicket reports whether a session ticket is stored for the active context.
func HasSessionTicket() bool {
	return ContextString("auth_ticket.ticket") != "" && ContextString("auth_ticket.CSRFPreventionToken") != ""
}

// ClearAuthTicket blanks stored session credentials in viper's state. Viper
// cannot delete keys, and an empty replacement map does not shadow values
// already loaded from the config file, so every existing subkey must be
// overwritten individually for WriteConfig to persist the cleared state.
func ClearAuthTicket() {
	clearCredentialSection(contextKey("auth_ticket"))
	if ActiveContext() == DefaultContext {
		clearCredentialSection("auth_ticket")
	}
}

// ClearAPIToken blanks any stored API token; see ClearAuthTicket for the
// viper key-shadowing details.
func ClearAPIToken() {
	clearCredentialSection(contextKey("api_token"))
	if ActiveContext() == DefaultContext {
		clearCredentialSection("api_token")
	}
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

	endpoint := ContextString("server_url")
	if endpoint == "" {
		return nil, errors.New("server URL is not configured")
	}

	normalizedEndpoint, err := NormalizeServerURL(endpoint)
	if err != nil {
		return nil, err
	}
	httpClient, err := NewHTTPClient(ContextBool("insecure"), ContextString("ca_cert"))
	if err != nil {
		return nil, err
	}
	apiEndpoint := normalizedEndpoint + "/api2/json"

	var realClient *proxmox.Client
	if HasAPIToken() {
		realClient = proxmox.NewClient(apiEndpoint,
			proxmox.WithHTTPClient(httpClient),
			proxmox.WithAPIToken(ContextString("api_token.token_id"), ContextString("api_token.secret")))
	} else if HasSessionTicket() {
		realClient = proxmox.NewClient(apiEndpoint,
			proxmox.WithHTTPClient(httpClient),
			proxmox.WithSession(ContextString("auth_ticket.ticket"), ContextString("auth_ticket.CSRFPreventionToken")))
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

// normalizeCredentials tidies the credential sections of one context map:
// entries blanked by ClearAuthTicket/ClearAPIToken are dropped (removing a
// section entirely once no credentials remain), and the CSRF token key viper
// lowercases internally is restored to its documented casing.
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

// migrateLegacySettings folds a pre-context flat configuration into the
// contexts map so the persisted file always uses the current layout. Legacy
// keys merge into their target context without overwriting newer values.
func migrateLegacySettings(settings map[string]any) {
	legacyKeys := []string{"server_url", "insecure", "ca_cert", "auth_ticket", "api_token"}
	if serverURL, _ := settings["server_url"].(string); serverURL != "" {
		target := DefaultContext
		if current, _ := settings["current_context"].(string); current != "" {
			target = current
		}
		contexts, ok := settings["contexts"].(map[string]any)
		if !ok {
			contexts = map[string]any{}
			settings["contexts"] = contexts
		}
		contextMap, ok := contexts[target].(map[string]any)
		if !ok {
			contextMap = map[string]any{}
			contexts[target] = contextMap
		}
		for _, key := range legacyKeys {
			if value, exists := settings[key]; exists {
				if _, taken := contextMap[key]; !taken {
					contextMap[key] = value
				}
			}
		}
		if _, ok := settings["current_context"]; !ok {
			settings["current_context"] = target
		}
	}
	for _, key := range legacyKeys {
		delete(settings, key)
	}
}

// normalizeSettings prepares the full settings map for persistence:
// legacy layouts are migrated, credentials are tidied per context, and
// contexts without a server URL (the deletion mechanism) are dropped.
func normalizeSettings(settings map[string]any) {
	migrateLegacySettings(settings)
	contexts, ok := settings["contexts"].(map[string]any)
	if !ok {
		delete(settings, "contexts")
		delete(settings, "current_context")
		return
	}
	for name, raw := range contexts {
		contextMap, ok := raw.(map[string]any)
		if !ok {
			delete(contexts, name)
			continue
		}
		for key, value := range contextMap {
			if text, ok := value.(string); ok && text == "" {
				delete(contextMap, key)
			}
		}
		normalizeCredentials(contextMap)
		if serverURL, _ := contextMap["server_url"].(string); serverURL == "" {
			delete(contexts, name)
		}
	}
	if len(contexts) == 0 {
		delete(settings, "contexts")
		delete(settings, "current_context")
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
	normalizeSettings(settings)
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
	_ = cmd.RegisterFlagCompletionFunc("output", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// RegisterNodeFlagCompletion wires dynamic node-name completion for the
// given flag by querying the cluster. Completion silently degrades when the
// CLI is not authenticated or the server is unreachable.
func RegisterNodeFlagCompletion(cmd *cobra.Command, flag string) {
	_ = cmd.RegisterFlagCompletionFunc(flag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := AuthenticatedClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()
		nodes, err := client.Nodes(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := make([]string, 0, len(nodes))
		for _, node := range nodes {
			names = append(names, node.Node)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	})
}

// ResolveVMID returns id unchanged when positive, or asks the cluster for
// the next free guest ID when id is zero.
func ResolveVMID(ctx context.Context, client interfaces.ProxmoxClientInterface, id int) (int, error) {
	if id > 0 {
		return id, nil
	}
	if id < 0 {
		return 0, errors.New("ID must be positive")
	}
	cluster, err := client.Cluster(ctx)
	if err != nil {
		return 0, fmt.Errorf("get cluster: %w", err)
	}
	next, err := cluster.NextID(ctx)
	if err != nil {
		return 0, fmt.Errorf("get next free ID: %w", err)
	}
	return next, nil
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

// RRDSummary condenses a series of RRD samples into latest/average/peak
// figures for display.
type RRDSummary struct {
	Timeframe        string  `json:"timeframe"`
	Samples          int     `json:"samples"`
	LatestCPU        float64 `json:"latest_cpu_percent"`
	AverageCPU       float64 `json:"average_cpu_percent"`
	PeakCPU          float64 `json:"peak_cpu_percent"`
	LatestMemory     uint64  `json:"latest_memory_bytes"`
	AverageMemory    uint64  `json:"average_memory_bytes"`
	PeakMemory       uint64  `json:"peak_memory_bytes"`
	MaxMemory        uint64  `json:"max_memory_bytes,omitempty"`
	AverageNetIn     float64 `json:"average_net_in_bps"`
	AverageNetOut    float64 `json:"average_net_out_bps"`
	AverageDiskRead  float64 `json:"average_disk_read_bps"`
	AverageDiskWrite float64 `json:"average_disk_write_bps"`
}

// SummarizeRRD aggregates RRD samples, skipping gaps the RRD reports as NaN.
func SummarizeRRD(timeframe string, samples []*proxmox.RRDData) RRDSummary {
	summary := RRDSummary{Timeframe: timeframe}
	var cpuSum, memSum, netInSum, netOutSum, diskReadSum, diskWriteSum float64
	for _, sample := range samples {
		if sample == nil || math.IsNaN(sample.CPU) || math.IsNaN(sample.Mem) {
			continue
		}
		summary.Samples++
		cpu := sample.CPU * 100
		cpuSum += cpu
		memSum += sample.Mem
		if !math.IsNaN(sample.NetIn) {
			netInSum += sample.NetIn
		}
		if !math.IsNaN(sample.NetOut) {
			netOutSum += sample.NetOut
		}
		if !math.IsNaN(sample.DiskRead) {
			diskReadSum += sample.DiskRead
		}
		if !math.IsNaN(sample.DiskWrite) {
			diskWriteSum += sample.DiskWrite
		}
		if cpu > summary.PeakCPU {
			summary.PeakCPU = cpu
		}
		if uint64(sample.Mem) > summary.PeakMemory {
			summary.PeakMemory = uint64(sample.Mem)
		}
		summary.LatestCPU = cpu
		summary.LatestMemory = uint64(sample.Mem)
		if sample.MaxMem > 0 {
			summary.MaxMemory = sample.MaxMem
		}
	}
	if summary.Samples > 0 {
		count := float64(summary.Samples)
		summary.AverageCPU = cpuSum / count
		summary.AverageMemory = uint64(memSum / count)
		summary.AverageNetIn = netInSum / count
		summary.AverageNetOut = netOutSum / count
		summary.AverageDiskRead = diskReadSum / count
		summary.AverageDiskWrite = diskWriteSum / count
	}
	return summary
}

// PrintRRDSummary renders an RRD summary as a small table.
func PrintRRDSummary(out io.Writer, subject string, summary RRDSummary) {
	const gib = 1024 * 1024 * 1024
	const mib = 1024 * 1024
	fmt.Fprintf(out, "Stats for %s (%s):\n", subject, summary.Timeframe)
	fmt.Fprintf(out, "Samples: %d\n", summary.Samples)
	if summary.Samples == 0 {
		fmt.Fprintln(out, "No data available for this timeframe")
		return
	}
	fmt.Fprintf(out, "CPU:    latest %.1f%%   avg %.1f%%   peak %.1f%%\n",
		summary.LatestCPU, summary.AverageCPU, summary.PeakCPU)
	memory := fmt.Sprintf("Memory: latest %.2f GiB   avg %.2f GiB   peak %.2f GiB",
		float64(summary.LatestMemory)/gib, float64(summary.AverageMemory)/gib, float64(summary.PeakMemory)/gib)
	if summary.MaxMemory > 0 {
		memory += fmt.Sprintf(" / %.2f GiB", float64(summary.MaxMemory)/gib)
	}
	fmt.Fprintln(out, memory)
	fmt.Fprintf(out, "Net:    in %.2f MiB/s   out %.2f MiB/s (avg)\n",
		summary.AverageNetIn/mib, summary.AverageNetOut/mib)
	fmt.Fprintf(out, "Disk:   read %.2f MiB/s   write %.2f MiB/s (avg)\n",
		summary.AverageDiskRead/mib, summary.AverageDiskWrite/mib)
}

// ParseTimeframe validates the shared --timeframe flag value.
func ParseTimeframe(value string) (proxmox.Timeframe, error) {
	switch value {
	case "hour", "day", "week", "month", "year":
		return proxmox.Timeframe(value), nil
	default:
		return "", fmt.Errorf("unsupported timeframe %q; use hour, day, week, month, or year", value)
	}
}

// AddTimeframeFlag registers the shared --timeframe flag on a stats command.
func AddTimeframeFlag(cmd *cobra.Command) {
	cmd.Flags().String("timeframe", "hour", "Sampling window: hour, day, week, month, or year")
	_ = cmd.RegisterFlagCompletionFunc("timeframe", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"hour", "day", "week", "month", "year"}, cobra.ShellCompDirectiveNoFileComp
	})
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
	if task.IsFailed {
		return fmt.Errorf("task %s failed: %s", task.UPID, task.ExitStatus)
	}
	if task.IsSuccessful {
		return nil
	}
	// Synchronous API responses (e.g. container config updates) carry no
	// task to poll; the change is already applied.
	if task.UPID == "" {
		return nil
	}
	seconds := int(math.Ceil(timeout.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	if err := task.WaitFor(ctx, seconds); err != nil {
		return fmt.Errorf("wait for task %s: %w", task.UPID, err)
	}
	if !task.IsSuccessful {
		return fmt.Errorf("task %s failed: %s", task.UPID, task.ExitStatus)
	}
	return nil
}
