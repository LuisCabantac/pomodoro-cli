package tui

import (
	"fmt"
	"math"
	"os/exec"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
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

type TimerKeyMap struct {
	Pause key.Binding
	Quit  key.Binding
}

func (k TimerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Pause, k.Quit}
}
func (k TimerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Pause, k.Quit},
	}
}

type PausedTimerKeyMap struct {
	Continue key.Binding
	Quit     key.Binding
}

func (p PausedTimerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{p.Continue, p.Quit}
}
func (p PausedTimerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{p.Continue, p.Quit},
	}
}

type BreakTimerKeyMap struct {
	Skip key.Binding
	TimerKeyMap
}

func (k BreakTimerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Pause, k.Skip, k.Quit}
}
func (k BreakTimerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Pause, k.Skip, k.Quit}}
}

type QuittingKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func (q QuittingKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{q.Confirm, q.Cancel}
}
func (q QuittingKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{q.Confirm, q.Cancel}}
}

type FinishedTimerKeyMap struct {
	Continue key.Binding
	Quit     key.Binding
}

func (f FinishedTimerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{f.Continue, f.Quit}
}
func (f FinishedTimerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{f.Continue, f.Quit},
	}
}

type Model struct {
	List          list.Model
	Choice        string
	Progress      progress.Model
	screen        screen
	state         state
	cycle         int
	Active        bool
	startTime     time.Time
	Args          []string
	quitting      bool
	progressWidth int
	Help          help.Model
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
		switch m.screen {
		case screenList:
			switch msg.String() {
			case "enter":
				i, ok := m.List.SelectedItem().(preset.Preset)
				if ok {
					m.Choice = i.ID
					m.screen = screenTimer
					m.state = stateWork
					m.cycle = 0
					m.Active = true
					m.startTime = time.Now()
					m.Progress = progress.New(progress.WithDefaultBlend())
					m.Progress.SetWidth(m.progressWidth)
					return m, tickCmd()
				}
				return m, nil
			}

		case screenTimer:
			switch msg.String() {
			case "space":
				if m.Progress.Percent() != 1.0 {
					m.Active = !m.Active
					return m, nil
				}

			case "enter":
				if m.Progress.Percent() == 1.0 {
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

			case "s":
				if m.state == stateShortBreak || m.state == stateLongBreak {
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

			case "ctrl+c", "q":
				m.Active = false
				m.quitting = true
				return m, nil

			case "esc":
				m.quitting = false
				return m, nil

			case "y":
				if m.quitting {
					m.Active = false
					m.cycle = 0
					m.Progress.SetPercent(0)
					m.state = 0
					m.screen = screenList
					m.quitting = false
					return m, nil
				}
				return m, nil
			}
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
		w := msg.Width - 2*2 - 4
		if w > 80 {
			w = 80
		}

		m.List.SetSize(msg.Width-h, msg.Height-v)

		m.progressWidth = w
		m.Progress.SetWidth(w)
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenList:
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

			var keys help.KeyMap
			switch {
			case m.screen == screenTimer && m.quitting:
				keys = QuittingKeyMap{
				Confirm: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm quit")),
				Cancel:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
				}
			case m.state == stateShortBreak || m.state == stateLongBreak:
				keys = BreakTimerKeyMap{
					TimerKeyMap: TimerKeyMap{
						Pause: key.NewBinding(key.WithKeys("space"), key.WithHelp("␣", "pause")),
						Quit:  key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
					},
					Skip: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "skip")),
				}
			case m.screen == screenTimer && m.Progress.Percent() != 1.0 && m.Active:
				keys = TimerKeyMap{
					Pause: key.NewBinding(key.WithKeys("space"), key.WithHelp("␣", "pause")),
					Quit:  key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
				}
			case m.screen == screenTimer && !m.Active:
				keys = PausedTimerKeyMap{
					Continue: key.NewBinding(key.WithKeys("space"), key.WithHelp("␣", "continue")),
					Quit:     key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
				}
			case m.screen == screenTimer && m.Progress.Percent() >= 1.0:
				keys = FinishedTimerKeyMap{
					Continue: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "continue")),
					Quit:     key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
				}
			default:
				keys = TimerKeyMap{
					Pause: key.NewBinding(key.WithKeys("space"), key.WithHelp("␣", "pause")),
					Quit:  key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
				}
			}

			helpView := m.Help.View(keys)

			content := lipgloss.JoinVertical(lipgloss.Left, info, m.Progress.View(), helpView)
			v := tea.NewView(docStyle.Render(content))

			cycleInfo := ""
			if m.state == stateWork && i.CycleBeforeLongBreak > 0 {
				cycleInfo = fmt.Sprintf(" [%d/%d]", m.cycle+1, i.CycleBeforeLongBreak)
			}

			windowTitle := fmt.Sprintf("Pomodoro CLI: %s%s - %d%%%s", i.Name, cycleInfo, int(math.Round(m.Progress.Percent()*100)), paused)
			if m.state == stateShortBreak {
				windowTitle = fmt.Sprintf("Pomodoro CLI: Short Break - %d%%%s", int(math.Round(m.Progress.Percent()*100)), paused)
			}
			if m.state == stateLongBreak {
				windowTitle = fmt.Sprintf("Pomodoro CLI: Long Break - %d%%%s", int(math.Round(m.Progress.Percent()*100)), paused)
			}

			v.WindowTitle = windowTitle
			return v
		}
	}

	v := tea.NewView(docStyle.Render(m.List.View()))
	v.WindowTitle = "Pomodoro CLI"
	v.AltScreen = true
	return v
}
