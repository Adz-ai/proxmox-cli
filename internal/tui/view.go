package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	faintStyle    = lipgloss.NewStyle().Faint(true)
	activeTab     = lipgloss.NewStyle().Bold(true).Underline(true)
	headerStyle   = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Reverse(true)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	runningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	pendingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	templateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	overlayStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
)

type column struct {
	title string
	width int
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	lines := []string{m.renderTitle(), m.renderTabs()}
	switch {
	case m.confirm != nil:
		lines = append(lines, m.renderOverlay(m.confirmLines()))
	case m.details != nil:
		lines = append(lines, m.renderOverlay(m.detailLines(*m.details)))
	case m.showHelp:
		lines = append(lines, m.renderOverlay(helpLines()))
	default:
		lines = append(lines, m.renderTable()...)
	}
	lines = append(lines, m.renderStatusLine(), m.renderHints())
	clip := lipgloss.NewStyle().MaxWidth(m.width)
	for index := range lines {
		lines[index] = clip.Render(lines[index])
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderTitle() string {
	parts := []string{titleStyle.Render("proxmox-cli"), "context: " + m.contextName}
	if !m.lastSync.IsZero() {
		parts = append(parts, "refreshed "+m.lastSync.Format("15:04:05"))
	}
	if m.loading {
		parts = append(parts, "refreshing...")
	}
	return strings.Join(parts, "  |  ")
}

func (m Model) renderTabs() string {
	tabs := []string{}
	counts := map[view]int{
		viewGuests:  m.countByKind(KindVM, KindLXC),
		viewNodes:   m.countByKind(KindNode),
		viewStorage: m.countByKind(KindStorage),
	}
	for _, v := range []view{viewGuests, viewNodes, viewStorage} {
		label := fmt.Sprintf("[%d] %s (%d)", int(v)+1, v.title(), counts[v])
		if v == m.view {
			tabs = append(tabs, activeTab.Render(label))
		} else {
			tabs = append(tabs, faintStyle.Render(label))
		}
	}
	return strings.Join(tabs, "   ")
}

func (m Model) renderTable() []string {
	columns, statusIndex := m.columns()
	headerCells := make([]string, len(columns))
	for index, col := range columns {
		headerCells[index] = pad(col.title, col.width)
	}
	lines := []string{headerStyle.Render(strings.Join(headerCells, "  "))}

	rows := m.visibleRows()
	page := m.pageSize()
	end := m.offset + page
	if end > len(rows) {
		end = len(rows)
	}
	for index := m.offset; index < end; index++ {
		resource := rows[index]
		cells := m.rowCells(resource, columns)
		if index == m.cursor {
			lines = append(lines, selectedStyle.Render(strings.Join(cells, "  ")))
			continue
		}
		if style, ok := m.statusStyle(resource); ok {
			cells[statusIndex] = style.Render(cells[statusIndex])
		}
		lines = append(lines, strings.Join(cells, "  "))
	}
	if len(rows) == 0 {
		empty := "No resources"
		if m.filter != "" {
			empty = fmt.Sprintf("No matches for filter %q", m.filter)
		}
		lines = append(lines, faintStyle.Render(empty))
	}
	for len(lines) < page+1 {
		lines = append(lines, "")
	}
	return lines
}

// columns returns the column layout for the active view along with the index
// of the status column, which gets state-dependent coloring.
func (m Model) columns() ([]column, int) {
	switch m.view {
	case viewNodes:
		return []column{
			{"NAME", m.flexWidth(63, 8)},
			{"STATUS", 10},
			{"CPU", 7},
			{"CPUS", 5},
			{"MEM", 10},
			{"MAXMEM", 10},
			{"MEM%", 6},
			{"DISK%", 6},
			{"UPTIME", 9},
		}, 1
	case viewStorage:
		return []column{
			{"NAME", m.flexWidth(64, 7)},
			{"NODE", 12},
			{"STATUS", 10},
			{"TYPE", 10},
			{"USED", 10},
			{"TOTAL", 10},
			{"USE%", 6},
			{"SHARED", 6},
		}, 2
	default:
		return []column{
			{"KIND", 4},
			{"VMID", 6},
			{"NAME", m.flexWidth(66, 8)},
			{"NODE", 12},
			{"STATUS", 12},
			{"CPU", 7},
			{"MEM", 10},
			{"MEM%", 6},
			{"UPTIME", 9},
		}, 4
	}
}

// flexWidth sizes the flexible name column from the terminal width, the sum
// of the fixed column widths, and the number of fixed columns (two spaces of
// separator each).
func (m Model) flexWidth(fixed, gaps int) int {
	width := m.width - fixed - gaps*2
	if width < 12 {
		return 12
	}
	if width > 32 {
		return 32
	}
	return width
}

func (m Model) rowCells(resource Resource, columns []column) []string {
	var values []string
	switch m.view {
	case viewNodes:
		values = []string{
			resource.Name,
			m.statusText(resource),
			FormatPercent(resource.CPU),
			fmt.Sprintf("%d", resource.MaxCPU),
			HumanBytes(resource.Mem),
			HumanBytes(resource.MaxMem),
			UsagePercent(resource.Mem, resource.MaxMem),
			UsagePercent(resource.Disk, resource.MaxDisk),
			FormatUptime(resource.Uptime),
		}
	case viewStorage:
		shared := "-"
		if resource.Shared {
			shared = "yes"
		}
		values = []string{
			resource.Name,
			resource.Node,
			m.statusText(resource),
			resource.Plugin,
			HumanBytes(resource.Disk),
			HumanBytes(resource.MaxDisk),
			UsagePercent(resource.Disk, resource.MaxDisk),
			shared,
		}
	default:
		values = []string{
			string(resource.Kind),
			fmt.Sprintf("%d", resource.VMID),
			resource.Name,
			resource.Node,
			m.statusText(resource),
			FormatPercent(resource.CPU),
			HumanBytes(resource.Mem),
			UsagePercent(resource.Mem, resource.MaxMem),
			FormatUptime(resource.Uptime),
		}
	}
	cells := make([]string, len(columns))
	for index, col := range columns {
		value := "-"
		if index < len(values) && values[index] != "" {
			value = values[index]
		}
		cells[index] = pad(value, col.width)
	}
	return cells
}

func (m Model) statusText(resource Resource) string {
	if action, busy := m.inFlight[resource.ID]; busy {
		return progressVerb(action) + "..."
	}
	if resource.Template {
		return "template"
	}
	if resource.Status == "" {
		return "-"
	}
	return resource.Status
}

func (m Model) statusStyle(resource Resource) (lipgloss.Style, bool) {
	if _, busy := m.inFlight[resource.ID]; busy {
		return pendingStyle, true
	}
	if resource.Template {
		return templateStyle, true
	}
	switch resource.Status {
	case "running", "online", "available":
		return runningStyle, true
	case "stopped", "offline":
		return faintStyle, true
	case "paused", "suspended", "unknown":
		return pendingStyle, true
	default:
		return lipgloss.Style{}, false
	}
}

func (m Model) renderStatusLine() string {
	parts := []string{}
	if m.filtering || m.filter != "" {
		indicator := "filter: /" + m.filter
		if m.filtering {
			indicator += "_"
		}
		parts = append(parts, indicator)
	}
	if m.status != "" {
		message := m.status
		if m.statusIsError {
			message = errorStyle.Render(message)
		}
		parts = append(parts, message)
	}
	return strings.Join(parts, "  |  ")
}

func (m Model) renderHints() string {
	var hints string
	switch {
	case m.confirm != nil:
		hints = "y confirm · any other key cancel"
	case m.details != nil, m.showHelp:
		hints = "any key to close"
	case m.filtering:
		hints = "type to filter · enter apply · esc clear"
	default:
		hints = "j/k move · 1/2/3 view · / filter · enter details · s start · d shutdown · x stop · r reboot · R refresh · ? help · q quit"
	}
	return faintStyle.Render(hints)
}

// renderOverlay centers a bordered panel in the table area (column header
// plus body rows) so the surrounding chrome stays in place.
func (m Model) renderOverlay(lines []string) string {
	box := overlayStyle.Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.width, m.pageSize()+1, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) confirmLines() []string {
	pending := *m.confirm
	label := fmt.Sprintf("%s %d (%s) on node %s", pending.resource.Kind, pending.resource.VMID, pending.resource.Name, pending.resource.Node)
	consequence := ""
	switch pending.action {
	case ActionStop:
		consequence = "This kills the guest immediately without a clean shutdown."
	case ActionReboot:
		consequence = "The guest will be restarted."
	case ActionShutdown:
		consequence = "The guest will be asked to power off cleanly."
	}
	return []string{
		titleStyle.Render(fmt.Sprintf("Confirm %s", pending.action)),
		"",
		label,
		consequence,
		"",
		faintStyle.Render("Press y to confirm, any other key to cancel."),
	}
}

func (m Model) detailLines(resource Resource) []string {
	lines := []string{titleStyle.Render(fmt.Sprintf("%s: %s", resource.Kind, resource.Name)), ""}
	add := func(label, value string) {
		if value != "" && value != "-" {
			lines = append(lines, fmt.Sprintf("%-12s %s", label, value))
		}
	}
	if resource.VMID > 0 {
		add("VMID", fmt.Sprintf("%d", resource.VMID))
	}
	add("Node", resource.Node)
	add("Status", m.statusText(resource))
	add("HA state", resource.HAState)
	add("Tags", resource.Tags)
	if resource.Kind == KindVM || resource.Kind == KindLXC || resource.Kind == KindNode {
		add("CPU usage", FormatPercent(resource.CPU))
		if resource.MaxCPU > 0 {
			add("CPUs", fmt.Sprintf("%d", resource.MaxCPU))
		}
		add("Memory", fmt.Sprintf("%s / %s (%s)", HumanBytes(resource.Mem), HumanBytes(resource.MaxMem), UsagePercent(resource.Mem, resource.MaxMem)))
		add("Uptime", FormatUptime(resource.Uptime))
	}
	if resource.MaxDisk > 0 {
		add("Disk", fmt.Sprintf("%s / %s (%s)", HumanBytes(resource.Disk), HumanBytes(resource.MaxDisk), UsagePercent(resource.Disk, resource.MaxDisk)))
	}
	if resource.Kind == KindStorage {
		add("Type", resource.Plugin)
		shared := "no"
		if resource.Shared {
			shared = "yes"
		}
		add("Shared", shared)
	}
	return lines
}

func helpLines() []string {
	return []string{
		titleStyle.Render("Keyboard reference"),
		"",
		"Navigation",
		"  j/k, arrows    move selection",
		"  g / G          jump to top / bottom",
		"  pgup / pgdn    page up / down",
		"  1 / 2 / 3      guests / nodes / storage view",
		"  tab            cycle views",
		"  /              filter (esc clears)",
		"  enter          details for the selection",
		"  R              refresh now",
		"",
		"Guest actions",
		"  s              start",
		"  d              shutdown (graceful, confirms)",
		"  x              stop (hard, confirms)",
		"  r              reboot (confirms)",
		"",
		"  q              quit",
	}
}
