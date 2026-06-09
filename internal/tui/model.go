package tui

import (
	"fmt"
	"strings"

	"github.com/ericdahl-dev/coauthor-cleaner/internal/clean"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/detect"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/git"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/guide"
	"github.com/ericdahl-dev/coauthor-cleaner/internal/scan"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenMain screen = iota
	screenConfirm
	screenSummary
	screenPush
	screenHelp
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
	return fmt.Sprintf("%s", src)
}
func (i findingItem) Description() string {
	preview := i.finding.Match
	if len(preview) > 48 {
		preview = preview[:45] + "..."
	}
	mark := "○"
	if i.finding.Selected {
		mark = "●"
	}
	return mark + " " + preview
}

type pushDoneMsg struct{ err error }

type Model struct {
	git       git.Runner
	scanOpts  scan.Options
	repoState git.RepoState
	findings  []detect.Finding
	list      list.Model
	preview   viewport.Model
	layout    layout
	screen    screen
	summary   string
	warnForce string
	pushPlan  pushPlan
	pushMsg   string
	cleanActs guide.CleanActions
	amendPub  bool
	err       error
}

func New(g git.Runner, opts scan.Options, findings []detect.Finding) Model {
	st, _ := g.RepoState()
	items := make([]list.Item, len(findings))
	for i, f := range findings {
		items[i] = findingItem{finding: f, index: i}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(colorAccent).Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(colorAccent)

	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	m := Model{
		git:       g,
		scanOpts:  opts,
		repoState: st,
		findings:  findings,
		list:      l,
		preview:   viewport.New(0, 0),
		screen:    screenMain,
	}
	m.syncPreview()
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout = computeLayout(msg.Width, msg.Height)
		m.list.SetWidth(m.layout.leftWidth - 4)
		m.list.SetHeight(m.layout.innerH)
		m.preview.Width = m.layout.rightWidth - 4
		m.preview.Height = m.layout.innerH
		return m, nil

	case pushDoneMsg:
		if msg.err != nil {
			m.pushMsg = "Push failed: " + msg.err.Error()
		} else if m.pushPlan.Force {
			m.pushMsg = "✓ Pushed with --force-with-lease"
		} else {
			m.pushMsg = "✓ Pushed to " + m.repoState.Upstream
		}
		st, _ := m.git.RepoState()
		m.repoState = st
		m.screen = screenSummary
		return m, nil

	case tea.KeyMsg:
		if m.screen != screenMain {
			return m.updateOverlay(msg)
		}
		return m.updateMain(msg)
	}

	if m.screen == screenMain {
		var cmd tea.Cmd
		prev := m.list.Index()
		m.list, cmd = m.list.Update(msg)
		if m.list.Index() != prev {
			m.syncPreview()
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "n", "N":
		for i := range m.findings {
			m.findings[i].Selected = false
		}
		m.refreshList()
	case "c":
		if m.selectedCount() == 0 {
			return m, nil
		}
		m.warnForce = m.forcePushWarning()
		m.amendPub = m.warnForce != ""
		m.screen = screenConfirm
	case "r", "R":
		return m.rescan()
	case "?":
		m.screen = screenHelp
	}
	return m, nil
}

func (m Model) updateOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenConfirm:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "b":
			m.screen = screenMain
		case "c":
			result, err := clean.Apply(m.git, clean.Options{Findings: m.findings, ForceAmend: true})
			if err != nil {
				m.err = err
				return m, tea.Quit
			}
			m.summary = result.Summary
			m.cleanActs = result.Actions
			st, _ := m.git.RepoState()
			m.repoState = st
			m.pushPlan = pushPlanAfterClean(st, result.Actions, m.amendPub)
			m.screen = screenSummary
		}
	case screenSummary:
		switch msg.String() {
		case "q", "ctrl+c", "enter":
			return m, tea.Quit
		case "p", "P":
			if m.pushPlan.Available {
				if m.pushPlan.Force {
					m.screen = screenPush
				} else {
					return m, m.pushCmd(false)
				}
			}
		case "r", "R":
			return m.rescan()
		case "esc", "b":
			m.screen = screenMain
		}
	case screenPush:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "b":
			m.screen = screenSummary
		case "p", "P":
			return m, m.pushCmd(true)
		}
	case screenHelp:
		switch msg.String() {
		case "q", "esc", "?", "enter":
			m.screen = screenMain
		}
	}
	return m, nil
}

