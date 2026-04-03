package tui

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
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
	up       key.Binding
	down     key.Binding
	toggle   key.Binding
	pull     key.Binding
	edit     key.Binding
	remove   key.Binding
	register key.Binding
	refresh  key.Binding
	open     key.Binding
	quit     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.toggle, k.register, k.pull, k.edit, k.remove, k.open, k.refresh, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.up, k.down, k.toggle, k.register, k.pull, k.edit, k.remove, k.open},
		{k.refresh, k.quit},
	}
}

type editState struct {
	originalName string
	inputs       []textinput.Model
	focusIndex   int
}

type registerState struct {
	inputs     []textinput.Model
	focusIndex int
}

type model struct {
	table         table.Model
	help          help.Model
	keys          keyMap
	items         []item
	status        string
	output        string
	lastError     error
	outputRows    int
	viewWidth     int
	editing       *editState
	registering   *registerState
	pendingDelete string
}

func NewModel() model {
	cols := []table.Column{
		{Title: "Name", Width: 16},
		{Title: "Description", Width: 40},
		{Title: "Status", Width: 16},
		{Title: "Ports", Width: 16},
		{Title: "Started", Width: 10},
		{Title: "Script", Width: 16},
		{Title: "Dir", Width: 40},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(14),
	)
	t.SetStyles(table.DefaultStyles())

	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	return model{
		table: t,
		help:  h,
		keys: keyMap{
			up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
			down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
			toggle:   key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter/space", "run/stop")),
			register: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "register project")),
			pull:     key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "git pull")),
			edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit project")),
			remove:   key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "remove project")),
			refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
			quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
			open:     key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open dir")),
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
		m.viewWidth = msg.Width
		tablePanelStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
		tableWidth := msg.Width - tablePanelStyle.GetHorizontalFrameSize()
		if tableWidth < 20 {
			tableWidth = 20
		}
		m.table.SetWidth(tableWidth)

		usableRows := msg.Height - 12
		if usableRows < 8 {
			usableRows = 8
		}

		topRows := usableRows / 2
		bottomRows := usableRows - topRows
		if topRows < 4 {
			topRows = 4
			bottomRows = usableRows - topRows
		}
		if bottomRows < 4 {
			bottomRows = 4
		}

		m.table.SetHeight(topRows)
		m.outputRows = bottomRows
		return m, nil
	case refreshMsg:
		if msg.err != nil {
			m.lastError = msg.err
			m.status = fmt.Sprintf("Refresh failed: %v", msg.err)
			return m, nil
		}
		m.pendingDelete = ""
		m.lastError = nil
		m.items = msg.items
		m.table.SetRows(rowsFor(msg.items))
		if len(msg.items) == 0 {
			m.status = "No projects registered. Press n to add one."
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
		if m.registering != nil {
			return m.updateRegistering(msg)
		}

		switch {
		case key.Matches(msg, m.keys.open):
			m.pendingDelete = ""
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			return m, openDirCmd(it)
		case key.Matches(msg, m.keys.pull):
			m.pendingDelete = ""
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			m.status = fmt.Sprintf("Running git pull for %q...", it.project.Name)
			return m, gitPullCmd(it)
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.refresh):
			m.pendingDelete = ""
			m.status = "Refreshing..."
			return m, refreshCmd()
		case key.Matches(msg, m.keys.toggle):
			m.pendingDelete = ""
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			return m, toggleCmd(it)
		case key.Matches(msg, m.keys.edit):
			m.pendingDelete = ""
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			if it.running {
				m.lastError = fmt.Errorf("project is running")
				m.status = "Stop the project before editing it."
				return m, nil
			}
			m.editing = newEditState(it.project)
			m.status = fmt.Sprintf("Editing %q. Tab to switch, enter to save, esc to cancel.", it.project.Name)
			return m, nil
		case key.Matches(msg, m.keys.remove):
			it, ok := m.selected()
			if !ok {
				return m, nil
			}
			if it.running {
				m.pendingDelete = ""
				m.lastError = fmt.Errorf("project is running")
				m.status = "Stop the project before removing it."
				return m, nil
			}
			if m.pendingDelete == it.project.Name {
				m.pendingDelete = ""
				m.status = fmt.Sprintf("Removing %q...", it.project.Name)
				return m, removeProjectCmd(it.project.Name)
			}
			m.pendingDelete = it.project.Name
			m.lastError = nil
			m.status = fmt.Sprintf("Press x again to remove %q.", it.project.Name)
			return m, nil
		case key.Matches(msg, m.keys.register):
			m.pendingDelete = ""
			m.registering = newRegisterState()
			m.status = "Registering project. Tab to switch, enter to save, esc to cancel."
			return m, nil
		}
		m.pendingDelete = ""
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
		helpView = formHelpView()
	} else if m.registering != nil {
		tableView = m.registeringView()
		helpView = formHelpView()
	}
	status := m.status
	if m.lastError != nil {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(status)
	}

	tablePanelStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	if m.viewWidth > 0 {
		tablePanelWidth := m.viewWidth - tablePanelStyle.GetHorizontalFrameSize()
		if tablePanelWidth > 0 {
			tablePanelStyle = tablePanelStyle.Width(tablePanelWidth)
		}
	}
	tablePanel := tablePanelStyle.Render(tableView)

	outputTitle := lipgloss.NewStyle().Bold(true).Render("Output")
	outputBody := lastNLines(m.output, m.outputRows)
	outputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if m.viewWidth > 0 {
		outputWidth := m.viewWidth - outputStyle.GetHorizontalFrameSize()
		if outputWidth > 0 {
			outputStyle = outputStyle.Width(outputWidth)
		}
	}
	outputView := outputStyle.
		Height(m.outputRows).
		Render(outputBody)

	dividerWidth := m.viewWidth
	if dividerWidth <= 0 {
		dividerWidth = 80
	}
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", dividerWidth))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		tablePanel,
		"",
		status,
		"",
		outputTitle,
		outputView,
		divider,
		helpView,
	)
}

