package ui

import "github.com/charmbracelet/lipgloss"

// Fixed MVP color palette (see PRD "Color palette").
var (
	styleAdded          = lipgloss.NewStyle().Background(lipgloss.Color("22")).Foreground(lipgloss.Color("15")) // green bg
	styleRemoved        = lipgloss.NewStyle().Background(lipgloss.Color("52")).Foreground(lipgloss.Color("15")) // red bg
	styleContext        = lipgloss.NewStyle()                                                                   // default
	styleLineNumber     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))                                 // gray
	styleWhitespaceOnly = lipgloss.NewStyle().Background(lipgloss.Color("54")).Foreground(lipgloss.Color("15")) // violet bg
	styleSearchMatch    = lipgloss.NewStyle().Background(lipgloss.Color("178")).Foreground(lipgloss.Color("0")) // highlight
	styleMessage        = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)

	// Sidebar status colors.
	styleStatusM = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // yellow
	styleStatusA = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))  // green
	styleStatusD = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
	styleStatusR = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))  // blue

	styleSelected = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("238"))
	styleFocused  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
)
