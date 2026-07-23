package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeSource struct {
	mu        sync.Mutex
	rows      []Resource
	err       error
	actions   []string
	actionErr error
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
	model := NewModel(Options{Source: source, ContextName: "default"})
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
	if len(model.visibleRows()) != 2 {
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
	output := model.View()
	for _, want := range []string{"proxmox-cli", "context: default", "VMID", "web", "db"} {
		if !strings.Contains(output, want) {
			t.Errorf("view missing %q:\n%s", want, output)
		}
	}
	model, _ = press(t, model, "?")
	if !strings.Contains(model.View(), "Keyboard reference") {
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
