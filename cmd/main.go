package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/LuisCabantac/pomodoro-cli/internal/config"
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
			fmt.Println("Usage: pomodoro-cli [-s <preset>]")
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
