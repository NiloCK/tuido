# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Tuido is a terminal UI (Go + Charm's Bubble Tea) for managing todos in the [x]it!
format. It recursively scans a directory tree for items in source files, presents
them in an interactive TUI, and can also operate non-interactively via subcommands.

## Commands

There is no Makefile; use the Go toolchain directly.

```sh
go run .                 # run the TUI in the current directory (dogfooding mode)
go build -o tuido .      # build a binary
go test ./...            # run all unit tests
go test ./tui/ -run TestFoo   # run a single test (package + -run regex)
go vet ./... && go fmt ./...  # vet and format
```

`go test ./...` is what CI runs (`.github/workflows/test.yml`). A handful of
version/upgrade tests are gated behind the `integration` build tag (e.g.
`utils/versioning_live_test.go`) and require network access; run them with
`go test -tags integration ./utils/`. The throwaway `main` programs under
`internal/testing/` are manual harnesses for the self-upgrade flow, not part of
the build or test suite.

### CLI subcommands

`main.go` dispatches on `os.Args[1]` before falling through to the TUI:

- `list [-z|--zzz] [-a|--all] [--max N] [path]` — print items to stdout (default path `.`)
- `create <text>` / `add <text>` — append a new item to the configured `writeto`
- `init` — interactive wizard (`tui.RunInitWizard`) to create a local/global config
- `version`, `help`

If no config is found at startup, `tui.ConfigFound` is false and the TUI refuses
to launch, directing the user to `tuido init`. The non-interactive subcommands
still work without a config.

## Architecture

Three packages plus the CLI shim:

- **`main.go`** — CLI argument dispatch and the non-interactive `list`/`create`
  implementations. Note these reuse exported TUI helpers (`tui.GetFiles`,
  `tui.GetItems`, `tui.SortItems`, `tui.GetConfig*`) rather than reimplementing
  discovery — so the file-scanning logic lives in `tui/`, not `tuido/`.
- **`tui/`** — the Bubble Tea app. `tui.go` defines the `tui` model and the
  `mode` enum that drives everything: `navigation`, `filter`, `edit`, `help`,
  `pomo`, `nag`, `peek`, `configViewer`, `upgrade`. `update.go` routes input by
  mode, `view.go` renders by mode. File discovery (`GetFiles`/`GetItems`) and
  sorting (`SortItems`) live here. `upgrade.go` is the in-app self-update flow.
- **`tuido/`** — pure item/domain logic, no TUI dependency. `Item` (`file`,
  `line`, `raw`) knows its status (`Open`/`Ongoing`/`Checked`/`Obsolete`), parses
  and rewrites its own line in the source file, and handles tags. `time.go`
  expands date shorthands (`d2w` → `#due=YYYY-MM-DD`).
- **`utils/`** — version checking and the self-upgrade download/replace machinery
  (`version` is injected at build time via goreleaser ldflags).

### Item model — the important invariant

An `Item` is a *pointer into a source file* (`file` + `line` + `raw`), not an
owned record. Mutations (`SetText`, status changes, snoozing) rewrite that one
line in place to preserve the rest of the file. When touching item logic, respect
that line-based editing contract — don't rewrite whole files.

### Config resolution

`tui/init.go` `loadConfig()` resolves with priority **cwd `.tuido` > global
`tuido.conf`** (in `os.UserConfigDir()`). `applyConfig` only overrides defaults
for fields the file actually sets. The `config` struct (`tui/config.go`):

- `extensions` — file types to parse (default `xit,md,txt`)
- `writeto` — file (append as lines) or directory (`YYYY-MM-DD.xit` per day);
  defaults to `~/.tuido/`
- `exclude` — directory/file globs to skip during traversal
- `frictionThreshold` — item count before the `nag` deterrent screen appears

To add a config field: add to the struct, parse it in `parseConfig`, honor it in
`applyConfig`, and set a default in `runConfig`.

## Dogfooding

This repo's own `.tuido` sets `extensions=go,md`, `writeto=readme.md`,
`frictionThreshold=3`. **Running tuido here writes new items into `readme.md`**
(its Roadmap section) and parses todos out of `.go` files — keep that in mind
before creating items during testing.

## Conventions

- The README is lowercase `readme.md` (it doubles as the `writeto` target).
- `git` write operations (add/commit) are off-limits per the user's global rules;
  read-only git inspection is fine.
