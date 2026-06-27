# Releasing

How a `zebra` release reaches GitHub Releases and npm. The pipeline is wired but
the human runs the publishing steps — agents never push.

## One-time setup

1. **npm account**: log in at <https://www.npmjs.com>. The package name
   `zebra-tui` is unscoped and currently free.
2. **npm automation token**: npm → _Access Tokens_ → _Generate New Token_ →
   **Automation** (bypasses 2FA in CI). Copy it.
3. **GitHub secret**: repo → _Settings → Secrets and variables → Actions → New
   repository secret_ → name `NPM_TOKEN`, value = the automation token.

## Per-release flow

Version `vX.Y.Z` (first release: `v0.0.1`). `npm/package.json` version is set
automatically from the tag during publish, so you don't edit it by hand.

1. **Merge `dev` into `main`** and push:

   ```sh
   git checkout main
   git merge --ff-only dev      # or open a PR dev -> main and merge it
   git push origin main
   ```

2. **Tag and push the tag** — this is the only action; it triggers the whole
   pipeline:

   ```sh
   git tag -a v0.1.0 -m "v0.1.0"
   git push origin v0.1.0
   ```

   The **Release** workflow (`.github/workflows/release.yml`) runs two jobs in
   one run:
   - `goreleaser` — cross-compiles the binaries and publishes a public GitHub
     Release with `zebra-<os>-<arch>` assets + `checksums.txt`.
   - `npm` (`needs: goreleaser`) — sets `npm/package.json` from the tag and runs
     `npm publish`.

   Both happen automatically because they share one workflow run; npm publishing
   can't live in a separate workflow keyed on the release event, since a Release
   published by GoReleaser's bot token does not trigger other workflows.

   To re-publish npm for an existing tag without re-tagging, run the **Deploy**
   workflow by hand: _Actions → Deploy → Run workflow_, entering the tag.

3. **Verify**:

   ```sh
   npm install -g zebra-tui
   zebra --version            # -> zebra v0.1.0
   ```

## Validating locally before tagging

GoReleaser can dry-run the whole build without publishing:

```sh
goreleaser release --snapshot --clean
ls dist/                     # inspect the produced binary names
```

## Notes

- Asset names in `.goreleaser.yaml` (`zebra-<os>-<arch>[.exe]`) must stay in
  sync with the platform map in `npm/install.js`; the postinstall builds the
  download URL from them.
- Homebrew and Scoop are not wired yet; add jobs to `deploy.yml` once a tap /
  bucket exists.
