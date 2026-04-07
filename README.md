# projector

TUI-only tool to register, run, monitor, and stop local web app projects.

## Install

```sh
go install project
```

Or build locally:

```sh
go build -o project .
# optionally move to PATH
mv project /usr/local/bin/project
```

## Quick start

```sh
project
```

Running `project` starts the interactive terminal UI.

## TUI controls

- `up/down` or `j/k`: move selection
- `enter` or `space`: start/stop selected project
- `n`: register a new project (`name`, `dir`, `script`, `description`)
- `e`: edit selected project (`name`, `dir`, `script`, `description`)
- `x`: remove selected project (press twice to confirm)
- `g`: run `git pull` in selected project directory
- `o`: open project directory in Output proxy terminal
- `r`: refresh project status
- `q`: quit

The Output panel shows command and proxy-terminal output.

For details, see [`docs/TUI.md`](docs/TUI.md).

## Storage

| File | Contents |
|------|----------|
| `~/.config/project/projects.json` | Registered projects |
| `~/.config/project/state.json` | Running process state (PID, PGID, start time) |

## Requirements

- macOS (port detection uses `lsof` and `pgrep`)
- Go 1.24.2+
