package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// headerHeight is the k9s-style header block: cluster info, menu hints, logo.
const headerHeight = 7

// The palette mirrors the k9s default skin.
var (
	colorAqua       = lipgloss.Color("#00ffff")
	colorBlue       = lipgloss.Color("#1e90ff") // dodgerblue
	colorOrange     = lipgloss.Color("#ffa500")
	colorFuchsia    = lipgloss.Color("#ff00ff")
	colorPapaya     = lipgloss.Color("#ffefd5") // papayawhip
	colorSeagreen   = lipgloss.Color("#2e8b57")
	colorLawnGreen  = lipgloss.Color("#7cfc00")
	colorTurquoise  = lipgloss.Color("#00ced1") // darkturquoise
	colorOrangeRed  = lipgloss.Color("#ff4500")
	colorDarkOrange = lipgloss.Color("#ff8c00")
	colorSlateGray  = lipgloss.Color("#778899") // lightslategray
	colorPurple     = lipgloss.Color("#9370db") // mediumpurple
	colorSteelBlue  = lipgloss.Color("#4682b4")
	colorPaleGreen  = lipgloss.Color("#98fb98")
	colorWhite      = lipgloss.Color("#ffffff")
	colorBlack      = lipgloss.Color("#000000")

	sectionStyle = lipgloss.NewStyle().Foreground(colorWhite).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(colorOrange)
	revStyle     = lipgloss.NewStyle().Foreground(colorAqua)
	cpuStyle     = lipgloss.NewStyle().Foreground(colorLawnGreen)
	memStyle     = lipgloss.NewStyle().Foreground(colorTurquoise)
	logoStyle    = lipgloss.NewStyle().Foreground(colorOrange).Bold(true)

	menuKeyStyle  = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	menuNumStyle  = lipgloss.NewStyle().Foreground(colorFuchsia).Bold(true)
	menuDescStyle = lipgloss.NewStyle().Foreground(colorWhite)

	borderStyle      = lipgloss.NewStyle().Foreground(colorBlue)
	titleStyle       = lipgloss.NewStyle().Foreground(colorAqua).Bold(true)
	titleDecorStyle  = lipgloss.NewStyle().Foreground(colorAqua)
	titleScopeStyle  = lipgloss.NewStyle().Foreground(colorFuchsia).Bold(true)
	titleCountStyle  = lipgloss.NewStyle().Foreground(colorPapaya).Bold(true)
	titleFilterStyle = lipgloss.NewStyle().Foreground(colorSeagreen).Bold(true)
	tableHeaderStyle = lipgloss.NewStyle().Foreground(colorWhite).Bold(true)
	cursorStyle      = lipgloss.NewStyle().Foreground(colorBlack).Background(colorAqua).Bold(true)
	rowStyleDefault  = lipgloss.NewStyle().Foreground(colorAqua)
	rowStyleStopped  = lipgloss.NewStyle().Foreground(colorSlateGray)
	rowStylePending  = lipgloss.NewStyle().Foreground(colorDarkOrange)
	rowStyleError    = lipgloss.NewStyle().Foreground(colorOrangeRed)
	rowStyleTemplate = lipgloss.NewStyle().Foreground(colorPurple)

	crumbStyle       = lipgloss.NewStyle().Foreground(colorBlack).Background(colorAqua).Bold(true)
	crumbActiveStyle = lipgloss.NewStyle().Foreground(colorBlack).Background(colorOrange).Bold(true)
	flashErrorStyle  = lipgloss.NewStyle().Foreground(colorOrangeRed)
	flashInfoStyle   = lipgloss.NewStyle().Foreground(colorPaleGreen)

	promptStyle     = lipgloss.NewStyle().Foreground(colorAqua).Bold(true)
	promptTextStyle = lipgloss.NewStyle().Foreground(colorWhite)

	describeKeyStyle   = lipgloss.NewStyle().Foreground(colorSteelBlue).Bold(true)
	describeValueStyle = lipgloss.NewStyle().Foreground(colorPapaya)
	helpSectionStyle   = lipgloss.NewStyle().Foreground(colorAqua).Bold(true)
	dimStyle           = lipgloss.NewStyle().Foreground(colorSlateGray)
)

// logoLines is "PVE" in the same block style as the k9s logo.
var logoLines = []string{
	` _____  __      __ ______ `,
	`|  __ \ \ \    / /|  ____|`,
	`| |__) | \ \  / / | |__   `,
	`|  ___/   \ \/ /  |  __|  `,
	`| |        \  /   | |____ `,
	`|_|         \/    |______|`,
}

