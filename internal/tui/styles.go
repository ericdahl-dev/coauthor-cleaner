package tui

import "github.com/charmbracelet/lipgloss"

// Palette inspired by tui-design: muted base, accent for focus, warn for danger.
var (
	colorAccent = lipgloss.Color("205")
	colorMuted  = lipgloss.Color("240")
	colorDim    = lipgloss.Color("238")
	colorWarn   = lipgloss.Color("214")
	colorError  = lipgloss.Color("196")
	colorOK     = lipgloss.Color("42")
	colorBorder = lipgloss.Color("63")
	colorPanel  = lipgloss.Color("60")
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	subStyle   = lipgloss.NewStyle().Foreground(colorMuted)
	helpStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	warnStyle  = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(colorOK)
	errorStyle = lipgloss.NewStyle().Foreground(colorError).Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	panelActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1)

	panelPassiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPanel).
				Padding(0, 1)

	panelTitleActive = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	panelTitlePassive = lipgloss.NewStyle().Bold(true).Foreground(colorMuted)

	footerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorDim).
			Foreground(colorMuted).
			Padding(0, 1)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2).
			Width(56)

	backdropStyle = lipgloss.NewStyle().Foreground(colorDim)
)
