# PRD: Zebra — Git Diff TUI

## Problem Statement

Developers who review local changes frequently use `git diff` in the terminal, but the default output is hard to read: no line numbers, no side-by-side comparison, poor whitespace visibility, and no way to navigate between changed files without leaving the terminal. Existing alternatives either require a GUI (VSCode, GitKraken) or are slow and hard to install (web-based tools). There is no fast, beautiful, keyboard-driven TUI that makes reviewing git diffs a first-class terminal experience.

## Solution

`zebra` is a terminal UI tool built with Bubble Tea (Go) that replaces `git diff` with a rich, navigable interface. It opens instantly in the current git repository, shows a sidebar of modified files with status indicators, and renders the diff with line numbers, syntax-aware coloring, side-by-side or inline modes, and visible whitespace — all keyboard-driven without leaving the terminal.

## User Stories

1. As a developer, I want to run `zebra` in any git repository directory and have the TUI open immediately, so that I can start reviewing changes without any configuration.
2. As a developer, I want to pass `--repo <path>` to open a diff for a repository in a different directory, so that I can review changes without changing my working directory.
3. As a developer, I want to see a sidebar listing all modified files with their status (M/A/D/R) and change counts (+N -N), so that I can quickly assess the scope of changes before diving in.
4. As a developer, I want status indicators in the sidebar to be color-coded (M yellow, A green, D red, R blue), so that I can distinguish file statuses at a glance.
5. As a developer, I want the diff scope to default to the working tree (unstaged changes), so that I see my in-progress work immediately.
6. As a developer, I want to toggle the diff scope between working tree, staged, and working tree + staged from the top control bar, so that I can review different states of my changes without restarting the tool.
7. As a developer, I want a persistent top control bar showing the current scope, view mode, and whitespace toggle, so that I always know what I'm looking at.
8. As a developer, I want to see keybinding hints in the footer, so that I can discover available actions without reading documentation.
9. As a developer, I want to navigate between sidebar and diff panel using left/right arrow keys, so that I can move focus intuitively.
10. As a developer, I want to select a file in the sidebar by pressing Enter and have the diff load without shifting focus away from the sidebar, so that I can keep browsing the file list fluidly.
11. As a developer, I want to use Tab to cycle focus between all focusable elements (sidebar, toggles, checkboxes, diff panel), so that I can reach any control with the keyboard.
12. As a developer, I want to scroll the focused panel with j/k or arrow keys, so that I can use vim-style or standard navigation interchangeably.
13. As a developer, I want to jump between diff hunks using n (next) and p (previous), so that I can skip unchanged context and focus on what actually changed.
14. As a developer, I want to toggle between inline and side-by-side diff modes with a global keybinding, so that I can choose the best layout for the file I'm reviewing.
15. As a developer, I want the side-by-side mode to show the original file on the left and the new file on the right, each with independent line numbers, so that line references match the actual files.
16. As a developer, I want added files in side-by-side mode to show an empty left panel and the new content on the right, so that the layout remains consistent regardless of file status.
17. As a developer, I want deleted files in side-by-side mode to show the original content on the left and an empty right panel, so that the layout remains consistent.
18. As a developer, I want to toggle "show whitespace" from the control bar, so that I can spot whitespace-only changes that would otherwise be invisible.
19. As a developer, I want spaces rendered as `·` and tabs as `→` when whitespace is visible, so that I can see exactly what whitespace characters are present.
20. As a developer, I want lines whose only change is whitespace to be highlighted with a distinct background color, so that I can immediately identify "safe" formatting-only changes.
21. As a developer, I want to search within the current file's diff using Ctrl+F, so that I can find a specific function or variable without scrolling manually.
22. As a developer, I want all search matches in the diff to be highlighted simultaneously, so that I can see the full distribution of results at once.
23. As a developer, I want to navigate between search matches using up/down arrow keys or Enter (next), so that I can move through results without learning new keybindings.
24. As a developer, I want to close the search input with Esc and return to normal diff navigation, so that I can resume keyboard navigation smoothly.
25. As a developer, I want to filter the file list in the sidebar using Ctrl+F when the sidebar is focused, so that I can find a specific file quickly in large changesets.
26. As a developer, I want the sidebar filter to be contextual — Ctrl+F behavior depends on which panel is focused — so that one keybinding serves both search needs.
27. As a developer, I want binary files to appear in the sidebar with the message "Binary file: diff not available" and the file size before and after (e.g. `24KB → 31KB`), so that I'm informed without the tool crashing or showing garbage.
28. As a developer, I want the TUI to open normally when the working tree is clean, showing an empty sidebar and a "Nothing to diff: working tree is clean" message in the diff panel, so that the state is communicated clearly without the tool exiting abruptly.
29. As a developer, I want to install `zebra` via `go install github.com/<user>/zebra@latest`, so that I can get the tool instantly if I have Go installed.
30. As a developer, I want to install `zebra` via `npm install -g zebra-tui`, so that I can get the tool without having Go installed.
31. As a developer, I want pre-compiled binaries for Linux, macOS, and Windows available on GitHub Releases, so that I can download and use the tool without any package manager.

