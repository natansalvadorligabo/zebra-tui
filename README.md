# zebra

A fast, keyboard-driven terminal UI for reviewing local git diffs, built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea).

`zebra` replaces `git diff` with a navigable interface: a sidebar of changed
files with status indicators and change counts, and a diff panel with line
numbers, inline or side-by-side layout, visible whitespace, and search.

## Install

```sh
go install github.com/salvadorligabo/zebra-tui@latest
```

## Usage

Run inside any git repository:

```sh
zebra
```

Review a repository elsewhere:

```sh
zebra --repo /path/to/repo
```

## Keybindings

| Key            | Action                                             |
| -------------- | -------------------------------------------------- |
| `Tab`          | Cycle focus (sidebar → toggles → diff)             |
| `←` / `→`      | Move focus between sidebar and diff panel          |
| `↑`/`↓`, `j`/`k` | Move sidebar selection / scroll the diff panel   |
| `Enter`        | Open the selected file (sidebar) / activate toggle |
| `n` / `p`      | Jump to next / previous hunk                       |
| `v`            | Toggle inline ↔ side-by-side view                  |
| `w`            | Toggle visible whitespace                          |
| `Ctrl+F`       | Search the diff / filter the file list (by focus)  |
| `Esc`          | Close search or filter                             |
| `q`, `Ctrl+C`  | Quit                                               |

The top control bar shows the current scope (working tree / staged /
worktree+staged), view mode, and whitespace toggle. Focus the scope control and
press `Space`/`Enter` to cycle scopes.

## Development

```sh
go test ./...
```

The codebase is split into pure, independently tested layers:

- `internal/git` — thin wrapper that shells out to `git` for raw diff output.
- `internal/diff` — pure parser turning unified diffs into a structured model.
- `internal/ui` — Bubble Tea model, update logic, and render functions.
