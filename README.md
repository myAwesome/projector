# projector

CLI tool to register, run, and stop local web app projects.

## Install

```sh
go install proj
```

Or build locally:

```sh
go build -o proj .
# optionally move to PATH
mv proj /usr/local/bin/proj
```

## Commands

### Register a project

```sh
proj register --name <name> --dir <path> --script <command>
```

| Flag | Description |
|------|-------------|
| `--name` | Unique project name |
| `--dir` | Project directory (script runs from here) |
| `--script` | Launch command or path to script |

Examples:

```sh
proj register --name "kanban" --dir ~/projects/kanban --script "./start.sh"
proj register --name "blog" --dir ~/projects/blog --script "docker compose up"
proj register --name "api" --dir ~/projects/api --script "go run ./cmd/server"
```

### List projects

```sh
proj list
```

Shows all registered projects with their current status, listened ports, and start time:

```
NAME    STATUS               PORTS       STARTED   SCRIPT
kanban  running (pid 12345)  3000, 5432  10:00:01  ./start.sh
blog    stopped              -           -          docker compose up
```

### Run a project

```sh
proj run <name>
```

Starts the project's script in the background. All child processes (server, db, client) run in the same process group, so `stop` can terminate them all at once.

### Stop a project

```sh
proj stop <name>
```

Sends `SIGTERM` to the entire process group, stopping the project and all its children.

### Interactive TUI

```sh
proj tui
```

Keyboard controls:

- `up/down` or `j/k`: move selection
- `enter` or `space`: start/stop selected project
- `r`: refresh project status
- `q`: quit

## Storage

| File | Contents |
|------|----------|
| `~/.config/proj/projects.json` | Registered projects |
| `~/.config/proj/state.json` | Running process state (PID, PGID, start time) |

## Requirements

- macOS (port detection uses `lsof` and `pgrep`)
- Go 1.24.2+
