# projector

CLI/TUI tool to register, run, monitor, and stop local web app projects.

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
# 1) register a project
proj register --name "kanban" --dir ~/projects/kanban --script "./start.sh" --description "Main Kanban app"

# 2) run in interactive mode
proj tui
```

## Commands

### Register a project

```sh
project register --name <name> --dir <path> --script <command> [--description <text>]
```

| Flag | Description |
|------|-------------|
| `--name` | Unique project name |
| `--dir` | Project directory (script runs from here) |
| `--script` | Launch command or path to script |
| `--description`, `-d` | Optional short description shown in list/TUI |

Examples:

```sh
project register --name "kanban" --dir ~/projects/kanban --script "./start.sh" --description "Main Kanban app"
project register --name "blog" --dir ~/projects/blog --script "docker compose up" -d "Marketing blog stack"
project register --name "api" --dir ~/projects/api --script "go run ./cmd/server" -d "Backend API"
```

### List projects

```sh
project list
```

Shows all registered projects with their current status, listened ports, and start time:

```
NAME    DESCRIPTION          STATUS               PORTS       STARTED   SCRIPT
kanban  Main Kanban app      running (pid 12345)  3000, 5432  10:00:01  ./start.sh
blog    Marketing blog stack stopped              -           -          docker compose up
```

### Run a project

```sh
project run <name>
```

Starts the project's script in the background. All child processes (server, db, client) run in the same process group, so `stop` can terminate them all at once.

### Stop a project

```sh
project stop <name>
```

Sends `SIGTERM` to the entire process group, stopping the project and all its children.

### Interactive TUI

```sh
proj tui
```

Shows all registered projects in one screen and allows starting/stopping the selected project.

Keyboard controls:

- `up/down` or `j/k`: move selection
- `enter` or `space`: start/stop selected project
- `n`: register a new project (`name`, `dir`, `script`, `description`)
- `e`: edit selected project (`name`, `dir`, `script`, `description`)
- `x`: remove selected project (press twice to confirm)
- `g`: run `git pull` in selected project directory
- `o`: open project dir in terminal
- `r`: refresh project status
- `q`: quit

For a more detailed TUI guide, see [`docs/TUI.md`](docs/TUI.md).

## Storage

| File | Contents |
|------|----------|
| `~/.config/project/projects.json` | Registered projects |
| `~/.config/project/state.json` | Running process state (PID, PGID, start time) |

## Requirements

- macOS (port detection uses `lsof` and `pgrep`)
- Go 1.24.2+

## Notes

- Existing CLI commands (`list`, `run`, `stop`) are still supported.
- TUI actions reuse the same runner/store internals as the CLI commands.