func formHelpView() string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		keyStyle.Render("tab/shift+tab")+": "+descStyle.Render("field"),
		"  •  ",
		keyStyle.Render("enter")+": "+descStyle.Render("save"),
		"  •  ",
		keyStyle.Render("esc")+": "+descStyle.Render("cancel"),
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

	dirInput := textinput.New()
	dirInput.Placeholder = "project directory"
	dirInput.SetValue(p.Dir)
	dirInput.CharLimit = 300

	scriptInput := textinput.New()
	scriptInput.Placeholder = "launch script or command"
	scriptInput.SetValue(p.Script)
	scriptInput.CharLimit = 300

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "description"
	descriptionInput.SetValue(p.Description)
	descriptionInput.CharLimit = 200

	return &editState{
		originalName: p.Name,
		inputs:       []textinput.Model{nameInput, dirInput, scriptInput, descriptionInput},
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
		nextDir := strings.TrimSpace(m.editing.inputs[1].Value())
		nextScript := strings.TrimSpace(m.editing.inputs[2].Value())
		nextDescription := strings.TrimSpace(m.editing.inputs[3].Value())
		if nextName == "" || nextDir == "" || nextScript == "" {
			m.lastError = fmt.Errorf("missing required fields")
			m.status = "Name, dir, and script are required."
			return m, nil
		}
		cmd := saveProjectCmd(m.editing.originalName, nextName, nextDescription, nextDir, nextScript)
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

func newRegisterState() *registerState {
	nameInput := textinput.New()
	nameInput.Placeholder = "name"
	nameInput.CharLimit = 100
	nameInput.Focus()

	dirInput := textinput.New()
	dirInput.Placeholder = "project directory"
	dirInput.CharLimit = 300

	scriptInput := textinput.New()
	scriptInput.Placeholder = "launch script or command"
	scriptInput.CharLimit = 300

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "description (optional)"
	descriptionInput.CharLimit = 200

	return &registerState{
		inputs:     []textinput.Model{nameInput, dirInput, scriptInput, descriptionInput},
		focusIndex: 0,
	}
}

func (r *registerState) blurAll() {
	for i := range r.inputs {
		r.inputs[i].Blur()
	}
}

func (m model) editingView() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(fmt.Sprintf(
			"Edit Project\n\nName\n%s\n\nDir\n%s\n\nScript\n%s\n\nDescription\n%s",
			m.editing.inputs[0].View(),
			m.editing.inputs[1].View(),
			m.editing.inputs[2].View(),
			m.editing.inputs[3].View(),
		))
}