type column struct {
	title string
	width int
	sort  string
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	if m.width < 70 || m.height < 16 {
		return "proxmox-cli: terminal too small (need at least 70x16)"
	}
	lines := m.renderHeader()
	if m.commandMode {
		lines = append(lines, m.renderPrompt())
	}
	lines = append(lines, m.renderBody()...)
	lines = append(lines, m.renderCrumbs())
	clip := lipgloss.NewStyle().MaxWidth(m.width)
	for index := range lines {
		lines[index] = clip.Render(lines[index])
	}
	return strings.Join(lines, "\n")
}

// renderHeader lays out the cluster info block, the menu hints, and the logo
// side by side, like the k9s header.
func (m Model) renderHeader() []string {
	info := m.infoLines()
	menu := menuLines()
	logo := make([]string, len(logoLines))
	for index, line := range logoLines {
		logo[index] = logoStyle.Render(line)
	}

	infoWidth := blockWidth(info)
	menuWidth := blockWidth(menu)
	logoWidth := blockWidth(logo)
	showMenu := m.width >= infoWidth+menuWidth+4
	showLogo := m.width >= infoWidth+menuWidth+logoWidth+8

	lines := make([]string, headerHeight)
	for index := range lines {
		line := padVisible(blockLine(info, index), infoWidth)
		if showMenu {
			line += "  " + padVisible(blockLine(menu, index), menuWidth)
		}
		if showLogo {
			gap := m.width - lipgloss.Width(line) - logoWidth - 1
			if gap > 0 {
				line += strings.Repeat(" ", gap) + blockLine(logo, index)
			}
		}
		lines[index] = line
	}
	return lines
}

func (m Model) infoLines() []string {
	pve := m.pveVersion
	if pve == "" {
		pve = "n/a"
	}
	rows := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"Context:", m.contextName, infoStyle},
		{"Cluster:", m.server, infoStyle},
		{"User:", m.user, infoStyle},
		{"CLI Rev:", m.cliVersion, revStyle},
		{"PVE Rev:", pve, revStyle},
		{"CPU:", m.clusterCPU(), cpuStyle},
		{"MEM:", m.clusterMEM(), memStyle},
	}
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		value := row.value
		if value == "" {
			value = "n/a"
		}
		lines = append(lines, sectionStyle.Render(pad(row.label, 9))+row.style.Render(truncate(value, 32)))
	}
	return lines
}

// clusterCPU is the load across all online nodes, weighted by core count.
func (m Model) clusterCPU() string {
	var used, cores float64
	for _, resource := range m.rows {
		if resource.Kind == KindNode {
			used += resource.CPU * float64(resource.MaxCPU)
			cores += float64(resource.MaxCPU)
		}
	}
	if cores == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.0f%%", used/cores*100)
}

func (m Model) clusterMEM() string {
	var used, total uint64
	for _, resource := range m.rows {
		if resource.Kind == KindNode {
			used += resource.Mem
			total += resource.MaxMem
		}
	}
	if total == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.0f%%", float64(used)/float64(total)*100)
}

type menuHint struct {
	key     string
	desc    string
	numeric bool
}

func menuLines() []string {
	columns := [][]menuHint{
		{
			{"1", "guests", true},
			{"2", "nodes", true},
			{"3", "storage", true},
			{"4", "tasks", true},
			{"/", "filter", false},
			{":", "command", false},
		},
		{
			{"s", "start", false},
			{"d", "shutdown", false},
			{"x", "stop", false},
			{"r", "reboot", false},
			{"c", "console", false},
			{"ctrl-d", "delete", false},
		},
		{
			{"t", "snapshots", false},
			{"enter", "describe", false},
			{"shft-x", "sort", false},
			{"R", "refresh", false},
			{"?", "help", false},
			{"q", "quit", false},
		},
	}
	lines := make([]string, headerHeight)
	for row := range lines {
		parts := make([]string, 0, len(columns))
		for _, col := range columns {
			if row >= len(col) {
				parts = append(parts, strings.Repeat(" ", 19))
				continue
			}
			hint := col[row]
			keyStyle := menuKeyStyle
			if hint.numeric {
				keyStyle = menuNumStyle
			}
			parts = append(parts, keyStyle.Render(pad("<"+hint.key+">", 9))+menuDescStyle.Render(pad(hint.desc, 10)))
		}
		lines[row] = strings.Join(parts, "")
	}
	return lines
}

func (m Model) renderPrompt() string {
	return " " + promptStyle.Render("> ") + promptTextStyle.Render(m.command) + cursorStyle.Render(" ")
}

