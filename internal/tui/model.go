package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/presentation"
	tea "github.com/charmbracelet/bubbletea"
)

const logo = `
  ███████╗██████╗  █████╗ ██████╗ ██╗  ██╗███████╗
  ██╔════╝██╔══██╗██╔══██╗██╔══██╗██║ ██╔╝██╔════╝
  ███████╗██████╔╝███████║██████╔╝█████╔╝ ███████╗
  ╚════██║██╔═══╝ ██╔══██║██╔══██╗██╔═██╗ ╚════██║
  ███████║██║     ██║  ██║██║  ██║██║  ██╗███████║
  ╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝`

type Service interface {
	List(context.Context, model.ListOptions) ([]model.Spark, error)
	Add(context.Context, string, app.AddOptions) (model.Spark, error)
	Edit(context.Context, string, string) (model.Spark, error)
	Important(context.Context, string) (model.Spark, error)
	Done(context.Context, string) (model.Spark, error)
	Remove(context.Context, string) error
}

type inputMode int

const (
	browseMode inputMode = iota
	addMode
	editMode
	childMode
	helpMode
)

type Model struct {
	ctx     context.Context
	service Service
	sparks  []model.Spark
	cursor  int
	mode    inputMode
	input   []rune
	width   int
	loading bool
	status  string
	palette presentation.Palette
}

type loadedMsg struct {
	sparks []model.Spark
	err    error
	status string
}

type Option func(*Model)

func WithColor(enabled bool) Option {
	return func(m *Model) {
		m.palette.Enabled = enabled
	}
}

func New(ctx context.Context, service Service, options ...Option) Model {
	m := Model{ctx: ctx, service: service, loading: true}
	for _, option := range options {
		option(&m)
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return m.load("")
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case loadedMsg:
		m.loading = false
		if msg.err != nil {
			m.status = "Error: " + app.FriendlyError(msg.err).Error()
			return m, nil
		}
		m.sparks = msg.sparks
		m.status = msg.status
		m.clampCursor()
		return m, nil
	case tea.KeyMsg:
		if m.mode == addMode || m.mode == editMode || m.mode == childMode {
			return m.updateInput(msg)
		}
		if m.mode == helpMode {
			if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
				m.mode = browseMode
			}
			return m, nil
		}
		return m.updateBrowse(msg)
	}
	return m, nil
}

func (m Model) updateBrowse(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor+1 < len(m.sparks) {
			m.cursor++
		}
	case "a":
		m.beginInput(addMode, nil)
	case "e":
		if selected := m.selected(); selected != nil {
			m.beginInput(editMode, []rune(selected.Title))
		}
	case "c":
		if m.selected() != nil {
			m.beginInput(childMode, nil)
		}
	case "i":
		if selected := m.selected(); selected != nil {
			m.loading = true
			return m, m.toggleImportant(*selected)
		}
	case "d":
		if selected := m.selected(); selected != nil {
			m.loading = true
			return m, m.markDone(*selected)
		}
	case "x":
		if selected := m.selected(); selected != nil {
			m.loading = true
			return m, m.remove(*selected)
		}
	case "r":
		m.loading = true
		return m, m.load("Refreshed")
	case "?":
		m.mode = helpMode
	}
	return m, nil
}

func (m Model) updateInput(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = browseMode
		m.input = nil
		m.status = "Cancelled"
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	case "ctrl+u":
		m.input = nil
	case "enter":
		title := strings.TrimSpace(string(m.input))
		if title == "" {
			m.status = "A title is required"
			return m, nil
		}
		mode := m.mode
		selected := m.selected()
		m.mode = browseMode
		m.input = nil
		m.loading = true
		switch mode {
		case addMode:
			return m, m.add(title, nil)
		case editMode:
			return m, m.edit(*selected, title)
		case childMode:
			return m, m.add(title, selected)
		}
	default:
		if key.Type == tea.KeyRunes {
			m.input = append(m.input, key.Runes...)
		}
	}
	return m, nil
}

func (m *Model) beginInput(mode inputMode, initial []rune) {
	m.mode = mode
	m.input = append([]rune(nil), initial...)
	m.status = ""
}

func (m *Model) selected() *model.Spark {
	if len(m.sparks) == 0 || m.cursor < 0 || m.cursor >= len(m.sparks) {
		return nil
	}
	return &m.sparks[m.cursor]
}

func (m *Model) clampCursor() {
	if len(m.sparks) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.sparks) {
		m.cursor = len(m.sparks) - 1
	}
}

func (m Model) load(status string) tea.Cmd {
	return func() tea.Msg {
		sparks, err := m.service.List(m.ctx, model.ListOptions{})
		return loadedMsg{sparks: sparks, err: err, status: status}
	}
}

func (m Model) after(action func() error, status string) tea.Cmd {
	return func() tea.Msg {
		if err := action(); err != nil {
			return loadedMsg{err: err}
		}
		sparks, err := m.service.List(m.ctx, model.ListOptions{})
		return loadedMsg{sparks: sparks, err: err, status: status}
	}
}

func (m Model) add(title string, parent *model.Spark) tea.Cmd {
	parentID := ""
	status := "Spark added"
	if parent != nil {
		parentID = strconv.FormatInt(parent.ID, 10)
		status = fmt.Sprintf("Child added under #%d", parent.ID)
	}
	return m.after(func() error {
		_, err := m.service.Add(m.ctx, title, app.AddOptions{Parent: parentID})
		return err
	}, status)
}

