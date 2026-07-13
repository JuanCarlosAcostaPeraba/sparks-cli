package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeService struct {
	sparks        []model.Spark
	searchResults []model.Spark
	listErr       error
	searchErr     error
	listOpts      model.ListOptions
	searchOpts    model.ListOptions
	searchQuery   string
	addedTitle    string
	addedParent   string
	editedID      string
	editedTitle   string
	importantID   string
	doneID        string
	removedID     string
	clearCount    int64
	clearCalls    int
	actionErr     error
}

func (f *fakeService) List(_ context.Context, opts model.ListOptions) ([]model.Spark, error) {
	f.listOpts = opts
	return append([]model.Spark(nil), f.sparks...), f.listErr
}

func (f *fakeService) Search(_ context.Context, query string, opts model.ListOptions) ([]model.Spark, error) {
	f.searchQuery, f.searchOpts = query, opts
	return append([]model.Spark(nil), f.searchResults...), f.searchErr
}

func (f *fakeService) Add(_ context.Context, title string, opts app.AddOptions) (model.Spark, error) {
	f.addedTitle, f.addedParent = title, opts.Parent
	return model.Spark{}, f.actionErr
}

func (f *fakeService) Edit(_ context.Context, id, title string) (model.Spark, error) {
	f.editedID, f.editedTitle = id, title
	return model.Spark{}, f.actionErr
}

func (f *fakeService) Important(_ context.Context, id string) (model.Spark, error) {
	f.importantID = id
	return model.Spark{}, f.actionErr
}

func (f *fakeService) Done(_ context.Context, id string) (model.Spark, error) {
	f.doneID = id
	return model.Spark{}, f.actionErr
}

func (f *fakeService) Remove(_ context.Context, id string) error {
	f.removedID = id
	return f.actionErr
}

func (f *fakeService) Clear(_ context.Context, all bool) (int64, error) {
	if all {
		return 0, errors.New("TUI must not clear all sparks")
	}
	f.clearCalls++
	return f.clearCount, f.actionErr
}

func TestModelLoadsAndNavigatesSparks(t *testing.T) {
	service := &fakeService{sparks: []model.Spark{
		{ID: 1, Title: "First"},
		{ID: 2, Title: "Second", Important: true},
	}}
	m := loadModel(t, New(context.Background(), service))

	view := m.View()
	for _, want := range []string{"███████╗", "Capture ideas", ">   #1", "important", "? help"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in view:\n%s", want, view)
		}
	}

	m, _ = update(t, m, key("down"))
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
	m, _ = update(t, m, key("j"))
	if m.cursor != 1 {
		t.Fatalf("cursor should remain clamped, got %d", m.cursor)
	}
	m, _ = update(t, m, key("up"))
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
}

func TestModelColorHighlightsSelectionIDsImportantAndFeedback(t *testing.T) {
	parentID := int64(1)
	service := &fakeService{sparks: []model.Spark{
		{ID: 1, Title: "Selected"},
		{ID: 2, Title: "Important child", Important: true, ParentID: &parentID},
	}}
	m := loadModel(t, New(context.Background(), service, WithColor(true)))
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	view := m.View()
	for _, want := range []string{
		"\x1b[1;35m",
		"\x1b[1;30;46m",
		"\x1b[36m#2\x1b[0m",
		"\x1b[36m#1\x1b[0m",
		"\x1b[1;33m[!] important\x1b[0m",
		"\x1b[35m?\x1b[0m",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in colored TUI:\n%s", want, view)
		}
	}

	m.status = "Error: unavailable"
	if !strings.Contains(m.View(), "\x1b[31mError: unavailable\x1b[0m") {
		t.Fatalf("error feedback was not colored:\n%s", m.View())
	}
}

func TestModelWithoutColorContainsNoANSI(t *testing.T) {
	service := &fakeService{sparks: []model.Spark{
		{ID: 1, Title: "Plain"},
		{ID: 2, Title: "Important", Important: true},
		{ID: 3, Title: "Completed", Done: true},
	}}
	m := loadModel(t, New(context.Background(), service, WithColor(false)))
	if strings.Contains(m.View(), "\x1b[") {
		t.Fatalf("plain TUI contains ANSI escapes:\n%s", m.View())
	}
	for _, want := range []string{"[ ] active", "[!] important", "[x] done"} {
		if !strings.Contains(m.View(), want) {
			t.Fatalf("expected %q in TUI:\n%s", want, m.View())
		}
	}
}

