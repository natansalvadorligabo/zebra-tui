package ui

import "charm.land/lipgloss/v2"

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

	// Zebra theme: a monochrome black/white mascot with a single bright accent.
	colorInk    = lipgloss.Color("231") // near-white "stripe"
	colorAccent = lipgloss.Color("39")  // bright blue accent
	colorMuted  = lipgloss.Color("244") // dimmed text

	borderColor        = lipgloss.Color("240")
	borderColorFocused = colorAccent

	// Header.
	styleWordmark = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("236")).Foreground(colorInk)
	styleTagline  = lipgloss.NewStyle().Italic(true).Background(lipgloss.Color("236")).Foreground(colorMuted)

	// Highlighted control-bar field for the currently focused Tab target.
	styleControlActive = lipgloss.NewStyle().Bold(true).Background(colorAccent).Foreground(lipgloss.Color("0"))

	// Empty-state splash.
	styleSplashArt  = lipgloss.NewStyle().Foreground(colorMuted)
	styleSplashWord = lipgloss.NewStyle().Bold(true).Foreground(colorInk)
)

// panelStyle returns the rounded border for a body panel, brightened when the
// panel currently holds focus.
func panelStyle(focused bool) lipgloss.Style {
	c := borderColor
	if focused {
		c = borderColorFocused
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c)
}