// renderBody draws the bordered table box, or an overlay box in its place.
func (m Model) renderBody() []string {
	innerHeight := m.pageSize() + 1
	switch {
	case m.confirm != nil:
		return m.renderBox(m.confirmTitle(), m.confirmLines(), innerHeight)
	case m.details != nil:
		return m.renderBox(m.describeTitle(*m.details), m.detailLines(*m.details), innerHeight)
	case m.snapshotsFor != nil:
		return m.renderBox(m.snapshotsTitle(), m.snapshotLines(), innerHeight)
	case m.showHelp:
		return m.renderBox(boxTitle("Help", "", 0, false), helpLines(), innerHeight)
	default:
		return m.renderBox(m.tableTitle(), m.tableLines(), innerHeight)
	}
}

func (m Model) snapshotsTitle() string {
	resource := *m.snapshotsFor
	scope := fmt.Sprintf("%s/%d", resource.Kind, resource.VMID)
	return boxTitle("Snapshots", scope, len(m.snapshots), true)
}

func (m Model) snapshotLines() []string {
	if m.snapshots == nil {
		return []string{"", dimStyle.Render("Loading snapshots...")}
	}
	lines := []string{tableHeaderStyle.Render(pad("NAME", 24) + "  " + pad("CREATED", 20) + "  " + pad("PARENT", 24) + "  DESCRIPTION")}
	for _, snapshot := range m.snapshots {
		created := "-"
		if snapshot.Created > 0 {
			created = FormatUnixTime(snapshot.Created)
		}
		parent := snapshot.Parent
		if parent == "" {
			parent = "-"
		}
		description := strings.TrimSpace(snapshot.Description)
		lines = append(lines, rowStyleDefault.Render(
			pad(snapshot.Name, 24)+"  "+pad(created, 20)+"  "+pad(parent, 24)+"  "+description))
	}
	if len(m.snapshots) == 0 {
		lines = append(lines, dimStyle.Render("No snapshots"))
	}
	lines = append(lines, "", dimStyle.Render("Press any key to close."))
	return lines
}

// renderBox draws a k9s-style bordered panel with the title embedded in the
// top border.
func (m Model) renderBox(title string, content []string, innerHeight int) []string {
	innerWidth := m.width - 4
	titleWidth := lipgloss.Width(title)
	fill := m.width - 3 - titleWidth
	if fill < 0 {
		fill = 0
	}
	lines := make([]string, 0, innerHeight+2)
	lines = append(lines, borderStyle.Render("┌─")+title+borderStyle.Render(strings.Repeat("─", fill)+"┐"))
	edge := borderStyle.Render("│")
	clip := lipgloss.NewStyle().MaxWidth(innerWidth)
	for index := 0; index < innerHeight; index++ {
		line := ""
		if index < len(content) {
			line = clip.Render(content[index])
		}
		lines = append(lines, edge+" "+padVisible(line, innerWidth)+" "+edge)
	}
	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", m.width-2)+"┘"))
	return lines
}

// boxTitle renders " Name(scope)[count] " with k9s title colors. The scope
// segment is omitted when scope is empty and withCount controls the counter.
func boxTitle(name, scope string, count int, withCount bool) string {
	title := " " + titleStyle.Render(name)
	if scope != "" {
		title += titleDecorStyle.Render("(") + titleScopeStyle.Render(scope) + titleDecorStyle.Render(")")
	}
	if withCount {
		title += titleDecorStyle.Render("[") + titleCountStyle.Render(fmt.Sprintf("%d", count)) + titleDecorStyle.Render("]")
	}
	return title + " "
}

func (m Model) tableTitle() string {
	scope := ""
	if m.view == viewGuests {
		scope = "all"
		if m.kindFilter != "" {
			scope = string(m.kindFilter)
		}
	}
	title := boxTitle(m.view.title(), scope, len(m.visibleRows()), true)
	if m.filtering || m.filter != "" {
		segment := "/" + m.filter
		if m.filtering {
			segment += "_"
		}
		title += titleDecorStyle.Render("<") + titleFilterStyle.Render(segment) + titleDecorStyle.Render("> ")
	}
	if m.loading {
		title += dimStyle.Render("refreshing... ")
	}
	return title
}

func (m Model) describeTitle(resource Resource) string {
	scope := string(resource.Kind) + "/" + resource.Name
	if resource.VMID > 0 {
		scope = fmt.Sprintf("%s/%d", resource.Kind, resource.VMID)
	}
	return boxTitle("Describe", scope, 0, false)
}