func TestModelAddsRootAndChild(t *testing.T) {
	service := &fakeService{sparks: []model.Spark{{ID: 7, Title: "Parent"}}}
	m := loadModel(t, New(context.Background(), service))

	m, _ = update(t, m, key("a"))
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Root idea")})
	m, cmd := update(t, m, key("enter"))
	m = applyCommand(t, m, cmd)
	if service.addedTitle != "Root idea" || service.addedParent != "" {
		t.Fatalf("root add = (%q, %q)", service.addedTitle, service.addedParent)
	}

	m, _ = update(t, m, key("c"))
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Child idea")})
	m, cmd = update(t, m, key("enter"))
	applyCommand(t, m, cmd)
	if service.addedTitle != "Child idea" || service.addedParent != "7" {
		t.Fatalf("child add = (%q, %q), want parent 7", service.addedTitle, service.addedParent)
	}
}

func TestModelEditsSelectedSpark(t *testing.T) {
	service := &fakeService{sparks: []model.Spark{{ID: 9, Title: "Old title"}}}
	m := loadModel(t, New(context.Background(), service))

	m, _ = update(t, m, key("e"))
	if string(m.input) != "Old title" {
		t.Fatalf("edit input = %q", string(m.input))
	}
	m, _ = update(t, m, key("ctrl+u"))
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("New title")})
	m, cmd := update(t, m, key("enter"))
	applyCommand(t, m, cmd)
	if service.editedID != "9" || service.editedTitle != "New title" {
		t.Fatalf("edit = (%q, %q)", service.editedID, service.editedTitle)
	}
}

func TestModelRunsSelectedActions(t *testing.T) {
	tests := []struct {
		key  string
		read func(*fakeService) string
	}{
		{"i", func(f *fakeService) string { return f.importantID }},
		{"d", func(f *fakeService) string { return f.doneID }},
		{"x", func(f *fakeService) string { return f.removedID }},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			service := &fakeService{sparks: []model.Spark{{ID: 12, Title: "Selected"}}}
			m := loadModel(t, New(context.Background(), service))
			m, cmd := update(t, m, key(tt.key))
			applyCommand(t, m, cmd)
			if got := tt.read(service); got != "12" {
				t.Fatalf("action id = %q, want 12", got)
			}
		})
	}
}

func TestModelSearchesAndClearsSearch(t *testing.T) {
	service := &fakeService{
		sparks:        []model.Spark{{ID: 1, Title: "Normal"}},
		searchResults: []model.Spark{{ID: 2, Title: "Release notes"}},
	}
	m := loadModel(t, New(context.Background(), service))
	m, _ = update(t, m, key("s"))
	if m.mode != searchMode || !strings.Contains(m.View(), "Search:") {
		t.Fatalf("search prompt did not open:\n%s", m.View())
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("release")})
	m.cursor = 9
	m, cmd := update(t, m, key("enter"))
	m = applyCommand(t, m, cmd)
	if service.searchQuery != "release" || service.searchOpts.IncludeDone {
		t.Fatalf("search = %q, opts = %#v", service.searchQuery, service.searchOpts)
	}
	if m.query != "release" || len(m.sparks) != 1 || m.sparks[0].ID != 2 || m.cursor != 0 {
		t.Fatalf("unexpected search model: query=%q cursor=%d sparks=%#v", m.query, m.cursor, m.sparks)
	}
	if !strings.Contains(m.View(), `Search:`) || !strings.Contains(m.View(), `"release"`) {
		t.Fatalf("active query is not visible:\n%s", m.View())
	}

	m, cmd = update(t, m, key("esc"))
	m = applyCommand(t, m, cmd)
	if m.query != "" || len(m.sparks) != 1 || m.sparks[0].ID != 1 {
		t.Fatalf("search was not cleared: query=%q sparks=%#v", m.query, m.sparks)
	}
}

func TestModelSearchCanBeCancelledAndBlankSearchClearsQuery(t *testing.T) {
	service := &fakeService{sparks: []model.Spark{{ID: 1, Title: "Normal"}}}
	m := loadModel(t, New(context.Background(), service))
	m.query = "existing"
	m, _ = update(t, m, key("s"))
	m, _ = update(t, m, key("esc"))
	if m.query != "existing" || m.mode != browseMode {
		t.Fatalf("cancel changed query: %q", m.query)
	}

	m, _ = update(t, m, key("s"))
	m, _ = update(t, m, key("ctrl+u"))
	m, cmd := update(t, m, key("enter"))
	m = applyCommand(t, m, cmd)
	if m.query != "" || service.searchQuery != "" {
		t.Fatalf("blank search did not clear query")
	}
}

