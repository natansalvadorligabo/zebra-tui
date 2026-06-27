package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/salvadorligabo/zebra-tui/internal/diff"
)

const sidebarWidth = 30

// View implements tea.Model.
func (m Model) View() string {
	if m.loadErr != nil {
		return styleMessage.Render("error: " + m.loadErr.Error())
	}

	control := RenderControlBar(m.scope.String(), m.view.String(), m.showWhitespace)
	sidebar := RenderSidebar(m.files, m.selected, m.filter, m.focus == focusSidebar)
	diffPanel := m.renderDiffPanel()

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(sidebarWidth).Render(sidebar),
		" ",
		diffPanel,
	)

	parts := []string{control, body}
	if m.searchActive {
		parts = append(parts, styleFocused.Render("search: "+m.search))
	}
	parts = append(parts, RenderFooter())
	return strings.Join(parts, "\n")
}

// renderDiffPanel renders the diff for the currently opened file.
func (m Model) renderDiffPanel() string {
	files := m.filteredFiles()
	if len(files) == 0 {
		return RenderEmptyMessage()
	}
	idx := m.opened
	if idx < 0 || idx >= len(files) {
		idx = 0
	}
	opts := RenderOpts{
		Width:          m.diffWidth(),
		ShowWhitespace: m.showWhitespace,
	}
	if m.searchActive {
		opts.Search = m.search
	}
	f := files[idx]
	body := m.renderFile(f, opts)
	return scrollLines(body, m.diffScroll, m.diffHeight())
}

func (m Model) renderFile(f diff.File, opts RenderOpts) string {
	if m.view == viewSideBySide {
		return RenderSideBySide(f, opts)
	}
	return RenderInline(f, opts)
}

func (m Model) diffWidth() int {
	w := m.width - sidebarWidth - 1
	if w < 20 {
		w = 20
	}
	return w
}

func (m Model) diffHeight() int {
	// Reserve rows for control bar and footer.
	h := m.height - 3
	if h < 1 {
		h = 1
	}
	return h
}

// scrollLines returns at most height lines of s starting at offset.
func scrollLines(s string, offset, height int) string {
	lines := strings.Split(s, "\n")
	if offset < 0 {
		offset = 0
	}
	if offset > len(lines) {
		offset = len(lines)
	}
	lines = lines[offset:]
	if height > 0 && len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
