package tui

import "github.com/charmbracelet/lipgloss"

// Dark terminal palette aligned with the Pages hero mock.
var (
	colorAccent = lipgloss.Color("#D6FF5F")
	colorViolet = lipgloss.Color("#9B7CFF")
	colorMuted  = lipgloss.Color("#8B8791")
	colorDim    = lipgloss.Color("#3C3A43")
	colorWarn   = lipgloss.Color("#FFD166")
	colorError  = lipgloss.Color("#FF6B81")
	colorOK     = lipgloss.Color("#D6FF5F")
	colorBorder = lipgloss.Color("#363341")
	colorPanel  = lipgloss.Color("#1A1B22")
	colorPanel2 = lipgloss.Color("#101116")
	colorText   = lipgloss.Color("#F7F5EF")
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorText)
	subStyle   = lipgloss.NewStyle().Foreground(colorMuted)
	helpStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	warnStyle  = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(colorOK).Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(colorError).Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colorBorder).
			Background(colorPanel2).
			Padding(0, 1)

	panelActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Background(colorPanel2).
				Padding(1, 2)

	panelPassiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPanel).
				Background(colorPanel2).
				Padding(1, 2)

	panelTitleActive = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)
	panelTitlePassive = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorMuted)

	badgeStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorViolet).
			Padding(0, 1)

	logoMarkStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	codeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Background(lipgloss.Color("#08090D")).
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorDim).
			Foreground(colorMuted).
			Padding(0, 1)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Background(colorPanel2).
			Padding(1, 2).
			Width(56)

	backdropStyle = lipgloss.NewStyle().Foreground(colorDim)
)