func (m Model) pushCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if force {
			err = m.git.PushForceWithLease()
		} else {
			err = m.git.Push()
		}
		return pushDoneMsg{err: err}
	}
}

func (m Model) rescan() (tea.Model, tea.Cmd) {
	result, err := scan.Run(m.git, m.scanOpts)
	if err != nil {
		m.err = err
		return m, tea.Quit
	}
	st, _ := m.git.RepoState()
	m.findings = result.Findings
	m.repoState = st
	m.summary = ""
	m.pushMsg = ""
	m.pushPlan = pushPlan{}
	m.screen = screenMain
	m.refreshList()
	return m, nil
}

func (m *Model) refreshList() {
	items := make([]list.Item, len(m.findings))
	for i, f := range m.findings {
		items[i] = findingItem{finding: f, index: i}
	}
	m.list.SetItems(items)
	m.syncPreview()
}

func (m *Model) syncPreview() {
	idx := m.list.Index()
	if idx < 0 || idx >= len(m.findings) {
		m.preview.SetContent(subStyle.Render("Select a finding to preview the cleanup."))
		return
	}
	f := m.findings[idx]
	var b strings.Builder
	b.WriteString(panelTitlePassive.Render("Preview") + "\n\n")
	b.WriteString(PreviewDiff(f))
	if f.Selected {
		b.WriteString("\n" + okStyle.Render("● selected for cleanup"))
	} else {
		b.WriteString("\n" + subStyle.Render("○ not selected"))
	}
	m.preview.SetContent(b.String())
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

func (m Model) forcePushWarning() string {
	for _, f := range m.findings {
		if !f.Selected {
			continue
		}
		if f.Source == detect.SourceCommitMessage || f.Source == detect.SourceCommitTrailer {
			if m.repoState.AmendingRewritesPushedCommit() {
				return "HEAD is already on the remote. Cleaning rewrites published history."
			}
		}
	}
	return ""
}

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render("Error: "+m.err.Error()) + "\n"
	}
	base := m.renderMain()
	if m.screen == screenMain {
		return base
	}
	return m.renderWithOverlay(base)
}

func (m Model) renderMain() string {
	var b strings.Builder
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n")
	b.WriteString(m.renderSplit())
	b.WriteString("\n")
	b.WriteString(m.renderFooter())
	return b.String()
}

func (m Model) renderStatusBar() string {
	st := m.repoState
	parts := []string{
		titleStyle.Render("Coauthor Cleaner"),
		fmt.Sprintf("branch %s", st.Branch),
	}
	if st.UpstreamExists {
		parts = append(parts, fmt.Sprintf("%s ↑%d ↓%d", st.Upstream, st.Ahead, st.Behind))
	} else {
		parts = append(parts, "no upstream")
	}
	if st.HasStagedChanges {
		parts = append(parts, "staged")
	}
	count := len(m.findings)
	sel := m.selectedCount()
	if count == 0 {
		parts = append(parts, okStyle.Render("clean"))
	} else {
		parts = append(parts, fmt.Sprintf("%d findings · %d selected", count, sel))
	}
	return statusBarStyle.Width(m.layout.width - 2).Render(strings.Join(parts, "  │  "))
}

