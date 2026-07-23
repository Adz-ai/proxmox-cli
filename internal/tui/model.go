package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	viewGuests view = iota
	viewNodes
	viewStorage
	viewCount
)

func (v view) title() string {
	switch v {
	case viewGuests:
		return "Guests"
	case viewNodes:
		return "Nodes"
	case viewStorage:
		return "Storage"
	default:
		return "Unknown"
	}
}

func (v view) includes(kind Kind) bool {
	switch v {
	case viewGuests:
		return kind == KindVM || kind == KindLXC
	case viewNodes:
		return kind == KindNode
	case viewStorage:
		return kind == KindStorage
	default:
		return false
	}
}

type resourcesMsg struct {
	rows []Resource
	err  error
}

type versionMsg struct {
	version string
}

type actionDoneMsg struct {
	resource Resource
	action   Action
	err      error
}

type tickMsg time.Time

type confirmState struct {
	resource Resource
	action   Action
}

// Model is the bubbletea model behind the TUI.
type Model struct {
	source      DataSource
	contextName string
	server      string
	user        string
	cliVersion  string
	pveVersion  string
	refresh     time.Duration

	width  int
	height int

	view       view
	kindFilter Kind
	rows       []Resource
	cursor     int
	offset     int
	filter     string
	filtering  bool

	commandMode bool
	command     string

	confirm  *confirmState
	details  *Resource
	showHelp bool

	loading       bool
	inFlight      map[string]Action
	status        string
	statusIsError bool
	lastSync      time.Time
}

// NewModel builds the initial model. Refresh defaults to five seconds.
func NewModel(opts Options) Model {
	refresh := opts.Refresh
	if refresh <= 0 {
		refresh = 5 * time.Second
	}
	return Model{
		source:      opts.Source,
		contextName: opts.ContextName,
		server:      opts.Server,
		user:        opts.User,
		cliVersion:  opts.CLIVersion,
		refresh:     refresh,
		loading:     true,
		inFlight:    map[string]Action{},
		status:      "connecting...",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchCmd(), m.versionCmd(), m.tickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampCursor()
		return m, nil
	case versionMsg:
		m.pveVersion = msg.version
		return m, nil
	case resourcesMsg:
		m.loading = false
		if msg.err != nil {
			m.setStatus(fmt.Sprintf("refresh failed: %v", msg.err), true)
			return m, nil
		}
		m.lastSync = time.Now()
		if m.statusIsError {
			m.setStatus("", false)
		}
		selected := m.selectedID()
		m.rows = sortResources(msg.rows)
		m.restoreSelection(selected)
		return m, nil
	case actionDoneMsg:
		delete(m.inFlight, msg.resource.ID)
		label := fmt.Sprintf("%s %d (%s)", msg.resource.Kind, msg.resource.VMID, msg.resource.Name)
		if msg.err != nil {
			m.setStatus(fmt.Sprintf("%s %s failed: %v", msg.action, label, msg.err), true)
			return m, nil
		}
		m.setStatus(fmt.Sprintf("%s: %s", label, doneVerb(msg.action)), false)
		m.loading = true
		return m, m.fetchCmd()
	case tickMsg:
		cmds := []tea.Cmd{m.tickCmd()}
		if !m.loading {
			m.loading = true
			cmds = append(cmds, m.fetchCmd())
		}
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return m, tea.Quit
	}
	if m.confirm != nil {
		pending := *m.confirm
		m.confirm = nil
		if key == "y" {
			return m.startAction(pending.resource, pending.action)
		}
		m.setStatus("cancelled", false)
		return m, nil
	}
	if m.details != nil {
		m.details = nil
		return m, nil
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}
	if m.filtering {
		return m.handleFilterKey(msg)
	}
	if m.commandMode {
		return m.handleCommandKey(msg)
	}
	switch key {
	case "q":
		return m, tea.Quit
	case "1":
		m.switchView(viewGuests)
	case "2":
		m.switchView(viewNodes)
	case "3":
		m.switchView(viewStorage)
	case "tab":
		m.switchView((m.view + 1) % viewCount)
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "pgup":
		m.moveCursor(-m.pageSize())
	case "pgdown":
		m.moveCursor(m.pageSize())
	case "home", "g":
		m.cursor = 0
		m.clampCursor()
	case "end", "G":
		m.cursor = len(m.visibleRows()) - 1
		m.clampCursor()
	case "/":
		m.filtering = true
	case ":":
		m.commandMode = true
		m.command = ""
	case "esc":
		m.filter = ""
		m.kindFilter = ""
		m.clampCursor()
	case "enter":
		if resource, ok := m.selectedResource(); ok {
			m.details = &resource
		}
	case "?":
		m.showHelp = true
	case "R":
		if !m.loading {
			m.loading = true
			m.setStatus("refreshing...", false)
			return m, m.fetchCmd()
		}
	case "s":
		return m.requestAction(ActionStart)
	case "d":
		return m.requestAction(ActionShutdown)
	case "x":
		return m.requestAction(ActionStop)
	case "r":
		return m.requestAction(ActionReboot)
	}
	return m, nil
}

