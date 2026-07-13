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
	sparks      []model.Spark
	listErr     error
	addedTitle  string
	addedParent string
	editedID    string
	editedTitle string
	importantID string
	doneID      string
	removedID   string
	actionErr   error
}

func (f *fakeService) List(context.Context, model.ListOptions) ([]model.Spark, error) {
	return append([]model.Spark(nil), f.sparks...), f.listErr
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

func TestModelShowsEmptyHelpAndErrors(t *testing.T) {
	service := &fakeService{}
	m := loadModel(t, New(context.Background(), service))
	if !strings.Contains(m.View(), "No active sparks") {
		t.Fatalf("missing empty state:\n%s", m.View())
	}
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