func (m Model) renderSplit() string {
	ly := m.layout
	if ly.width == 0 {
		ly = computeLayout(80, 24)
	}

	leftTitle := panelTitleActive.Render("Findings")

	leftBody := m.list.View()
	if len(m.findings) == 0 {
		leftBody = subStyle.Render("✓ No AI attribution markers found.\n\nPress r to rescan.")
	}

	left := panelActiveStyle.
		Width(ly.leftWidth).
		Height(ly.bodyHeight).
		Render(leftTitle + "\n" + leftBody)

	right := panelPassiveStyle.
		Width(ly.rightWidth).
		Height(ly.bodyHeight).
		Render(m.preview.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func (m Model) renderFooter() string {
	keys := m.contextKeys()
	return footerStyle.Width(m.layout.width - 2).Render(keys)
}

func (m Model) contextKeys() string {
	switch m.screen {
	case screenConfirm:
		return "confirm  │  c clean  │  esc back  │  q quit"
	case screenSummary:
		if m.pushPlan.Available {
			return "done  │  p push  │  r rescan  │  q quit"
		}
		return "done  │  r rescan  │  q quit"
	case screenPush:
		return "⚠ force push  │  p push --force-with-lease  │  esc back  │  q quit"
	case screenHelp:
		return "help  │  esc close"
	default:
		return "space toggle  │  a all  │  n none  │  c clean  │  r rescan  │  ? help  │  q quit"
	}
}

func (m Model) renderWithOverlay(base string) string {
	var content string
	switch m.screen {
	case screenConfirm:
		content = m.viewConfirmModal()
	case screenSummary:
		content = m.viewSummaryModal()
	case screenPush:
		content = m.viewPushModal()
	case screenHelp:
		content = m.viewHelpModal()
	default:
		return base
	}

	dimmed := backdropStyle.Render(base)
	modal := modalStyle.Render(content)
	placed := lipgloss.Place(
		m.layout.width, m.layout.height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
	// Overlay: show dimmed base with modal centered (replace full view for clarity)
	_ = dimmed
	return placed + "\n" + footerStyle.Width(m.layout.width-2).Render(m.contextKeys())
}

func (m Model) viewConfirmModal() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Apply cleanups?") + "\n\n")
	if m.warnForce != "" {
		b.WriteString(warnStyle.Render("⚠ "+m.warnForce) + "\n\n")
	}
	for _, f := range m.findings {
		if !f.Selected {
			continue
		}
		line := "  ● " + string(f.Source)
		if f.FilePath != "" {
			line += ": " + f.FilePath
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m Model) viewSummaryModal() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Done") + "\n\n")
	b.WriteString(m.summary)
	if m.pushMsg != "" {
		b.WriteString("\n" + m.pushMsg + "\n")
	}
	if m.pushPlan.Available && m.pushPlan.Force {
		b.WriteString("\n" + warnStyle.Render("Push requires --force-with-lease."))
	} else if m.pushPlan.Reason != "" {
		b.WriteString("\n" + subStyle.Render(m.pushPlan.Reason))
	}
	return b.String()
}

func (m Model) viewPushModal() string {
	return warnStyle.Render("Force push?") + "\n\n" +
		"This rewrites the commit on GitHub.\n" +
		"Only continue if no one else based work\n" +
		"on the old commit."
}

func (m Model) viewHelpModal() string {
	return titleStyle.Render("Keybindings") + "\n\n" +
		"Navigation\n" +
		"  ↑/↓     move selection\n" +
		"  space   toggle finding\n" +
		"  a       select all\n" +
		"  n       select none\n\n" +
		"Actions\n" +
		"  c       clean selected\n" +
		"  p       push (after clean)\n" +
		"  r       rescan repo\n" +
		"  ?       this help\n" +
		"  q       quit"
}

func Run(g git.Runner, opts scan.Options) error {
	opts.Staged = true
	opts.Commit = "HEAD"
	result, err := scan.Run(g, opts)
	if err != nil {
		return err
	}
	m := New(g, opts, result.Findings)
	m.layout = computeLayout(80, 24)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