func (m Model) edit(spark model.Spark, title string) tea.Cmd {
	return m.after(func() error {
		_, err := m.service.Edit(m.ctx, strconv.FormatInt(spark.ID, 10), title)
		return err
	}, fmt.Sprintf("Spark #%d edited", spark.ID))
}

func (m Model) toggleImportant(spark model.Spark) tea.Cmd {
	return m.after(func() error {
		_, err := m.service.Important(m.ctx, strconv.FormatInt(spark.ID, 10))
		return err
	}, fmt.Sprintf("Spark #%d importance toggled", spark.ID))
}

func (m Model) markDone(spark model.Spark) tea.Cmd {
	return m.after(func() error {
		_, err := m.service.Done(m.ctx, strconv.FormatInt(spark.ID, 10))
		return err
	}, fmt.Sprintf("Spark #%d completed", spark.ID))
}

func (m Model) remove(spark model.Spark) tea.Cmd {
	return m.after(func() error {
		return m.service.Remove(m.ctx, strconv.FormatInt(spark.ID, 10))
	}, fmt.Sprintf("Spark #%d removed", spark.ID))
}

func (m Model) View() string {
	var view strings.Builder
	view.WriteString(m.palette.Paint(presentation.Logo, logo))
	view.WriteString("\n\n  Capture ideas, tasks and nested thoughts without leaving the terminal.\n")
	view.WriteString("  " + m.key("↑/↓") + " or " + m.key("j/k") + " navigate · " +
		m.key("a") + " add · " + m.key("e") + " edit · " + m.key("i") + " important · " +
		m.key("c") + " child · " + m.key("d") + " done · " + m.key("x") + " remove\n\n")

	if m.mode == helpMode {
		view.WriteString(m.palette.Paint(presentation.Important, "  HELP") + "\n")
		view.WriteString("  a add a root spark       e edit the selected spark\n")
		view.WriteString("  c add a child            i toggle important\n")
		view.WriteString("  d mark done              x remove\n")
		view.WriteString("  r refresh                q quit\n")
		view.WriteString("  Press ?, Esc or q to return.\n")
		return view.String()
	}

	view.WriteString(m.palette.Paint(presentation.Muted, "  SEL  ID     STATE       TITLE"))
	if m.width >= 72 {
		view.WriteString(m.palette.Paint(presentation.Muted, "                            PARENT"))
	}
	view.WriteByte('\n')
	view.WriteString(m.palette.Paint(presentation.Muted, "  ───  ─────  ──────────  ─────────────────────────────"))
	if m.width >= 72 {
		view.WriteString(m.palette.Paint(presentation.Muted, "  ──────"))
	}
	view.WriteByte('\n')

	if len(m.sparks) == 0 {
		view.WriteString("       No active sparks. Press a to capture one.\n")
	} else {
		for index, spark := range m.sparks {
			pointer := " "
			if index == m.cursor {
				pointer = ">"
			}
			state := "active"
			if spark.Important {
				state = "important"
			}
			title := truncate(spark.Title, 36)
			row := fmt.Sprintf("   %s   #%-4d  %-10s  %-36s", pointer, spark.ID, state, title)
			if m.width >= 72 {
				parent := "—"
				if spark.ParentID != nil {
					parent = fmt.Sprintf("#%d", *spark.ParentID)
				}
				row += "  " + parent
			}
			if index == m.cursor {
				row = m.palette.Paint(presentation.Selected, row)
			} else {
				row = m.colorRow(row, spark)
			}
			view.WriteString(row)
			view.WriteByte('\n')
		}
	}

	view.WriteByte('\n')
	if m.loading {
		view.WriteString(m.palette.Paint(presentation.Warning, "  Working…") + "\n")
	}
	switch m.mode {
	case addMode:
		fmt.Fprintf(&view, "  New spark: %s█\n", string(m.input))
	case editMode:
		fmt.Fprintf(&view, "  Edit title: %s█\n", string(m.input))
	case childMode:
		selected := m.selected()
		fmt.Fprintf(&view, "  Child of #%d: %s█\n", selected.ID, string(m.input))
	default:
		if m.status != "" {
			role := presentation.Success
			if strings.HasPrefix(m.status, "Error:") {
				role = presentation.Error
			} else if m.status == "Cancelled" || m.status == "A title is required" {
				role = presentation.Warning
			}
			fmt.Fprintf(&view, "  %s\n", m.palette.Paint(role, m.status))
		}
		view.WriteString("  " + m.key("?") + " help · " + m.key("r") + " refresh · " + m.key("q") + " quit\n")
	}
	return view.String()
}

func (m Model) key(value string) string {
	return m.palette.Paint(presentation.Key, value)
}

func (m Model) colorRow(row string, spark model.Spark) string {
	if spark.ParentID != nil {
		parentID := fmt.Sprintf("#%d", *spark.ParentID)
		parentStart := strings.LastIndex(row, parentID)
		if parentStart >= 0 {
			row = row[:parentStart] + m.palette.Paint(presentation.ID, parentID) + row[parentStart+len(parentID):]
		}
	}
	id := fmt.Sprintf("#%d", spark.ID)
	idStart := strings.Index(row, id)
	if idStart >= 0 {
		row = row[:idStart] + m.palette.Paint(presentation.ID, id) + row[idStart+len(id):]
	}
	if spark.Important {
		stateStart := strings.Index(row, "important")
		if stateStart >= 0 {
			row = row[:stateStart] + m.palette.Paint(presentation.Important, "important") + row[stateStart+len("important"):]
		}
	}
	return row
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}