## Implementation Decisions

- **Language & framework:** Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI, [Lip Gloss](https://github.com/charmbracelet/lipgloss) for styling.
- **Git integration:** Shell out to the `git` binary to retrieve diffs (`git diff`, `git diff --cached`, `git diff HEAD`). The git layer is a thin wrapper that runs the command and returns raw output; it is not responsible for parsing.
- **Diff parsing:** A dedicated parser transforms the raw unified diff format into an internal model (list of files → list of hunks → list of lines with type: context/added/removed). This module is pure and has no I/O.
- **Internal diff model:** Each file entry carries: path, status (M/A/D/R), added line count, removed line count, is-binary flag, binary size before/after, and a list of parsed hunks.
- **View modes:** Inline and side-by-side are rendering strategies applied to the same internal model — no separate data path. Toggle is a global piece of state in the root model.
- **Focus model:** A single `focusedPanel` enum in the root model tracks which element has keyboard focus. Left/right arrows update this value. Tab cycles through a defined ordered list of focusable elements.
- **Search state:** Each panel (sidebar filter, diff search) carries its own independent search state (query string, match positions, current match index). Ctrl+F activates the relevant one based on `focusedPanel`.
- **Whitespace rendering:** Applied at render time — the raw line string is transformed to replace spaces with `·` and tabs with `→` when the whitespace toggle is on. Lines detected as whitespace-only diffs receive a distinct Lip Gloss background style.
- **Color palette (fixed for MVP):** Added lines: green background; removed lines: red background; context lines: default; line numbers: gray; whitespace-only diff lines: purple/violet background; status M: yellow, A: green, D: red, R: blue.
- **Binary file handling:** The git layer detects binary files from the diff output (`Binary files ... differ`) and populates the `is-binary` flag and sizes. The diff panel renders the static message instead of hunk content.
- **Distribution — Go:** Standard `go install` with a `main` package at the repo root. GitHub Actions builds cross-platform binaries on tag push using `goreleaser`.
- **Distribution — npm:** A wrapper npm package (`zebra-tui`) with platform-specific optional dependencies (`zebra-tui-linux-x64`, `zebra-tui-darwin-arm64`, `zebra-tui-win32-x64`, etc.). The `postinstall` script selects and extracts the correct binary. Binaries are sourced from GitHub Releases.
- **CLI entrypoint:** Single command `zebra`, flags: `--repo <path>` (default: `.`). No subcommands for MVP.

## Testing Decisions

- **What makes a good test:** Tests should assert external behavior — what goes in, what comes out — not internal state or implementation details. Prefer table-driven tests with representative fixtures.
- **Git layer:** Tested against real fixture git repositories (created in `TestMain` or `t.TempDir()`). Assert that the correct raw diff string is returned for known repo states. Do not mock git.
- **Diff parser:** Pure unit tests. Input: raw unified diff strings (stored as `.diff` fixture files). Output: the internal diff model struct. Cover edge cases: empty diff, binary file, renamed file, added file, deleted file, whitespace-only changes, multi-hunk files.
- **Render/view:** Snapshot tests on the string output of the render functions given a known model state. Cover: inline mode, side-by-side mode, show whitespace on/off, binary file message, empty repo message, search highlight state.
- **Focus/navigation:** Unit tests on the root model's `Update` function — send key messages, assert the resulting model state (focused panel, selected file, scroll offset). No I/O needed.

## Out of Scope

- Interactive staging (selecting hunks or lines to `git add`)
- Custom themes or user-configurable color palettes
- Diff between arbitrary commits or branches (only working tree / staged / HEAD for MVP)
- Syntax highlighting by language
- Mouse support
- i18n / localization
- Git blame integration
- Merge conflict resolution

## Further Notes

- The project name is **zebra** — the binary is `zebra`, the npm package is `zebra-tui`, the GitHub repo will be `zebra-tui` (to avoid collision with existing `zebra` packages).
- The tool should feel instant — diff parsing and rendering should happen before the first frame is drawn if possible, or show a loading state only if the repo is very large.
- Bubble Tea's architecture (Elm-style Model/Update/View) maps naturally to this design: one root model, child models per panel, messages for key events and git data loaded.
