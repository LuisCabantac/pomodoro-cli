# Pomodoro CLI

A terminal-based Pomodoro timer with a polished TUI, preset system, and cross-platform desktop notifications.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)
[![Release](https://img.shields.io/github/v/release/LuisCabantac/pomodoro-cli?include_prereleases)](https://github.com/LuisCabantac/pomodoro-cli/releases)

## Features

- **TUI timer** – real-time countdown with progress bar and cycle tracking (`[2/4]`)
- **Preset system** – switch between timer configurations (Classic 25/5, Short 15/3, or your own)
- **Desktop notifications** – native alerts on Linux (notify-send), macOS (osascript), and Windows (toast)
- **Skip breaks** – press `s` to cut a break short
- **Safe quit** – `q` / `Esc` / `Ctrl+C` prompts for confirmation before exiting

## Install

### From source (Go 1.26+)

```bash
go install github.com/LuisCabantac/pomodoro-cli@latest
```

Or clone and build:

```bash
git clone https://github.com/LuisCabantac/pomodoro-cli.git
cd pomodoro-cli
go build -o pomodoro-cli .
```

### Pre-built binary

Download the latest release from the [releases page](https://github.com/LuisCabantac/pomodoro-cli/releases).

## Usage

```
Usage: pomodoro-cli [flags] [command]

Flags:
  -h, --help          Show this help message
  -s, --start <id>    Start a timer with a preset by ID (skips preset list)

Commands:
  create              Create a new preset
```

### Quick start

```bash
./pomodoro-cli
```

Start directly with a preset:

```bash
./pomodoro-cli -s classic
./pomodoro-cli -s short
```

### Create a custom preset

```bash
./pomodoro-cli create \
  --name "Deep Focus" \
  --work 50 \
  --short_break 10 \
  --long_break 20 \
  --cycle 3
```

### Controls

| Key | Context | Action |
|---|---|---|
| `↑` / `↓` | Preset list | Navigate presets |
| `Enter` | Preset list | Select preset |
| `Space` | Timer | Pause / resume |
| `s` | Break | Skip break |
| `q` / `Esc` / `Ctrl+C` | Timer | Request quit |
| `y` / `Esc` | Quit prompt | Confirm / cancel quit |

## Presets

On first run, two built-in presets are created automatically:

| ID | Name | Work | Short Break | Long Break | Cycle |
|---|---|---|---|---|---|
| `classic` | Classic | 25 min | 5 min | 10 min | 4 |
| `short` | Short | 15 min | 3 min | 10 min | 4 |

Custom presets are stored in `~/.config/pomodoro-cli/presets.json` (Linux), `~/Library/Application Support/pomodoro-cli/presets.json` (macOS), or `%APPDATA%\pomodoro-cli\presets.json` (Windows).

## Build

```bash
go build -o pomodoro-cli .
```

Cross-compile for Linux amd64 (static binary, CGO disabled):

```bash
./scripts/build.sh
```

## License

[MIT](LICENSE)
