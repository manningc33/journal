# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Go CLI (`journal`) for terminal journaling — a rewrite of two original Bash
scripts (`today.sh` / `jcommit.sh`). Their on-disk layout and formatting (path,
header, commit message) were deliberate and must be preserved byte-for-byte;
they now live as defaults in `internal/config`, overridable via config.

## Commands

```bash
make test        # go test ./...
make test-race   # race detector
make cover       # coverage report (coverage.out)
make build       # -> bin/journal
make all         # fmt + vet + test + build
go test ./internal/entry -run TestDateFromPath   # a single package / test
```

CI (`.github/workflows/ci.yml`) gates on `gofmt -l` being empty, `go vet`,
`go test -race`, and golangci-lint — run `make fmt vet test` before assuming green.

## Architecture

Single binary, `today` / `commit` subcommands. The entrypoint (`cmd/journal`)
and `internal/cli` only parse args and orchestrate; all logic lives in small,
single-purpose packages:

- `internal/datefmt` — converts strftime patterns (e.g. `%Y/%m-%b`) to Go layouts.
  **One converter serves both formatting and parsing**, which is what makes the
  date↔path round trip exact. Pure, no deps.
- `internal/config` — loads TOML over `Defaults()` (which encodes the original
  scripts' behavior); a missing file is not an error. `FormatConfig.Render`
  applies the strftime format plus the lowercase/squeeze transforms.
- `internal/entry` — pure date↔path mapping. `DateFromPath` validates a path by
  **re-deriving the canonical path from the parsed date and comparing** — this
  is the layout check (Go's `time.Parse` does NOT validate weekday, so don't rely
  on that; see the round-trip instead).
- `internal/editor`, `internal/linter`, `internal/vcs` — the three side effects.
- `internal/run` — `Runner` interface wrapping `os/exec`. `run.Fake` (non-test
  file, so it's shared) records calls; this is how the side-effecting packages
  and `cli` are tested without spawning processes.

## Conventions to preserve

- New external-command interactions go through `run.Runner`, never `os/exec`
  directly outside `internal/run`. This keeps everything testable with `run.Fake`.
- Any change to default formatting must keep parity with the original scripts'
  output (e.g. entry `2026/06-jun/28-sun.md` with header
  `# june 28, 2026 (sunday 16:43)`). Tests in `config`/`entry`/`datefmt` assert
  this exact output — update them deliberately, not to make red go green.
- `commit` must derive its date from the path (validated), commit exactly one
  entry, and mutate the file only after validation succeeds.
- Module path is `github.com/manningc33/journal`.
