package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"project/internal/runner"
	"project/internal/store"
)

type item struct {
	project store.Project
	running bool
	rs      runner.RunState
}

type refreshMsg struct {
	items []item
	err   error
}

type actionMsg struct {
	text   string
	output string
	err    error
}

type keyMap struct {
	up      key.Binding
	down    key.Binding
	toggle  key.Binding
	pull    key.Binding
	edit    key.Binding
	refresh key.Binding
	open    key.Binding
	quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.toggle, k.pull, k.edit, k.open, k.refresh, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.up, k.down, k.toggle, k.pull, k.edit, k.open},
		{k.refresh, k.quit},
	}
}

type editState struct {
	originalName string
	inputs       []textinput.Model
	focusIndex   int
}

type model struct {
	table      table.Model
	help       help.Model
	keys       keyMap
	items      []item
	status     string
	output     string
	lastError  error
	outputRows int
	editing    *editState
}

func NewModel() model {
	cols := []table.Column{
		{Title: "Name", Width: 16},
		{Title: "Description", Width: 24},
		{Title: "Status", Width: 22},
		{Title: "Ports", Width: 16},
		{Title: "Started", Width: 10},
		{Title: "Script", Width: 40},
		{Title: "Dir", Width: 80},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(14),
	)
	t.SetStyles(table.DefaultStyles())

	return model{
		table: t,
		help:  help.New(),
		keys: keyMap{
			up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
			down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
			toggle:  key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter/space", "run/stop")),
			pull:    key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "git pull")),
			edit:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit name/desc")),
			refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
			quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
			open:    key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open dir")),
		},
		status:     "Loading projects...",
		output:     "No command output yet.",
		outputRows: 6,
	}
}

func (m model) Init() tea.Cmd {
	return refreshCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 2)
		h := msg.Height - (m.outputRows + 9)
		if h < 5 {
			h = 5
		}
		m.table.SetHeight(h)
		return m, nil
	case refreshMsg:
		if msg.err != nil {
			m.lastError = msg.err
			m.status = fmt.Sprintf("Refresh failed: %v", msg.err)
			return m, nil
		}
		m.lastError = nil
		m.items = msg.items
		m.table.SetRows(rowsFor(msg.items))
		if len(msg.items) == 0 {
			m.status = "No projects registered. Use: proj register ..."
		} else {
			m.status = fmt.Sprintf("%d project(s). Press enter to run/stop selected.", len(msg.items))
		}
		return m, nil
	case actionMsg:
		if msg.output != "" {
			m.output = msg.output
		}
		if msg.err != nil {
			m.lastError = msg.err
			if msg.text != "" {
				m.status = msg.text
			} else {
				m.status = fmt.Sprintf("Action failed: %v", msg.err)
			}
			return m, nil
		}
		m.lastError = nil
		m.status = msg.text
		return m, refreshCmd()
	case tea.KeyMsg:
		if m.editing != nil {
			return m.updateEditing(msg)
		}

		switch {
		case key.Matches(msg, m.keys.open):
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			return m, openDirCmd(it)
		case key.Matches(msg, m.keys.pull):
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			m.status = fmt.Sprintf("Running git pull for %q...", it.project.Name)
			return m, gitPullCmd(it)
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.refresh):
			m.status = "Refreshing..."
			return m, refreshCmd()
		case key.Matches(msg, m.keys.toggle):
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			return m, toggleCmd(it)
		case key.Matches(msg, m.keys.edit):
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			if it.running {
				m.lastError = fmt.Errorf("project is running")
				m.status = "Stop the project before renaming it."
				return m, nil
			}
			m.editing = newEditState(it.project)
			m.status = fmt.Sprintf("Editing %q. Tab to switch, enter to save, esc to cancel.", it.project.Name)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	title := lipgloss.NewStyle().Bold(true).Render("projector TUI")
	helpView := m.help.View(m.keys)
	tableView := m.table.View()
	if m.editing != nil {
		tableView = m.editingView()
		helpView = "tab/shift+tab: field  •  enter: save  •  esc: cancel"
	}
	status := m.status
	if m.lastError != nil {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(status)
	}

	outputTitle := lipgloss.NewStyle().Bold(true).Render("Output")
	outputBody := lastNLines(m.output, m.outputRows)
	outputView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Height(m.outputRows).
		Render(outputBody)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		tableView,
		"",
		status,
		"",
		outputTitle,
		outputView,
		helpView,
	)
}

func (m model) selected() (item, bool) {
	i := m.table.Cursor()
	if i < 0 || i >= len(m.items) {
		return item{}, false
	}
	return m.items[i], true
}