func (m Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filtering = false
	case "esc":
		m.filtering = false
		m.filter = ""
	case "backspace":
		if runes := []rune(m.filter); len(runes) > 0 {
			m.filter = string(runes[:len(runes)-1])
		}
	default:
		switch msg.Type {
		case tea.KeyRunes:
			m.filter += string(msg.Runes)
		case tea.KeySpace:
			m.filter += " "
		}
	}
	m.clampCursor()
	return m, nil
}

func (m Model) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.commandMode = false
		return m.executeCommand(strings.ToLower(strings.TrimSpace(m.command)))
	case "esc":
		m.commandMode = false
		m.command = ""
	case "backspace":
		if runes := []rune(m.command); len(runes) > 0 {
			m.command = string(runes[:len(runes)-1])
		}
	default:
		switch msg.Type {
		case tea.KeyRunes:
			m.command += string(msg.Runes)
		case tea.KeySpace:
			m.command += " "
		}
	}
	return m, nil
}

// executeCommand dispatches a k9s-style ":" command.
func (m Model) executeCommand(command string) (tea.Model, tea.Cmd) {
	switch command {
	case "":
	case "q", "quit", "q!":
		return m, tea.Quit
	case "guests", "guest", "all":
		m.switchView(viewGuests)
	case "vm", "vms", "qemu":
		m.switchView(viewGuests)
		m.kindFilter = KindVM
		m.clampCursor()
	case "lxc", "ct", "container", "containers":
		m.switchView(viewGuests)
		m.kindFilter = KindLXC
		m.clampCursor()
	case "nodes", "node", "no":
		m.switchView(viewNodes)
	case "storage", "st", "sto":
		m.switchView(viewStorage)
	case "help", "h":
		m.showHelp = true
	default:
		m.setStatus(fmt.Sprintf("invalid command %q", command), true)
	}
	return m, nil
}

func (m Model) requestAction(action Action) (tea.Model, tea.Cmd) {
	resource, ok := m.selectedResource()
	if !ok {
		return m, nil
	}
	if resource.Kind != KindVM && resource.Kind != KindLXC {
		m.setStatus("lifecycle actions apply to VMs and containers only", true)
		return m, nil
	}
	if resource.Template {
		m.setStatus("templates do not support lifecycle actions", true)
		return m, nil
	}
	if pending, busy := m.inFlight[resource.ID]; busy {
		m.setStatus(fmt.Sprintf("%s %d already has a %s in progress", resource.Kind, resource.VMID, pending), true)
		return m, nil
	}
	if action == ActionStart {
		return m.startAction(resource, action)
	}
	m.confirm = &confirmState{resource: resource, action: action}
	return m, nil
}

func (m Model) startAction(resource Resource, action Action) (tea.Model, tea.Cmd) {
	m.inFlight[resource.ID] = action
	m.setStatus(fmt.Sprintf("%s %s %d (%s)...", progressVerb(action), resource.Kind, resource.VMID, resource.Name), false)
	source := m.source
	return m, func() tea.Msg {
		err := source.Guest(context.Background(), resource, action)
		return actionDoneMsg{resource: resource, action: action, err: err}
	}
}