func (m model) updateRegistering(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.registering = nil
		m.lastError = nil
		m.status = "Register canceled."
		return m, nil
	case tea.KeyEnter:
		name := strings.TrimSpace(m.registering.inputs[0].Value())
		dir := strings.TrimSpace(m.registering.inputs[1].Value())
		script := strings.TrimSpace(m.registering.inputs[2].Value())
		description := strings.TrimSpace(m.registering.inputs[3].Value())
		if name == "" || dir == "" || script == "" {
			m.lastError = fmt.Errorf("missing required fields")
			m.status = "Name, dir, and script are required."
			return m, nil
		}
		cmd := registerProjectCmd(name, dir, script, description)
		m.registering = nil
		m.status = "Registering project..."
		return m, cmd
	case tea.KeyTab, tea.KeyShiftTab, tea.KeyUp, tea.KeyDown:
		m.registering.blurAll()
		if msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyUp {
			m.registering.focusIndex--
		} else {
			m.registering.focusIndex++
		}
		if m.registering.focusIndex >= len(m.registering.inputs) {
			m.registering.focusIndex = 0
		}
		if m.registering.focusIndex < 0 {
			m.registering.focusIndex = len(m.registering.inputs) - 1
		}
		m.registering.inputs[m.registering.focusIndex].Focus()
		return m, nil
	}

	var cmd tea.Cmd
	m.registering.inputs[m.registering.focusIndex], cmd = m.registering.inputs[m.registering.focusIndex].Update(msg)
	return m, cmd
}

func (m model) registeringView() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(fmt.Sprintf(
			"Register Project\n\nName\n%s\n\nDir\n%s\n\nScript\n%s\n\nDescription\n%s",
			m.registering.inputs[0].View(),
			m.registering.inputs[1].View(),
			m.registering.inputs[2].View(),
			m.registering.inputs[3].View(),
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

func registerProjectCmd(name, dir, script, description string) tea.Cmd {
	return func() tea.Msg {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return actionMsg{
				text: "Failed to resolve project directory.",
				err:  err,
			}
		}

		err = store.Add(store.Project{
			Name:        name,
			Description: description,
			Dir:         absDir,
			Script:      script,
		})
		if errors.Is(err, store.ErrExists) {
			return actionMsg{
				text: fmt.Sprintf("Project %q already exists.", name),
				err:  err,
			}
		}
		if err != nil {
			return actionMsg{
				text: "Failed to register project.",
				err:  err,
			}
		}
		return actionMsg{text: fmt.Sprintf("Registered %q (%s).", name, absDir)}
	}
}

func saveProjectCmd(currentName, nextName, nextDescription, nextDir, nextScript string) tea.Cmd {
	return func() tea.Msg {
		absDir, err := filepath.Abs(nextDir)
		if err != nil {
			return actionMsg{
				text: "Failed to resolve project directory.",
				err:  err,
			}
		}

		if err := store.UpdateProject(currentName, nextName, nextDescription, absDir, nextScript); err != nil {
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

func removeProjectCmd(name string) tea.Cmd {
	return func() tea.Msg {
		if err := store.Remove(name); err != nil {
			if err == store.ErrNotFound {
				return actionMsg{
					text: fmt.Sprintf("Project %q was not found.", name),
					err:  err,
				}
			}
			return actionMsg{
				text: fmt.Sprintf("Failed to remove %q.", name),
				err:  err,
			}
		}

		return actionMsg{text: fmt.Sprintf("Removed %q.", name)}
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
