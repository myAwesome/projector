# TUI Guide

`proj tui` starts an interactive terminal UI built with Bubble Tea + Bubbles.

## Purpose

The TUI provides a single screen for project operations that otherwise require multiple CLI calls.

## Features

- Displays registered projects in a table.
- Shows per-project status (`stopped` or `running (pid X)`).
- Shows listening TCP ports for running projects.
- Shows project start time for running projects.
- Allows start/stop actions for the selected row.
- Supports manual refresh.

## Keyboard controls

- `up/down` or `j/k`: move selection
- `enter` or `space`: start/stop selected project
- `r`: refresh table state
- `q` or `ctrl+c`: quit

## Data flow

1. On startup, TUI loads projects from `store.Load()`.
2. For each project, it checks runtime state via `runner.IsRunning()`.
3. For running projects, it resolves ports via `runner.Ports()`.
4. On toggle action:
   - if stopped -> `runner.Start(project)`
   - if running -> `runner.Stop(name)`

## Scope and compatibility

- The TUI is additive and does not replace existing CLI commands.
- Project metadata and run state are stored in the same files:
  - `~/.config/proj/projects.json`
  - `~/.config/proj/state.json`

## Known limitations

- Port discovery relies on `lsof` and `pgrep` (macOS-focused behavior).
- Refresh is manual (`r`) rather than periodic auto-refresh.