func (m Model) fetchCmd() tea.Cmd {
	source := m.source
	return func() tea.Msg {
		rows, err := source.Resources(context.Background())
		return resourcesMsg{rows: rows, err: err}
	}
}

func (m Model) versionCmd() tea.Cmd {
	source := m.source
	return func() tea.Msg {
		version, err := source.Version(context.Background())
		if err != nil {
			return versionMsg{version: "n/a"}
		}
		return versionMsg{version: version}
	}
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(m.refresh, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) setStatus(message string, isError bool) {
	m.status = message
	m.statusIsError = isError
}

func (m *Model) switchView(target view) {
	if m.view == target {
		m.kindFilter = ""
		return
	}
	m.view = target
	m.cursor = 0
	m.offset = 0
	m.filter = ""
	m.filtering = false
	m.kindFilter = ""
}

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	m.clampCursor()
}

func (m *Model) clampCursor() {
	count := len(m.visibleRows())
	if m.cursor >= count {
		m.cursor = count - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	page := m.pageSize()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+page {
		m.offset = m.cursor - page + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

// pageSize is the number of table rows that fit in the current terminal:
// total height minus the header block, the table border and column header,
// the crumbs line, and the command prompt when it is open.
func (m *Model) pageSize() int {
	chrome := headerHeight + 3 + 1
	if m.commandMode {
		chrome++
	}
	page := m.height - chrome
	if page < 1 {
		page = 1
	}
	return page
}

func (m *Model) visibleRows() []Resource {
	rows := make([]Resource, 0, len(m.rows))
	for _, resource := range m.rows {
		if !m.view.includes(resource.Kind) {
			continue
		}
		if m.kindFilter != "" && resource.Kind != m.kindFilter {
			continue
		}
		if m.filter != "" && !matchesFilter(resource, m.filter) {
			continue
		}
		rows = append(rows, resource)
	}
	return rows
}

func (m *Model) selectedResource() (Resource, bool) {
	rows := m.visibleRows()
	if m.cursor < 0 || m.cursor >= len(rows) {
		return Resource{}, false
	}
	return rows[m.cursor], true
}

func (m *Model) selectedID() string {
	if resource, ok := m.selectedResource(); ok {
		return resource.ID
	}
	return ""
}

func (m *Model) restoreSelection(id string) {
	if id != "" {
		for index, resource := range m.visibleRows() {
			if resource.ID == id {
				m.cursor = index
				break
			}
		}
	}
	m.clampCursor()
}

func matchesFilter(resource Resource, filter string) bool {
	needle := strings.ToLower(strings.TrimSpace(filter))
	if needle == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		string(resource.Kind),
		fmt.Sprintf("%d", resource.VMID),
		resource.Name,
		resource.Node,
		resource.Status,
		resource.Tags,
	}, " "))
	return strings.Contains(haystack, needle)
}

func sortResources(rows []Resource) []Resource {
	sorted := make([]Resource, len(rows))
	copy(sorted, rows)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if kindOrder(a.Kind) != kindOrder(b.Kind) {
			return kindOrder(a.Kind) < kindOrder(b.Kind)
		}
		if a.VMID != b.VMID {
			return a.VMID < b.VMID
		}
		if a.Node != b.Node {
			return a.Node < b.Node
		}
		return a.Name < b.Name
	})
	return sorted
}

// kindOrder groups guests together so the guests view sorts by VMID across
// both VM and LXC rows.
func kindOrder(kind Kind) int {
	switch kind {
	case KindVM, KindLXC:
		return 0
	case KindNode:
		return 1
	default:
		return 2
	}
}

func progressVerb(action Action) string {
	switch action {
	case ActionStart:
		return "starting"
	case ActionShutdown:
		return "shutting down"
	case ActionStop:
		return "stopping"
	case ActionReboot:
		return "rebooting"
	default:
		return string(action)
	}
}

func doneVerb(action Action) string {
	switch action {
	case ActionStart:
		return "started"
	case ActionShutdown:
		return "shut down"
	case ActionStop:
		return "stopped"
	case ActionReboot:
		return "rebooted"
	default:
		return string(action) + " completed"
	}
}