func (m Model) confirmTitle() string {
	return boxTitle("Confirm", string(m.confirm.action), 0, false)
}

func (m Model) tableLines() []string {
	columns := m.columns()
	headerCells := make([]string, len(columns))
	for index, col := range columns {
		title := col.title
		if col.sort != "" && col.sort == m.sortKey {
			arrow := "↑"
			if !m.sortAsc {
				arrow = "↓"
			}
			title += arrow
		}
		headerCells[index] = pad(title, col.width)
	}
	lines := []string{tableHeaderStyle.Render(strings.Join(headerCells, "  "))}

	rows := m.visibleRows()
	page := m.pageSize()
	end := m.offset + page
	if end > len(rows) {
		end = len(rows)
	}
	innerWidth := m.width - 4
	for index := m.offset; index < end; index++ {
		resource := rows[index]
		cells := m.rowCells(resource, columns)
		row := pad(strings.Join(cells, "  "), innerWidth)
		if index == m.cursor {
			lines = append(lines, cursorStyle.Render(row))
			continue
		}
		lines = append(lines, m.rowStyle(resource).Render(row))
	}
	if len(rows) == 0 {
		empty := "No resources"
		if m.filter != "" {
			empty = fmt.Sprintf("No matches for filter %q", m.filter)
		}
		lines = append(lines, dimStyle.Render(empty))
	}
	return lines
}

// rowStyle colors a whole row by guest state, like k9s status-based rows.
// Completed tasks render dim so running tasks and failures stand out.
func (m Model) rowStyle(resource Resource) lipgloss.Style {
	if resource.Kind == KindTask {
		switch resource.Status {
		case "running":
			return rowStylePending
		case "OK":
			return rowStyleStopped
		default:
			return rowStyleError
		}
	}
	if _, busy := m.inFlight[resource.ID]; busy {
		return rowStylePending
	}
	if resource.Template {
		return rowStyleTemplate
	}
	switch resource.Status {
	case "running", "online", "available":
		return rowStyleDefault
	case "stopped", "offline":
		return rowStyleStopped
	case "paused", "suspended":
		return rowStylePending
	case "unknown", "error":
		return rowStyleError
	default:
		return rowStyleDefault
	}
}

// columns returns the column layout for the active view, sized against the
// box interior. The sort field names the semantic sort key for the column.
func (m Model) columns() []column {
	inner := m.width - 4
	switch m.view {
	case viewNodes:
		return []column{
			{"NAME", flexWidth(inner, 63, 8), "name"},
			{"STATUS", 10, "status"},
			{"CPU", 7, "cpu"},
			{"CPUS", 5, ""},
			{"MEM", 10, "mem"},
			{"MAXMEM", 10, ""},
			{"MEM%", 6, ""},
			{"DISK%", 6, "used"},
			{"UPTIME", 9, "age"},
		}
	case viewStorage:
		return []column{
			{"NAME", flexWidth(inner, 64, 7), "name"},
			{"NODE", 12, "node"},
			{"STATUS", 10, "status"},
			{"TYPE", 10, ""},
			{"USED", 10, "used"},
			{"TOTAL", 10, "total"},
			{"USE%", 6, ""},
			{"SHARED", 6, ""},
		}
	case viewTasks:
		return []column{
			{"STARTED", 15, "start"},
			{"TYPE", flexWidth(inner, 82, 6), "type"},
			{"TARGET", 10, "target"},
			{"NODE", 12, "node"},
			{"USER", 18, "user"},
			{"DURATION", 9, ""},
			{"STATUS", 18, "status"},
		}
	default:
		return []column{
			{"KIND", 4, ""},
			{"VMID", 6, "id"},
			{"NAME", flexWidth(inner, 66, 8), "name"},
			{"NODE", 12, "node"},
			{"STATUS", 12, "status"},
			{"CPU", 7, "cpu"},
			{"MEM", 10, "mem"},
			{"MEM%", 6, ""},
			{"UPTIME", 9, "age"},
		}
	}
}

