package tui

import (
	"fmt"
	"strings"

	"github.com/Skeyelab/coauthor-cleaner/internal/clean"
	"github.com/Skeyelab/coauthor-cleaner/internal/detect"
	"github.com/Skeyelab/coauthor-cleaner/internal/git"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenList screen = iota
	screenConfirm
	screenSummary
)

type findingItem struct {
	finding detect.Finding
	index   int
}

func (i findingItem) FilterValue() string { return i.finding.Match }
func (i findingItem) Title() string {
	src := string(i.finding.Source)
	if i.finding.FilePath != "" {
		return fmt.Sprintf("%s  %s:%d", src, i.finding.FilePath, i.finding.LineNumber)
	}
	return fmt.Sprintf("%s  %s", src, i.finding.Match)
}
func (i findingItem) Description() string {
	preview := i.finding.Match
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}
	mark := "[ ]"
	if i.finding.Selected {
		mark = "[x]"
	}
	return mark + " " + preview
}

type Model struct {
	git      git.Runner
	findings []detect.Finding
	list     list.Model
	screen   screen
	summary  string
	err      error
	width    int
	height   int
}

func New(g git.Runner, findings []detect.Finding) Model {
	items := make([]list.Item, len(findings))
	for i, f := range findings {
		items[i] = findingItem{finding: f, index: i}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Coauthor Cleaner"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	return Model{
		git:      g,
		findings: findings,
		list:     l,
		screen:   screenList,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 6)
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case screenList:
			return m.updateList(msg)
		case screenConfirm:
			return m.updateConfirm(msg)
		case screenSummary:
			if msg.String() == "q" || msg.String() == "enter" {
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case " ":
		idx := m.list.Index()
		if idx >= 0 && idx < len(m.findings) {
			m.findings[idx].Selected = !m.findings[idx].Selected
			m.refreshList()
		}
	case "a", "A":
		for i := range m.findings {
			m.findings[i].Selected = true
		}
		m.refreshList()
	case "c":
		if m.selectedCount() == 0 {
			return m, nil
		}
		m.screen = screenConfirm
	case "?":
		// help shown in view
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "b":
		m.screen = screenList
	case "c":
		result, err := clean.Apply(m.git, clean.Options{Findings: m.findings})
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.summary = result.Summary
		m.screen = screenSummary
	}
	return m, nil
}

func (m *Model) refreshList() {
	items := make([]list.Item, len(m.findings))
	for i, f := range m.findings {
		items[i] = findingItem{finding: f, index: i}
	}
	m.list.SetItems(items)
}

func (m Model) selectedCount() int {
	n := 0
	for _, f := range m.findings {
		if f.Selected {
			n++
		}
	}
	return n
}

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render("Error: "+m.err.Error()) + "\n"
	}

	switch m.screen {
	case screenConfirm:
		return m.viewConfirm()
	case screenSummary:
		return m.viewSummary()
	default:
		return m.viewList()
	}
}

func (m Model) viewList() string {
	if len(m.findings) == 0 {
		return titleStyle.Render("Coauthor Cleaner") + "\n\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("No AI attribution markers found.") + "\n"
	}

	help := helpStyle.Render("[space] toggle  [a] accept all  [c] clean  [q] quit  [?] help")
	header := fmt.Sprintf("Found %d possible AI attribution markers\n\n", len(m.findings))
	return titleStyle.Render("Coauthor Cleaner") + "\n\n" + header + m.list.View() + "\n" + help
}

func (m Model) viewConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Apply cleanups?") + "\n\n")
	for _, f := range m.findings {
		if !f.Selected {
			continue
		}
		b.WriteString("  ✓ ")
		b.WriteString(string(f.Source))
		if f.FilePath != "" {
			b.WriteString(": " + f.FilePath)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[c] clean  [b] back  [q] quit"))
	return b.String()
}

func (m Model) viewSummary() string {
	return titleStyle.Render("Done") + "\n\n" + m.summary + "\n" + helpStyle.Render("[q] quit")
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func Run(g git.Runner, findings []detect.Finding) error {
	m := New(g, findings)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
