package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/LuisCabantac/pomodoro-cli/internal/config"
	"github.com/LuisCabantac/pomodoro-cli/internal/preset"
	"github.com/LuisCabantac/pomodoro-cli/internal/tui"
)

func main() {
	presets, err := config.LoadItems()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	var startPresetID string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			fmt.Println("Usage: pomodoro-cli [flags] [command]")
			fmt.Println()
			fmt.Println("Flags:")
			fmt.Println("  -s, --start <preset>   start timer with the given preset")
			fmt.Println("  -h, --help             show this help")
			fmt.Println()
			fmt.Println("Commands:")
			fmt.Println("  create   create a new preset with --name, --work, --short_break, --long_break, --cycle")
			os.Exit(0)

		case "s", "start":
			if i+1 < len(args) {
				startPresetID = args[i+1]
				found := false
				for _, p := range presets {
					if p.ID == startPresetID {
						found = true
						break
					}
				}

				if !found {
					fmt.Fprintf(os.Stderr, "pomodoro-cli: preset '%s' not found\n", startPresetID)
					fmt.Fprintf(os.Stderr, "Available presets:\n")
					for _, p := range presets {
						fmt.Fprintf(os.Stderr, "- %s\n", p.ID)
					}
					fmt.Fprintln(os.Stderr)
					os.Exit(1)
				}

				i++
			} else {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: option '%s' requires a preset name\n", args[i])
				fmt.Fprintf(os.Stderr, "See 'pomodoro-cli --help' for usage.\n")
				os.Exit(1)
			}

		case "c", "create":
			if len(args) < 11 {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: '%s' requires --name, --work, --short_break, --long_break, and --cycle\n", args[i])
				fmt.Fprintf(os.Stderr, "See 'pomodoro-cli --help' for usage.\n")
				os.Exit(1)
			}

			var (
				id, name                           string
				work, shortBreak, longBreak, cycle int
			)
			for i := 1; i < len(args); i += 2 {
				switch args[i] {
				case "--name":
					name = args[i+1]
					for _, p := range presets {
						if p.Name == name {
							fmt.Fprintf(os.Stderr, "pomodoro-cli: preset '%s' already exists\n", name)
							os.Exit(1)
							break
						}
						id = strings.Join(strings.Split(strings.ReplaceAll(strings.ToLower(name), "'", ""), " "), "_")
					}
				case "--work":
					work, err = strconv.Atoi(args[i+1])
					if err != nil {
						fmt.Fprintf(os.Stderr, "pomodoro-cli: invalid value for '--work': '%s' is not a number\n", args[i+1])
						os.Exit(1)
					}
				case "--short_break":
					shortBreak, err = strconv.Atoi(args[i+1])
					if err != nil {
						fmt.Fprintf(os.Stderr, "pomodoro-cli: invalid value for '--short_break': '%s' is not a number\n", args[i+1])
						os.Exit(1)
					}
				case "--long_break":
					longBreak, err = strconv.Atoi(args[i+1])
					if err != nil {
						fmt.Fprintf(os.Stderr, "pomodoro-cli: invalid value for '--long_break': '%s' is not a number\n", args[i+1])
						os.Exit(1)
					}
				case "--cycle":
					cycle, err = strconv.Atoi(args[i+1])
					if err != nil {
						fmt.Fprintf(os.Stderr, "pomodoro-cli: invalid value for '--cycle': '%s' is not a number\n", args[i+1])
						os.Exit(1)
					}
				default:
					fmt.Fprintf(os.Stderr, "pomodoro-cli: unknown flag '%s'\n", args[i])
					os.Exit(1)
				}
			}

			if name == "" {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: 'create' requires --name\n")
				os.Exit(1)
			}
			if id == "" {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: 'create' requires --name\n")
				os.Exit(1)
			}
			if work == 0 {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: 'create' requires --work\n")
				os.Exit(1)
			}
			if shortBreak == 0 {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: 'create' requires --short_break\n")
				os.Exit(1)
			}
			if longBreak == 0 {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: 'create' requires --long_break\n")
				os.Exit(1)
			}
			if cycle == 0 {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: 'create' requires --cycle\n")
				os.Exit(1)
			}

			presets = append(presets, preset.Preset{
				ID:                   id,
				Name:                 name,
				WorkMin:              work,
				ShortBreakMin:        shortBreak,
				LongBreakMin:         longBreak,
				CycleBeforeLongBreak: cycle,
			})

			if err := config.WriteItems(presets); err != nil {
				fmt.Fprintf(os.Stderr, "pomodoro-cli: failed to save preset: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("pomodoro-cli: \"%s\" preset created\n", name)
			os.Exit(0)

		default:
			fmt.Fprintf(os.Stderr, "pomodoro-cli: unknown option '%s'\n", args[i])
			fmt.Fprintf(os.Stderr, "See 'pomodoro-cli --help' for usage.\n")
			os.Exit(1)
		}
	}

	items := make([]list.Item, len(presets))
	for i, p := range presets {
		items[i] = p
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Presets"
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select preset")),
		}
	}

	var m tui.Model
	if startPresetID != "" {
		m = tui.NewModelWithPreset(l, startPresetID)
	} else {
		m = tui.NewModel(l)
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
