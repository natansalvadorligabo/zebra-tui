package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// zebraArt is the ASCII mascot shown on the empty-state splash. Keep it in a
// raw string literal so the stripes stay aligned in a monospace cell grid.
const zebraArt = `             ______
          .-'      '-.
        .'   ()       '.
       /   //////////   \
      |   ////////////   |==<
      |   ////////////   |
       \   //////////   /
        '.            .'
          \  |    |  /
           | |    | |
           |_|    |_|`

// RenderTitleBar renders the full-width header showing the "zebra" wordmark,
// painted across the whole terminal width.
func RenderTitleBar(width int) string {
	if width <= 0 {
		width = 80
	}
	left := styleWordmark.Render(" zebra ")
	pad := width - lipgloss.Width(left)
	if pad < 0 {
		pad = 0
	}
	return left + styleTagline.Render(strings.Repeat(" ", pad))
}

// RenderSplash centers the zebra mascot, wordmark, and the empty-tree message
// within the given diff-panel dimensions.
func RenderSplash(width, height int) string {
	if width <= 0 {
		width = 40
	}
	if height <= 0 {
		height = 12
	}
	block := lipgloss.JoinVertical(
		lipgloss.Center,
		styleSplashArt.Render(zebraArt),
		"",
		styleSplashWord.Render("z · e · b · r · a"),
		styleMessage.Render(emptyMessage),
	)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}
