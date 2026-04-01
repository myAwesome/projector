package tui

import (
	"fmt"
	"time"
    "os/exec"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
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
	text string
	err  error
}

type keyMap struct {
	up      key.Binding
	down    key.Binding
	toggle  key.Binding
	refresh key.Binding
	open    key.Binding
	quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.toggle,k.open, k.refresh, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.up, k.down, k.toggle, k.open},
		{k.refresh, k.quit},
	}
}

type model struct {
	table     table.Model
	help      help.Model
	keys      keyMap
	items     []item
	status    string
	lastError error
}

func NewModel() model {
	cols := []table.Column{
		{Title: "Name", Width: 16},
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
			refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
			quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
			open:    key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open dir")),
		},
		status: "Loading projects...",
	}
}

func (m model) Init() tea.Cmd {
	return refreshCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 2)
		h := msg.Height - 8
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
		if msg.err != nil {
			m.lastError = msg.err
			m.status = fmt.Sprintf("Action failed: %v", msg.err)
			return m, nil
		}
		m.lastError = nil
		m.status = msg.text
		return m, refreshCmd()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.open):
        	it, ok := m.selected()
        	if !ok {
        		return m, nil
        	}
        	return m, openDirCmd(it)
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
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	title := lipgloss.NewStyle().Bold(true).Render("projector TUI")
	helpView := m.help.View(m.keys)
	status := m.status
	if m.lastError != nil {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(status)
	}
	return lipgloss.JoinVertical(lipgloss.Left, title, "", m.table.View(), "", status, helpView)
}

func (m model) selected() (item, bool) {
	i := m.table.Cursor()
	if i < 0 || i >= len(m.items) {
		return item{}, false
	}
	return m.items[i], true
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

		rows = append(rows, table.Row{
			it.project.Name,
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