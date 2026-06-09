package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/LuisCabantac/pomodoro-cli/internal/tui"
)

func NewModel() tui.Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
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

	m := tui.Model{
		List:     l,
		Progress: progress.New(progress.WithDefaultBlend()),
		Active:   false,
	}

	return m
}

func main() {
	m := NewModel()

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
