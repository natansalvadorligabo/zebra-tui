---
name: release
description: >
  Preflight a zebra release and hand the maintainer the exact push/publish
  commands. Use when the user wants to cut a release, ship a version, tag a
  version (e.g. v0.0.1), or publish to GitHub Releases or npm.
---

# Release

Getting a tag to the repo and out through the pipeline. The agent does the
**preflight** — proving the branch is releasable and the asset wiring is
correct — then the **handoff**: the maintainer pushes and publishes, because
this repo's rule is **never push** (see AGENTS.md). GoReleaser and the Deploy
workflow do the actual publishing; the agent never tags-and-pushes.

The canonical maintainer runbook is [`docs/RELEASING.md`](../../../docs/RELEASING.md) —
this skill is the agent's preflight, not a second copy of it.

## Steps

### 1. Detect current version and propose the bump

Find the latest released version, then let the **conventional commits** since it
choose the **semver** bump — never ask the user to pick the number blind.

```sh
git describe --tags --abbrev=0           # latest tag; empty -> no release yet
git log <latest-tag>..HEAD --oneline     # commits to classify (omit range if none)
```

Classify the range and apply the highest bump present:

- `feat!:` / `fix!:` / a `BREAKING CHANGE:` footer -> **major** (`X`)
- any `feat:` -> **minor** (`Y`)
- only `fix:` (or `perf:`) -> **patch** (`Z`)
- only `docs:`/`chore:`/`build:`/`test:` -> patch, and say the release is
  optional since nothing user-facing changed.

With no prior tag, the first release is `v0.0.1`. Releases are cut from `main`;
work must already be merged or ready to merge from `dev`.

**Completion criterion:** the current version is stated, the bump is justified by
the commits, and the resulting `vX.Y.Z` is confirmed with the user.

### 2. Preflight the branch

On the release branch:

```sh
go vet ./... && go test ./... && go build ./...
git status --short
```

Also confirm the module path matches the GitHub repo (`go.mod` and imports use
the same org as `git remote -v`) — a mismatch breaks `go install`.

**Completion criterion:** vet/tests/build exit 0, working tree clean, module
path matches the remote.

### 3. Verify the release wiring is consistent

The npm postinstall builds its download URL from the GoReleaser asset names, so
the two must agree:

- `.goreleaser.yaml` `archives.name_template` produces `zebra-<os>-<arch>` raw
  binaries (`formats: [binary]`).
- `npm/install.js`'s `OS`/`ARCH` maps yield those same names.
- `goreleaser check` passes. If `goreleaser` is installed, also dry-run and
  inspect the output names:

  ```sh
  goreleaser check
  goreleaser release --snapshot --clean && ls dist/
  ```

**Completion criterion:** every `dist/` binary name is reproducible from
`npm/install.js`, and `goreleaser check` exits 0.

### 4. Hand off — do not push

Present the maintainer the command sequence from `docs/RELEASING.md`: merge
`dev`→`main`, then `git tag` + `git push` the tag. That one push runs the
Release workflow, whose `goreleaser` and `npm` jobs publish the GitHub Release
and the npm package in a single run — pushing the tag is the only action. Flag
any one-time setup still missing (the `NPM_TOKEN` GitHub secret).

**Completion criterion:** the maintainer has the exact push/tag/publish commands
and the list of any unmet prerequisites. The agent has pushed nothing.
