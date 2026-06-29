# journal

A small terminal journaling tool. It opens dated markdown entries in your editor
and records them as git commits.

## Install

```bash
go install github.com/manningc33/journal/cmd/journal@latest
# or, from a checkout:
make install
```

## Usage

```bash
journal today                 # open today's entry (creating it if needed)
journal today -1              # open yesterday's entry (-N = N days ago)
journal commit <file> <msg>   # append <msg>, lint, and git-commit the entry
journal help
```

Entries are stored as `<journal_dir>/<YYYY>/<MM-mon>/<DD-day>.md`, e.g.
`2026/06-jun/28-sun.md`. New entries start with a header line and open with the
cursor on line 3. `commit` derives the commit date from the file's path,
validates the path matches the layout, and commits one entry at a time as
`YYYY-MM-DD <msg>`.

## Configuration

Optional, at `~/.config/journal/config.toml` (or `$XDG_CONFIG_HOME/journal/`).
Without it, the built-in defaults reproduce the original scripts exactly — only
`journal_dir` (default `~/journal`) typically needs setting. See
[`config.example.toml`](config.example.toml) for every option, including the
strftime-style date formats and the editor/linter commands.

Set `JOURNAL_CONFIG` to point at a config file in a non-standard location.

## Development

```bash
make test        # unit tests
make test-race   # tests with the race detector
make cover       # coverage report
make all         # fmt + vet + test + build
```