func newEditState(p store.Project) *editState {
	nameInput := textinput.New()
	nameInput.Placeholder = "name"
	nameInput.SetValue(p.Name)
	nameInput.CharLimit = 100
	nameInput.Focus()

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "description"
	descriptionInput.SetValue(p.Description)
	descriptionInput.CharLimit = 200

	return &editState{
		originalName: p.Name,
		inputs:       []textinput.Model{nameInput, descriptionInput},
		focusIndex:   0,
	}
}

func (m model) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editing = nil
		m.lastError = nil
		m.status = "Edit canceled."
		return m, nil
	case tea.KeyEnter:
		nextName := strings.TrimSpace(m.editing.inputs[0].Value())
		nextDescription := strings.TrimSpace(m.editing.inputs[1].Value())
		if nextName == "" {
			m.lastError = fmt.Errorf("name is required")
			m.status = "Name cannot be empty."
			return m, nil
		}
		cmd := saveMetadataCmd(m.editing.originalName, nextName, nextDescription)
		m.editing = nil
		m.status = "Saving changes..."
		return m, cmd
	case tea.KeyTab, tea.KeyShiftTab, tea.KeyUp, tea.KeyDown:
		m.editing.blurAll()
		if msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyUp {
			m.editing.focusIndex--
		} else {
			m.editing.focusIndex++
		}
		if m.editing.focusIndex >= len(m.editing.inputs) {
			m.editing.focusIndex = 0
		}
		if m.editing.focusIndex < 0 {
			m.editing.focusIndex = len(m.editing.inputs) - 1
		}
		m.editing.inputs[m.editing.focusIndex].Focus()
		return m, nil
	}

	var cmd tea.Cmd
	m.editing.inputs[m.editing.focusIndex], cmd = m.editing.inputs[m.editing.focusIndex].Update(msg)
	return m, cmd
}

func (e *editState) blurAll() {
	for i := range e.inputs {
		e.inputs[i].Blur()
	}
}

func (m model) editingView() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(fmt.Sprintf(
			"Edit Project\n\nName\n%s\n\nDescription\n%s",
			m.editing.inputs[0].View(),
			m.editing.inputs[1].View(),
		))
}

func refreshCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := store.Load()
		if err != nil {
			return refreshMsg{err: err}
		}

		items := make([]item, 0, len(projects))
		for _, p := range projects {
			rs, running := runner.IsRunning(p.Name)
			items = append(items, item{
				project: p,
				running: running,
				rs:      rs,
			})
		}
		return refreshMsg{items: items}
	}
}

func toggleCmd(it item) tea.Cmd {
	return func() tea.Msg {
		if it.running {
			if err := runner.Stop(it.project.Name); err != nil {
				return actionMsg{err: err}
			}
			return actionMsg{text: fmt.Sprintf("Stopped %q.", it.project.Name)}
		}

		rs, err := runner.Start(it.project)
		if err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{text: fmt.Sprintf("Started %q (pid %d).", it.project.Name, rs.PID)}
	}
}

func saveMetadataCmd(currentName, nextName, nextDescription string) tea.Cmd {
	return func() tea.Msg {
		if err := store.UpdateMetadata(currentName, nextName, nextDescription); err != nil {
			if err == store.ErrExists {
				return actionMsg{
					text: fmt.Sprintf("Project %q already exists.", nextName),
					err:  err,
				}
			}
			return actionMsg{
				text: "Failed to update project metadata.",
				err:  err,
			}
		}
		return actionMsg{text: fmt.Sprintf("Updated project %q.", nextName)}
	}
}

func gitPullCmd(it item) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "-C", it.project.Dir, "pull")
		out, err := cmd.CombinedOutput()

		output := strings.TrimSpace(string(out))
		if output == "" {
			output = "(no output)"
		}

		if err != nil {
			return actionMsg{
				text:   fmt.Sprintf("git pull failed for %q.", it.project.Name),
				output: output,
				err:    err,
			}
		}

		return actionMsg{
			text:   fmt.Sprintf("git pull finished for %q.", it.project.Name),
			output: output,
		}
	}
}

func rowsFor(items []item) []table.Row {
	rows := make([]table.Row, 0, len(items))
	for _, it := range items {
		status := "stopped"
		ports := "-"
		started := "-"

		if it.running {
			status = fmt.Sprintf("running (pid %d)", it.rs.PID)
			ports = runner.FormatPorts(runner.Ports(it.rs.PGID, it.rs.PID))
			started = it.rs.StartedAt.Local().Format(time.TimeOnly)
		}
		description := it.project.Description
		if description == "" {
			description = "-"
		}

		rows = append(rows, table.Row{
			it.project.Name,
			description,
			status,
			ports,
			started,
			it.project.Script,
			it.project.Dir,
		})
	}
	return rows
}

func openDirCmd(it item) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`
        tell application "Terminal"
            activate
            do script "cd %s"
        end tell
        `, it.project.Dir))

		if err := cmd.Start(); err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{text: fmt.Sprintf("Opened %q in terminal.", it.project.Dir)}
	}
}

func lastNLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}
