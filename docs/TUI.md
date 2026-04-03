# TUI Guide

`project` starts an interactive terminal UI built with Bubble Tea + Bubbles.

## Purpose

The TUI provides a single screen for project operations.

## Features

- Displays registered projects in a table.
- Shows each project's optional description.
- Shows per-project status (`stopped` or `running (pid X)`).
- Shows listening TCP ports for running projects.
- Shows project start time for running projects.
- Allows start/stop actions for the selected row.
- Allows registering a new project from the TUI.
- Allows editing selected project name, description, directory, and script.
- Allows removing a selected project.
- Allows running `git pull` for the selected project.
- Supports manual refresh.
- Shows command output in a bottom output panel.

## Keyboard controls

- `up/down` or `j/k`: move selection
- `enter` or `space`: start/stop selected project
- `n`: register a new project (`name`, `dir`, `script`, `description`)
- `e`: edit selected project (`name`, `dir`, `script`, `description`)
- `x`: remove selected project (press twice to confirm)
- `g`: run `git pull` in selected project directory
- `o`: open selected project directory in Output proxy terminal
- `r`: refresh table state
- `q` or `ctrl+c`: quit

## Data flow

1. On startup, TUI loads projects from `store.Load()`.
2. For each project, it checks runtime state via `runner.IsRunning()`.
3. For running projects, it resolves ports via `runner.Ports()`.
4. On toggle action:
   - if stopped -> `runner.Start(project)`
   - if running -> `runner.Stop(name)`

## Storage

Project metadata and run state are stored in:

- `~/.config/project/projects.json`
- `~/.config/project/state.json`

## Known limitations

- Port discovery relies on `lsof` and `pgrep` (macOS-focused behavior).
- Refresh is manual (`r`) rather than periodic auto-refresh.
- Editing/removal is blocked while a project is running.
