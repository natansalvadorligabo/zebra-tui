# zebra-tui

A fast, keyboard-driven terminal UI for reviewing your **local git diffs**.

This npm package is a thin wrapper: on install it downloads the prebuilt
`zebra` binary for your platform from the matching
[GitHub Release](https://github.com/natansalvadorligabo/zebra-tui/releases)
and exposes it as the `zebra` command.

```sh
npm install -g zebra-tui
zebra            # run inside any git repository
zebra --version
```

`git` must be installed and on your `PATH` — `zebra` shells out to it.

Supported platforms: Linux, macOS, and Windows on `x64` / `arm64`. On other
platforms, install from source instead:

```sh
go install github.com/natansalvadorligabo/zebra-tui@latest
```

Full documentation, keybindings, and screenshots:
<https://github.com/natansalvadorligabo/zebra-tui>.

## License

[MIT](https://github.com/natansalvadorligabo/zebra-tui/blob/main/LICENSE)
