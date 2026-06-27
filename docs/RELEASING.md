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

2. **Tag and push the tag** (this triggers the Release workflow):

   ```sh
   git tag -a v0.0.1 -m "v0.0.1"
   git push origin v0.0.1
   ```

   The **Release** workflow (`.github/workflows/release.yml`) runs GoReleaser,
   which cross-compiles the binaries and creates a **draft** GitHub Release with
   `zebra-<os>-<arch>` assets + `checksums.txt`.

3. **Review and publish the draft Release** on the GitHub Releases page. (The
   draft is a safety net for the first releases; once you trust the flow, set
   `release.draft: false` in `.goreleaser.yaml` to publish automatically.)

4. **Publish to npm**: repo → _Actions → Deploy → Run workflow_, enter the tag
   (`v0.0.1`). The **npm** job sets `npm/package.json` to `0.0.1` and runs
   `npm publish`. Requires the GitHub Release assets to already exist (step 3),
   because the package's postinstall downloads them.

5. **Verify**:

   ```sh
   npm install -g zebra-tui
   zebra --version            # -> zebra v0.0.1
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
- Homebrew and Scoop jobs in `deploy.yml` are stubs (`if: false`) for later.
