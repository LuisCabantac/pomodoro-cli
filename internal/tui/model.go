package tui

import (
	"fmt"
	"math"
	"os/exec"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/LuisCabantac/pomodoro-cli/internal/config"
	"github.com/LuisCabantac/pomodoro-cli/internal/preset"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type screen int
type state int
type tickMsg time.Time

const (
	screenList screen = iota
	screenTimer
)

const (
	stateWork state = iota
	stateShortBreak
	stateLongBreak
)

type Model struct {
	List      list.Model
	Choice    string
	Progress  progress.Model
	screen    screen
	state     state
	cycle     int
	Active    bool
	startTime time.Time
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return config.LoadItemsCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case preset.PresetsLoadedMsg:
		m.Active = true
		m.List.SetItems(preset.PresetsToItems(msg.Presets))
		return m, nil

	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.screen == screenList {
				i, ok := m.List.SelectedItem().(preset.Preset)
				if ok {
					m.Choice = i.ID
					m.screen = screenTimer
					m.state = stateWork
					m.cycle = 0
					m.Active = true
					m.startTime = time.Now()
					return m, tickCmd()
				}
			}

			if m.screen == screenTimer && m.Progress.Percent() == 1.0 {
				i, ok := m.List.SelectedItem().(preset.Preset)
				if ok {
					switch m.state {
					case stateWork:
						m.cycle++
						if m.cycle >= i.CycleBeforeLongBreak {
							m.state = stateLongBreak
							m.cycle = 0
						} else {
							m.state = stateShortBreak
						}
					case stateShortBreak:
						m.state = stateWork
					case stateLongBreak:
						m.state = stateWork
						m.cycle = 0
					}
				}
				m.Progress.SetPercent(0)
				m.startTime = time.Now()
				return m, tickCmd()
			}

			return m, nil

		case "space":
			if m.screen == screenTimer {
				m.Active = !m.Active
			}
			return m, nil
		}

	case tickMsg:
		if m.Progress.Percent() == 1.0 {
			var summary, body string
			switch m.state {
			case stateWork:
				summary, body = "Time's Up!", "Take a 5-minute break. Step away from the screen."
			case stateShortBreak:
				summary, body = "Break Over!", "Back to focus. Let's crush the next session!"
			case stateLongBreak:
				summary, body = "Long Break Done!", "Ready to start a new cycle. Let's go!"
			}
			notifyCmd := func() tea.Msg {
				_ = exec.Command("notify-send", "-u", "critical", "-i", "appointment-soon", "-a", "Pomodoro CLI", summary, body).Run()
				return nil
			}

			return m, notifyCmd
		}

		if !m.Active {
			return m, tickCmd()
		}

		i, ok := m.List.SelectedItem().(preset.Preset)
		if ok {
			duration := i.WorkMin
			if m.state == stateShortBreak {
				duration = i.ShortBreakMin
			}
			if m.state == stateLongBreak {
				duration = i.LongBreakMin
			}
			cmd := m.Progress.IncrPercent(1.0 / float64(duration*60))
			return m, tea.Batch(tickCmd(), cmd)
		}

	case progress.FrameMsg:
		var cmd tea.Cmd
		m.Progress, cmd = m.Progress.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()

		m.List.SetSize(msg.Width-h, msg.Height-v)

		m.Progress.SetWidth(msg.Width - 2*2 - 4)
		if m.Progress.Width() > 80 {
			m.Progress.SetWidth(80)
		}
	}

	var cmd tea.Cmd
	if m.screen == screenList {
		m.List, cmd = m.List.Update(msg)
	}
	return m, cmd
}

func (m Model) View() tea.View {
	switch m.screen {
	case screenTimer:
		i, ok := m.List.SelectedItem().(preset.Preset)
		if ok {
			total := i.WorkMin * 60
			if m.state == stateShortBreak {
				total = i.ShortBreakMin * 60
			} else if m.state == stateLongBreak {
				total = i.LongBreakMin * 60
			}
			elapsed := int(m.Progress.Percent() * float64(total))
			remaining := total - elapsed
			remainingStr := fmt.Sprintf("%dm%ds", remaining/60, remaining%60)
			info := ""

			paused := " (PAUSED)"
			if m.Active {
				paused = ""
			}

			if remaining > 0 {
				switch m.state {
				case stateWork:
					info = fmt.Sprintf("%s - %s%s", time.Now().Format("3:04PM"), remainingStr, paused)
				case stateShortBreak:
					info = fmt.Sprintf("%s - Short Break%s - %s", time.Now().Format("3:04PM"), paused, remainingStr)
				case stateLongBreak:
					info = fmt.Sprintf("%s - Long Break%s - %s", time.Now().Format("3:04PM"), paused, remainingStr)
				default:
					info = fmt.Sprintf("%s - %s%s", time.Now().Format("3:04PM"), remainingStr, paused)
				}
			} else {
				switch m.state {
				case stateWork:
					info = "Time for a break!"
				case stateShortBreak:
					info = "Back to focus"
				case stateLongBreak:
					info = "New cycle ready"
				default:
					info = "Complete!"
				}
			}

			content := lipgloss.JoinVertical(lipgloss.Left, info, m.Progress.View())
			v := tea.NewView(docStyle.Render(content))

			windowTitle := fmt.Sprintf("Pomodoro CLI: %s - %d%%%s", i.Name, int(math.Round(m.Progress.Percent()*100)), paused)
			if m.state == stateShortBreak {
				windowTitle = fmt.Sprintf("Pomodoro CLI: Short Break - %d%%%s", int(math.Round(m.Progress.Percent()*100)), paused)
			}
			if m.state == stateLongBreak {
				windowTitle = fmt.Sprintf("Pomodoro CLI: Long Break - %d%%%s", int(math.Round(m.Progress.Percent()*100)), paused)
			}

			v.WindowTitle = windowTitle
			v.AltScreen = true
			return v
		}
	}

	v := tea.NewView(docStyle.Render(m.List.View()))
	v.WindowTitle = "Pomodoro CLI"
	v.AltScreen = true
	return v
}