// flexWidth sizes the flexible name column from the available width, the sum
// of the fixed column widths, and the number of fixed columns (two spaces of
// separator each).
func flexWidth(available, fixed, gaps int) int {
	width := available - fixed - gaps*2
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
	case viewTasks:
		duration := "-"
		if resource.End > resource.Start && resource.Start > 0 {
			duration = FormatUptime(uint64(resource.End - resource.Start))
		}
		values = []string{
			FormatUnixTime(resource.Start),
			resource.Name,
			resource.Target,
			resource.Node,
			resource.User,
			duration,
			resource.Status,
		}
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

// renderCrumbs draws the k9s breadcrumb chips with the flash message on the
// right.
func (m Model) renderCrumbs() string {
	crumbs := []string{strings.ToLower(m.view.title())}
	switch {
	case m.confirm != nil:
		crumbs = append(crumbs, "confirm")
	case m.details != nil:
		crumbs = append(crumbs, "describe")
	case m.snapshotsFor != nil:
		crumbs = append(crumbs, "snapshots")
	case m.showHelp:
		crumbs = append(crumbs, "help")
	}
	parts := make([]string, 0, len(crumbs))
	for index, crumb := range crumbs {
		style := crumbStyle
		if index == len(crumbs)-1 {
			style = crumbActiveStyle
		}
		parts = append(parts, style.Render(" <"+crumb+"> "))
	}
	line := " " + strings.Join(parts, " ")
	if m.status != "" {
		flash := flashInfoStyle
		if m.statusIsError {
			flash = flashErrorStyle
		}
		// Truncate long messages rather than dropping them so errors
		// always surface.
		available := m.width - lipgloss.Width(line) - 2
		if available > 8 {
			message := flash.Render(truncate(m.status, available))
			gap := m.width - lipgloss.Width(line) - lipgloss.Width(message) - 1
			if gap > 0 {
				line += strings.Repeat(" ", gap) + message
			}
		}
	}
	return line
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
		"",
		menuDescStyle.Render(fmt.Sprintf("Confirm %s of %s", pending.action, label)),
		rowStyleError.Render(consequence),
		"",
		dimStyle.Render("Press y to confirm, any other key to cancel."),
	}
}

func (m Model) detailLines(resource Resource) []string {
	lines := []string{""}
	add := func(label, value string) {
		if value != "" && value != "-" {
			lines = append(lines, describeKeyStyle.Render(pad(label, 12))+describeValueStyle.Render(value))
		}
	}
	if resource.Kind == KindTask {
		add("UPID", resource.ID)
		add("Type", resource.Name)
		add("Target", resource.Target)
		add("Node", resource.Node)
		add("User", resource.User)
		add("Status", resource.Status)
		add("Started", FormatUnixTime(resource.Start))
		if resource.End > resource.Start && resource.Start > 0 {
			add("Duration", FormatUptime(uint64(resource.End-resource.Start)))
		}
		lines = append(lines, "", dimStyle.Render("Press any key to close."))
		return lines
	}
	if resource.VMID > 0 {
		add("VMID", fmt.Sprintf("%d", resource.VMID))
	}
	add("Name", resource.Name)
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
	lines = append(lines, "", dimStyle.Render("Press any key to close."))
	return lines
}

func helpLines() []string {
	entry := func(key, desc string) string {
		return "  " + menuKeyStyle.Render(pad("<"+key+">", 9)) + menuDescStyle.Render(desc)
	}
	return []string{
		"",
		helpSectionStyle.Render("Navigation"),
		entry("j/k", "move selection (arrows work too)"),
		entry("g/G", "jump to top / bottom"),
		entry("pgup/pgdn", "page up / down"),
		entry("1/2/3/4", "guests / nodes / storage / tasks view"),
		entry("tab", "cycle views"),
		entry("/", "filter rows (esc clears)"),
		entry(":", "command mode (guests, vm, lxc, nodes, storage, tasks, quit)"),
		entry("shift-x", "sort by column (I id, N name, O node, S status, C cpu, M mem, A age, U used, T total)"),
		entry("enter", "describe the selection"),
		entry("R", "refresh now"),
		"",
		helpSectionStyle.Render("Guest actions"),
		entry("s", "start"),
		entry("d", "shutdown (graceful, confirms)"),
		entry("x", "stop (hard, confirms)"),
		entry("r", "reboot (confirms)"),
		entry("c", "console (interactive shell, Ctrl+] to exit)"),
		entry("t", "snapshots"),
		entry("ctrl-d", "delete (stopped guests, confirms)"),
		"",
		entry("q", "quit"),
	}
}

// blockWidth is the widest visible line in a block.
func blockWidth(lines []string) int {
	width := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > width {
			width = w
		}
	}
	return width
}

// blockLine returns line index of a block, or "" past its end.
func blockLine(lines []string, index int) string {
	if index < len(lines) {
		return lines[index]
	}
	return ""
}

// padVisible right-pads a possibly styled string to width terminal cells.
func padVisible(s string, width int) string {
	if gap := width - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}
