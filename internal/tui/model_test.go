package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeSource struct {
	mu        sync.Mutex
	rows      []Resource
	tasks     []Resource
	snaps     []Snapshot
	err       error
	actions   []string
	actionErr error
	shellErr  error
	shellRan  bool
}

func (f *fakeSource) Resources(ctx context.Context) ([]Resource, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows, f.err
}

func (f *fakeSource) Guest(ctx context.Context, resource Resource, action Action) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.actions = append(f.actions, fmt.Sprintf("%s %s", action, resource.ID))
	return f.actionErr
}

func (f *fakeSource) Version(ctx context.Context) (string, error) {
	return "8.4.1", nil
}

func (f *fakeSource) Tasks(ctx context.Context) ([]Resource, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tasks, nil
}

func (f *fakeSource) Snapshots(ctx context.Context, resource Resource) ([]Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.snaps, nil
}

func (f *fakeSource) Shell(resource Resource) (ShellSession, error) {
	if f.shellErr != nil {
		return nil, f.shellErr
	}
	return func(stdin io.Reader, stdout, stderr io.Writer) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.shellRan = true
		return nil
	}, nil
}

func (f *fakeSource) recorded() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string{}, f.actions...)
}

func fixtureRows() []Resource {
	return []Resource{
		{Kind: KindVM, ID: "qemu/100", VMID: 100, Name: "web", Node: "pve1", Status: "running", Mem: 512, MaxMem: 1024, Uptime: 3600},
		{Kind: KindLXC, ID: "lxc/200", VMID: 200, Name: "db", Node: "pve2", Status: "stopped"},
		{Kind: KindNode, ID: "node/pve1", Name: "pve1", Node: "pve1", Status: "online", MaxCPU: 8},
		{Kind: KindStorage, ID: "storage/pve1/local", Name: "local", Node: "pve1", Status: "available", Disk: 100, MaxDisk: 200},
	}
}

func newTestModel(t *testing.T, rows []Resource) (Model, *fakeSource) {
	t.Helper()
	source := &fakeSource{rows: rows}
	model := NewModel(Options{
		Source:      source,
		ContextName: "default",
		Server:      "pve.example.com",
		User:        "root@pam!cli",
		CLIVersion:  "test",
	})
	next, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	next, _ = next.Update(resourcesMsg{rows: rows})
	return next.(Model), source
}

