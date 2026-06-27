# Architecture

This document describes the full design of `zebra` — its layering, data model,
state model, and rendering strategy. For *what* the product does and *why*, see
[`PRD.md`](PRD.md). For user-facing keybindings, see the
[`README`](../README.md).

## Design goals

1. **Instant.** The first diff is loaded synchronously before the first frame is
   drawn — there is no spinner on startup for a normal repository.
2. **Pure where it counts.** Everything except the git wrapper is a pure
   function of input, so each layer is testable in isolation without touching
   the filesystem.
3. **Keyboard-first.** Every action is reachable from the keyboard; there is no
   mouse path and no hidden modal state.
4. **No git library.** `zebra` shells out to the `git` binary. It inherits the
   user's git configuration and behavior exactly, and carries no parsing of the
   object database.

## Layering

Three layers, each in its own package, ordered by purity. The lower layers must
**not** depend on the ones above:

```
          ┌──────────────────────────────┐
 impure   │  internal/ui                 │  Bubble Tea Model + render
   ▲      │   - model.go   (state)       │  (pure, except loader.go)
   │      │   - render.go  (view)        │
   │      │   - loader.go  (the one seam)│ ── LoadFiles ──┐
   │      └──────────────────────────────┘                │
   │      ┌──────────────────────────────┐                │
   │      │  internal/diff               │  pure parser    │
   │      │   - Parse(raw) → []File      │ <───────────────┘
   │      └──────────────────────────────┘   raw diff text
   │      ┌──────────────────────────────┐
 pure ▼   │  internal/git                │  spawns `git`, returns RAW output
          └──────────────────────────────┘
```

### `internal/git` — the I/O boundary

The only code that touches the filesystem or spawns processes. It is a thin
wrapper that runs `git diff` variants and returns **raw** unified-diff text; it
never parses diffs.

- `IsRepo(path)` — guards startup.
- `Scope` enum maps to git arguments:
  - `ScopeWorkingTree` → `git diff`
  - `ScopeStaged` → `git diff --cached`
  - `ScopeAll` (worktree + staged) → `git diff HEAD`
- Binary files are detected from git's own `Binary files ... differ` line; the
  wrapper surfaces the before/after sizes rather than diff content.

Because this is the only impure layer, it is the only one whose tests need real
fixture repositories (created in a `t.TempDir()`).

### `internal/diff` — the pure parser

`Parse` turns raw unified-diff text into a structured model and contains **all**
diff understanding. No I/O, no shelling out.

```
File   { Path, OldPath, Status (M/A/D/R), Added, Removed,
         IsBinary, SizeBefore, SizeAfter, Hunks []Hunk }
Hunk   { Header, OldStart, NewStart, Lines []Line }
Line   { Kind (Context/Added/Removed), Text, OldNum, NewNum,
         WhitespaceOnly bool }
```

The parser tracks old/new line numbers as it walks each hunk and flags lines
whose only change is whitespace, so the UI can highlight formatting-only edits.
It is tested purely: raw diff strings in, model structs out, covering empty
diffs, binary files, renames, adds, deletes, whitespace-only changes, and
multi-hunk files.

### `internal/ui` — the Bubble Tea layer

The root `Model` plus pure render functions. Follows the Elm architecture used
throughout Bubble Tea:

- `Model.Update` is a **value receiver** returning a new `Model` — mutate the
  local copy and return it; never share pointers to mutable state.
- `loader.go` holds `LoadFiles`, the **single seam** that ties the parser to the
  git wrapper. It is the *only* place in the package that performs I/O.
  Everything else in `ui` is a pure function of model state, which is what makes
  the render and navigation logic snapshot-testable without git.

## Startup & the event loop

`main.go`:

1. Resolves the `--repo` flag to an absolute path (default `.`).
2. Verifies it is a git repository (`git.IsRepo`).
3. Loads the initial working-tree diff **synchronously** (`ui.LoadFiles`) so the
   first frame renders instantly.
4. Hands the populated model to the Bubble Tea program and runs the event loop.

After startup, scope changes reload **asynchronously** via a `loadFilesCmd`
`tea.Cmd`, so the UI never blocks on git while the user is interacting.

## State model

A single root `Model` carries all state — there are no hidden globals.

### Focus

A single ordered `focusOrder` slice defines the `Tab` cycle through focusable
controls:

```
sidebar → scope → view → whitespace → diff
```

`←`/`→` move focus between the sidebar and the diff panel. Adding a control
means updating both `focusOrder` and `activateFocused`.

### Modal input

Two modal text-input states capture printable text **before** normal
keybindings apply:

- `filterActive` — filters the file list (when the sidebar is focused).
- `searchActive` — searches within the current diff (when the diff is focused).

`Ctrl+F` activates whichever is relevant to the focused panel; `Esc` closes it.
`tea.KeyPressMsg` exposes the printable rune via `msg.Text`. All keyboard
handling funnels through `handleKey`.

### Search

Each panel carries its own independent search state — query string, the list of
match positions, and the current match index — so the sidebar filter and the
diff search never interfere.

## Rendering strategy

Rendering is **offset-based, not widget-based**. The diff panel is rendered to a
single string, and `scrollLines` slices that string by `diffScroll` to produce
the visible window. This keeps the layout logic in one place and makes the whole
view a pure function of state.

The consequence: several helpers must stay in **lockstep** with the actual
render layout in `render.go`, because inline and side-by-side modes produce
different row counts for the same hunk:

- `hunkStartRows` — where each hunk begins, for `n`/`p` navigation.
- `renderedRowText` — the text of a rendered row, for search match mapping.

If you change how rows are laid out, update these helpers in tandem or hunk
navigation and search will drift out of alignment with what's on screen.

### View modes

Inline and side-by-side are two **rendering strategies over the same model** —
there is no separate data path. The toggle is a single piece of root-model
state. Side-by-side shows the original file on the left and the new file on the
right, each with independent line numbers; added files leave the left panel
empty, deleted files leave the right panel empty.

### Whitespace & color

Whitespace rendering is applied at render time: spaces become `·` and tabs
become `→` when the toggle is on. Lines flagged `WhitespaceOnly` by the parser
get a distinct background. The MVP palette is fixed (added: green; removed: red;
line numbers: gray; whitespace-only: violet; statuses M/A/D/R: yellow/green/
red/blue) — themes are out of scope.

## Testing strategy

Each layer is tested against its purity boundary:

- **`git`** — real fixture repos; assert the raw diff string for known states.
- **`diff`** — pure unit tests; raw diff in, model out.
- **`ui`** — drive `Update` with key messages and assert resulting state;
  snapshot-test render functions against known models. No git in the loop.

Keep new logic pure enough to test the same way rather than reaching for the
filesystem.

## Distribution

- **Go:** standard `go install` against the `main` package at the repo root.
- **Binaries:** GitHub Actions builds cross-platform binaries on tag push
  (`v*`) — see [`.github/workflows/release.yml`](../.github/workflows/release.yml).
- **Package managers (future):** an npm wrapper with platform-specific optional
  dependencies, plus Homebrew/Scoop, sourced from GitHub Releases — scaffolded
  in [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml).

## Out of scope (by design)

Interactive staging, custom themes, arbitrary commit/branch diffs, language
syntax highlighting, mouse support, i18n, blame, and merge-conflict resolution
are intentionally excluded from the MVP. See [`PRD.md`](PRD.md) for the rationale.