func TestModelTogglesCompletedVisibilityAndSearchScope(t *testing.T) {
	service := &fakeService{
		sparks:        []model.Spark{{ID: 1, Title: "Active"}},
		searchResults: []model.Spark{{ID: 2, Title: "Completed match", Done: true}},
	}
	m := loadModel(t, New(context.Background(), service))
	m, cmd := update(t, m, key("v"))
	m = applyCommand(t, m, cmd)
	if !m.showAll || !service.listOpts.IncludeDone || !strings.Contains(m.View(), "View: all sparks") {
		t.Fatalf("all view not enabled: opts=%#v\n%s", service.listOpts, m.View())
	}

	m.query = "match"
	m, cmd = update(t, m, key("r"))
	applyCommand(t, m, cmd)
	if service.searchQuery != "match" || !service.searchOpts.IncludeDone {
		t.Fatalf("search did not inherit all view: query=%q opts=%#v", service.searchQuery, service.searchOpts)
	}

	m, cmd = update(t, m, key("v"))
	m = applyCommand(t, m, cmd)
	if m.showAll || service.searchOpts.IncludeDone || !strings.Contains(m.View(), "View: active only") {
		t.Fatalf("active-only view not restored: opts=%#v\n%s", service.searchOpts, m.View())
	}
}

func TestModelConfirmsBeforeClearingCompleted(t *testing.T) {
	service := &fakeService{sparks: []model.Spark{{ID: 1, Title: "Active"}}, clearCount: 3}
	m := loadModel(t, New(context.Background(), service))
	m, _ = update(t, m, key("C"))
	if m.mode != confirmClearMode || !strings.Contains(m.View(), "Clear all completed sparks? y/n") {
		t.Fatalf("clear confirmation missing:\n%s", m.View())
	}
	m, _ = update(t, m, key("n"))
	if service.clearCalls != 0 || m.status != "Clear cancelled" {
		t.Fatalf("clear cancellation failed: calls=%d status=%q", service.clearCalls, m.status)
	}

	m, _ = update(t, m, key("C"))
	m, cmd := update(t, m, key("y"))
	m = applyCommand(t, m, cmd)
	if service.clearCalls != 1 || m.status != "Cleared 3 completed spark(s)" {
		t.Fatalf("clear result: calls=%d status=%q", service.clearCalls, m.status)
	}
}

func TestModelShowsSearchAndClearErrors(t *testing.T) {
	service := &fakeService{searchErr: errors.New("search unavailable")}
	m := loadModel(t, New(context.Background(), service))
	m, _ = update(t, m, key("s"))
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("query")})
	m, cmd := update(t, m, key("enter"))
	m = applyCommand(t, m, cmd)
	if !strings.Contains(m.status, "search unavailable") {
		t.Fatalf("missing search error: %q", m.status)
	}

	service.searchErr = nil
	service.actionErr = errors.New("clear unavailable")
	m.query = ""
	m, _ = update(t, m, key("C"))
	m, cmd = update(t, m, key("y"))
	m = applyCommand(t, m, cmd)
	if !strings.Contains(m.status, "clear unavailable") {
		t.Fatalf("missing clear error: %q", m.status)
	}
}

func TestModelShowsEmptyHelpAndErrors(t *testing.T) {
	service := &fakeService{}
	m := loadModel(t, New(context.Background(), service))
	if !strings.Contains(m.View(), "No active sparks") {
		t.Fatalf("missing empty state:\n%s", m.View())
	}
	m.query = "missing"
	if !strings.Contains(m.View(), `No sparks match "missing"`) {
		t.Fatalf("missing search empty state:\n%s", m.View())
	}
	m.query = ""
	m.showAll = true
	if !strings.Contains(m.View(), "No sparks yet") {
		t.Fatalf("missing all-sparks empty state:\n%s", m.View())
	}
	m.showAll = false
	m, _ = update(t, m, key("?"))
	if !strings.Contains(m.View(), "HELP") {
		t.Fatalf("missing help view:\n%s", m.View())
	}
	m, _ = update(t, m, key("esc"))
	service.listErr = errors.New("database unavailable")
	m, cmd := update(t, m, key("r"))
	m = applyCommand(t, m, cmd)
	if !strings.Contains(m.View(), "Error: database unavailable") {
		t.Fatalf("missing error feedback:\n%s", m.View())
	}
}

func TestModelRejectsBlankInputAndCancels(t *testing.T) {
	service := &fakeService{}
	m := loadModel(t, New(context.Background(), service))
	m, _ = update(t, m, key("a"))
	m, cmd := update(t, m, key("enter"))
	if cmd != nil || m.status != "A title is required" {
		t.Fatalf("blank input status = %q", m.status)
	}
	m, _ = update(t, m, key("esc"))
	if m.mode != browseMode || m.status != "Cancelled" {
		t.Fatalf("cancel did not return to browse mode")
	}
}

func loadModel(t *testing.T, m Model) Model {
	t.Helper()
	return applyCommand(t, m, m.Init())
}

func applyCommand(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected command")
	}
	updated, _ := update(t, m, cmd())
	return updated
}

func update(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(msg)
	result, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model has type %T", updated)
	}
	return result, cmd
}

func key(value string) tea.KeyMsg {
	switch value {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
	}
}