func press(t *testing.T, model Model, keys ...string) (Model, tea.Cmd) {
	t.Helper()
	var cmd tea.Cmd
	for _, key := range keys {
		var msg tea.KeyMsg
		switch key {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "backspace":
			msg = tea.KeyMsg{Type: tea.KeyBackspace}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		var next tea.Model
		next, cmd = model.Update(msg)
		model = next.(Model)
	}
	return model, cmd
}

func TestViewSwitchingFiltersRowsByKind(t *testing.T) {
	model, _ := newTestModel(t, fixtureRows())
	if got := len(model.visibleRows()); got != 2 {
		t.Fatalf("guests view should show 2 rows, got %d", got)
	}
	model, _ = press(t, model, "2")
	rows := model.visibleRows()
	if len(rows) != 1 || rows[0].Kind != KindNode {
		t.Fatalf("nodes view rows = %+v", rows)
	}
	model, _ = press(t, model, "3")
	rows = model.visibleRows()
	if len(rows) != 1 || rows[0].Kind != KindStorage {
		t.Fatalf("storage view rows = %+v", rows)
	}
	model, _ = press(t, model, "tab")
	if model.view != viewTasks {
		t.Fatal("tab from storage should reach the tasks view")
	}
	model, _ = press(t, model, "tab")
	if model.view != viewGuests || len(model.visibleRows()) != 2 {
		t.Fatal("tab should cycle back to the guests view")
	}
}

func TestFilterNarrowsAndEscClears(t *testing.T) {
	model, _ := newTestModel(t, fixtureRows())
	model, _ = press(t, model, "/", "w", "e", "b")
	rows := model.visibleRows()
	if len(rows) != 1 || rows[0].Name != "web" {
		t.Fatalf("filtered rows = %+v", rows)
	}
	model, _ = press(t, model, "enter")
	if model.filtering {
		t.Fatal("enter should leave filter-input mode")
	}
	if len(model.visibleRows()) != 1 {
		t.Fatal("filter should still apply after enter")
	}
	model, _ = press(t, model, "esc")
	if len(model.visibleRows()) != 2 {
		t.Fatal("esc should clear the filter")
	}
}

func TestCursorNavigationClamps(t *testing.T) {
	model, _ := newTestModel(t, fixtureRows())
	model, _ = press(t, model, "j", "j", "j")
	if model.cursor != 1 {
		t.Fatalf("cursor should clamp to last row, got %d", model.cursor)
	}
	model, _ = press(t, model, "g")
	if model.cursor != 0 {
		t.Fatalf("g should jump to top, got %d", model.cursor)
	}
	model, _ = press(t, model, "G")
	if model.cursor != 1 {
		t.Fatalf("G should jump to bottom, got %d", model.cursor)
	}
}

func TestStartActionRunsWithoutConfirmation(t *testing.T) {
	model, source := newTestModel(t, fixtureRows())
	model, cmd := press(t, model, "s")
	if cmd == nil {
		t.Fatal("start should return an action command")
	}
	if _, busy := model.inFlight["qemu/100"]; !busy {
		t.Fatal("start should mark the row in flight")
	}
	msg := cmd()
	done, ok := msg.(actionDoneMsg)
	if !ok || done.err != nil {
		t.Fatalf("unexpected action result: %#v", msg)
	}
	if got := source.recorded(); len(got) != 1 || got[0] != "start qemu/100" {
		t.Fatalf("recorded actions = %v", got)
	}
	next, _ := model.Update(msg)
	model = next.(Model)
	if _, busy := model.inFlight["qemu/100"]; busy {
		t.Fatal("completion should clear the in-flight marker")
	}
	if !strings.Contains(model.status, "started") {
		t.Fatalf("status = %q", model.status)
	}
}

func TestStopRequiresConfirmation(t *testing.T) {
	model, source := newTestModel(t, fixtureRows())
	model, cmd := press(t, model, "x")
	if cmd != nil || model.confirm == nil {
		t.Fatal("stop should open a confirmation instead of acting")
	}
	model, _ = press(t, model, "n")
	if model.confirm != nil || len(source.recorded()) != 0 {
		t.Fatal("declining should cancel without calling the API")
	}
	model, _ = press(t, model, "x")
	_, cmd = press(t, model, "y")
	if cmd == nil {
		t.Fatal("confirming should return an action command")
	}
	cmd()
	if got := source.recorded(); len(got) != 1 || got[0] != "stop qemu/100" {
		t.Fatalf("recorded actions = %v", got)
	}
}

func TestActionsRejectedForNodesAndTemplates(t *testing.T) {
	model, source := newTestModel(t, fixtureRows())
	model, _ = press(t, model, "2")
	model, cmd := press(t, model, "s")
	if cmd != nil || len(source.recorded()) != 0 {
		t.Fatal("actions must not run on node rows")
	}
	if !model.statusIsError {
		t.Fatalf("expected an error status, got %q", model.status)
	}

	rows := []Resource{{Kind: KindVM, ID: "qemu/900", VMID: 900, Name: "tmpl", Node: "pve1", Template: true}}
	model, source = newTestModel(t, rows)
	model, cmd = press(t, model, "s")
	if cmd != nil || len(source.recorded()) != 0 {
		t.Fatal("actions must not run on templates")
	}
	if !strings.Contains(model.status, "template") {
		t.Fatalf("status = %q", model.status)
	}
}

func TestRefreshErrorIsSurfaced(t *testing.T) {
	model, _ := newTestModel(t, fixtureRows())
	next, _ := model.Update(resourcesMsg{err: errors.New("connection refused")})
	model = next.(Model)
	if !model.statusIsError || !strings.Contains(model.status, "refresh failed") {
		t.Fatalf("status = %q", model.status)
	}
	if len(model.rows) != 4 {
		t.Fatal("a failed refresh should keep the previous rows")
	}
}

func TestSelectionSurvivesRefresh(t *testing.T) {
	model, _ := newTestModel(t, fixtureRows())
	model, _ = press(t, model, "j")
	rows := append([]Resource{
		{Kind: KindVM, ID: "qemu/50", VMID: 50, Name: "new", Node: "pve1", Status: "running"},
	}, fixtureRows()...)
	next, _ := model.Update(resourcesMsg{rows: rows})
	model = next.(Model)
	selected, ok := model.selectedResource()
	if !ok || selected.ID != "lxc/200" {
		t.Fatalf("selection should follow the row ID, got %+v", selected)
	}
}

func TestViewRendersTableAndOverlays(t *testing.T) {
	model, _ := newTestModel(t, fixtureRows())
	next, _ := model.Update(versionMsg{version: "8.4.1"})
	model = next.(Model)
	output := model.View()
	for _, want := range []string{"Context:", "default", "pve.example.com", "root@pam!cli", "8.4.1", "VMID", "web", "db", "Guests", "<guests>"} {
		if !strings.Contains(output, want) {
			t.Errorf("view missing %q:\n%s", want, output)
		}
	}
	model, _ = press(t, model, "?")
	if !strings.Contains(model.View(), "Guest actions") {
		t.Error("help overlay should render")
	}
	model, _ = press(t, model, "q")
	model, _ = press(t, model, "enter")
	if !strings.Contains(model.View(), "Memory") {
		t.Errorf("details overlay should render:\n%s", model.View())
	}
	model, _ = press(t, model, "esc", "x")
	if !strings.Contains(model.View(), "Confirm stop") {
		t.Errorf("confirm overlay should render:\n%s", model.View())
	}
}

func TestSortingByColumn(t *testing.T) {
	model, _ := newTestModel(t, fixtureRows())
	names := func() []string {
		rows := model.visibleRows()
		out := make([]string, len(rows))
		for i, r := range rows {
			out[i] = r.Name
		}
		return out
	}
	if got := names(); got[0] != "web" || got[1] != "db" {
		t.Fatalf("default order should be VMID asc, got %v", got)
	}
	model, _ = press(t, model, "N")
	if got := names(); got[0] != "db" || got[1] != "web" {
		t.Fatalf("N should sort by name asc, got %v", got)
	}
	model, _ = press(t, model, "N")
	if got := names(); got[0] != "web" || got[1] != "db" {
		t.Fatalf("N again should invert to desc, got %v", got)
	}
	model, _ = press(t, model, "2")
	if model.sortKey != "" {
		t.Fatal("switching views should reset the sort")
	}
}

func TestTasksView(t *testing.T) {
	model, source := newTestModel(t, fixtureRows())
	source.tasks = []Resource{
		{Kind: KindTask, ID: "UPID:pve1:1", Name: "vzdump", Target: "100", Node: "pve1", User: "root@pam", Status: "OK", Start: 100, End: 160},
		{Kind: KindTask, ID: "UPID:pve1:2", Name: "qmstart", Target: "100", Node: "pve1", User: "root@pam", Status: "running", Start: 200},
	}
	model, cmd := press(t, model, "4")
	if cmd == nil {
		t.Fatal("opening the tasks view should trigger a fetch")
	}
	next, _ := model.Update(cmd())
	model = next.(Model)
	rows := model.visibleRows()
	if len(rows) != 2 || rows[0].ID != "UPID:pve1:2" {
		t.Fatalf("tasks should list most recent first, got %+v", rows)
	}
	if !strings.Contains(model.View(), "vzdump") {
		t.Error("tasks view should render task types")
	}
}

func TestConsoleGuards(t *testing.T) {
	model, source := newTestModel(t, fixtureRows())
	source.shellErr = errors.New("console requires session login")
	model, cmd := press(t, model, "c")
	if cmd != nil || !model.statusIsError || !strings.Contains(model.status, "session login") {
		t.Fatalf("shell error should flash, got %q", model.status)
	}

	source.shellErr = nil
	_, cmd = press(t, model, "c")
	if cmd == nil {
		t.Fatal("console should return an exec command")
	}

	model, _ = press(t, model, "2")
	model, cmd = press(t, model, "c")
	if cmd != nil || !model.statusIsError {
		t.Fatal("console must not open on node rows")
	}
}

func TestDeleteRequiresStoppedGuestAndConfirmation(t *testing.T) {
	model, source := newTestModel(t, fixtureRows())
	ctrlD := tea.KeyMsg{Type: tea.KeyCtrlD}

	next, cmd := model.Update(ctrlD)
	model = next.(Model)
	if cmd != nil || model.confirm != nil || !model.statusIsError {
		t.Fatalf("delete on a running guest should be refused, got %q", model.status)
	}

	model, _ = press(t, model, "j")
	next, _ = model.Update(ctrlD)
	model = next.(Model)
	if model.confirm == nil || model.confirm.action != ActionDelete {
		t.Fatal("delete on a stopped guest should ask for confirmation")
	}
	_, cmd = press(t, model, "y")
	if cmd == nil {
		t.Fatal("confirming delete should return an action command")
	}
	cmd()
	if got := source.recorded(); len(got) != 1 || got[0] != "delete lxc/200" {
		t.Fatalf("recorded actions = %v", got)
	}
}

func TestSnapshotsOverlay(t *testing.T) {
	model, source := newTestModel(t, fixtureRows())
	source.snaps = []Snapshot{{Name: "pre-upgrade", Created: 1700000000, Description: "before v2"}}
	model, cmd := press(t, model, "t")
	if cmd == nil || model.snapshotsFor == nil {
		t.Fatal("t should open the snapshots overlay")
	}
	next, _ := model.Update(cmd())
	model = next.(Model)
	output := model.View()
	if !strings.Contains(output, "pre-upgrade") || !strings.Contains(output, "<snapshots>") {
		t.Errorf("snapshots overlay should render:\n%s", output)
	}
	model, _ = press(t, model, "q")
	if model.snapshotsFor != nil {
		t.Fatal("any key should close the snapshots overlay")
	}
}

func TestCommandModeSwitchesViews(t *testing.T) {
	model, _ := newTestModel(t, fixtureRows())
	model, _ = press(t, model, ":", "n", "o", "d", "e", "s", "enter")
	rows := model.visibleRows()
	if len(rows) != 1 || rows[0].Kind != KindNode {
		t.Fatalf(":nodes should switch to the nodes view, got %+v", rows)
	}

	model, _ = press(t, model, ":", "v", "m", "enter")
	rows = model.visibleRows()
	if model.kindFilter != KindVM || len(rows) != 1 || rows[0].Kind != KindVM {
		t.Fatalf(":vm should scope guests to VMs, got filter %q rows %+v", model.kindFilter, rows)
	}
	model, _ = press(t, model, "esc")
	if model.kindFilter != "" || len(model.visibleRows()) != 2 {
		t.Fatal("esc should clear the kind scope")
	}

	model, cmd := press(t, model, ":", "b", "o", "g", "u", "s", "enter")
	if cmd != nil || !model.statusIsError {
		t.Fatalf("unknown command should flash an error, got %q", model.status)
	}

	_, cmd = press(t, model, ":", "q", "enter")
	if cmd == nil {
		t.Fatal(":q should quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal(":q should produce a quit message")
	}
}
