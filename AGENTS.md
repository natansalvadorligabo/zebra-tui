## What this is

`zebra` is a keyboard-driven terminal UI for reviewing local git diffs, built on
Bubble Tea v2 (`charm.land/bubbletea/v2`) and Lip Gloss v2. It shells out to the
`git` binary; it does not use a git library.

## Commands

```sh
go test ./...                       # run all tests
go test ./internal/diff -run Parse  # run a single test by name
go build -o zebra.exe .             # build the binary
go run . --repo /path/to/repo       # run against another repo (defaults to .)
go vet ./...                        # vet
```

There is no Makefile or lint config; use the standard `go` toolchain. Go 1.26.

## Git workflow

**Always check the current branch before committing.** Run `git branch
--show-current` (or `git status`) and confirm you are on the intended branch —
feature work lands on `dev`, never directly on `main`. The branch can change
between turns, so verify every time rather than assuming.

**Never `git push`.** Only the maintainer pushes. Agents may stage, commit, and
create branches locally, but pushing to any remote is off-limits — leave pushing
to the human.

## Architecture

Three layers, each in its own package, ordered by purity. Respect these
boundaries — the lower layers must not depend on the ones above:

- **`internal/git`** — the only code that touches the filesystem or spawns
  processes. Thin wrapper that runs `git diff` variants and returns *raw* output;
  it never parses diffs. `Scope` (working tree / staged / worktree+staged) maps
  to the git args.
- **`internal/diff`** — a *pure* parser (no I/O, no shelling out). `Parse` turns
  raw unified-diff text into `[]diff.File` → `Hunk` → `Line`, tracking old/new
  line numbers and detecting whitespace-only changes. All diff understanding
  lives here.
- **`internal/ui`** — the Bubble Tea `Model` plus pure render functions.
  `LoadFiles` (in `loader.go`) is the single seam that ties the parser to the git
  wrapper; it is the *only* place in `ui` that performs I/O. Everything else in
  the package is a pure function of model state.

`main.go` loads the initial working-tree diff synchronously before starting the
program so the first frame renders instantly, then hands off to the Bubble Tea
event loop (scope changes reload asynchronously via `loadFilesCmd`).

### UI model conventions

- `Model.Update` is a value receiver returning a new `Model` — follow the
  Elm-architecture style already in place; mutate the local copy and return it.
- Keyboard handling funnels through `handleKey`. Two modal input states
  (`filterActive` for the sidebar, `searchActive` for the diff) capture text
  *before* normal keybindings apply; `tea.KeyPressMsg` exposes printable text via
  `msg.Text`.
- `focusOrder` defines the Tab cycle through focusable controls
  (sidebar → scope → view → whitespace → diff). Adding a control means updating
  this slice and `activateFocused`.
- Rendering is offset-based, not widget-based: the diff panel is rendered to a
  single string and `scrollLines` slices it by `diffScroll`. Several helpers
  (`hunkStartRows`, `renderedRowText`) must stay in lockstep with the actual
  render layout in `render.go` (inline vs side-by-side produce different row
  counts) — if you change how rows are laid out, update these in tandem or hunk
  navigation and search will drift.

## Testing

**TDD is mandatory.** For every feature or bug fix, invoke the project `tdd`
skill (`.agents/skills/tdd/SKILL.md`) and follow it: write one failing test
first (red), confirm it fails for the right reason, then write the minimal code
to pass (green), and refactor. Work in vertical slices — one test → one
implementation → repeat — never write all tests up front or implementation
before its covering test exists.

**After every bug fix or feature**, run the full suite via the `run-tests`
skill (`.agents/skills/run-tests/SKILL.md`) before declaring the task done.
The suite must be green — no regressions allowed.

Each layer is tested in isolation against its purity boundary: `diff` tests feed
raw diff strings and assert on the parsed model; `ui` tests drive the model and
render functions without touching git. Keep new logic pure enough to test the
same way rather than reaching for the filesystem.

## Reference

`docs/PRD.md` holds the product requirements; `README.md` documents user-facing
keybindings and scopes.
