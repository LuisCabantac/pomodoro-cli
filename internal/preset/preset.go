package preset

import (
	"fmt"

	"charm.land/bubbles/v2/list"
)

type PresetList struct {
	Presets []Preset `json:"presets"`
}

type Preset struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	WorkMin              int    `json:"work_min"`
	ShortBreakMin        int    `json:"short_break_min"`
	LongBreakMin         int    `json:"long_break_min"`
	CycleBeforeLongBreak int    `json:"cycles_before_long_break"`
}

type PresetsLoadedMsg struct {
	Presets []Preset
	Err     error
}

func (p Preset) Title() string { return p.Name }
func (p Preset) Description() string {
	return fmt.Sprintf("Work: %dm, Short Break: %dm, Long Break: %dm", p.WorkMin, p.ShortBreakMin, p.LongBreakMin)
}
func (p Preset) FilterValue() string { return p.Name }

func InitialPresets() []Preset {
	return []Preset{
		{ID: "classic", Name: "Classic", WorkMin: 25, ShortBreakMin: 5, LongBreakMin: 10, CycleBeforeLongBreak: 4},
		{ID: "short", Name: "Short", WorkMin: 15, ShortBreakMin: 3, LongBreakMin: 10, CycleBeforeLongBreak: 4},
	}
}

func PresetsToItems(presets []Preset) []list.Item {
	items := make([]list.Item, len(presets))
	for i, p := range presets {
		items[i] = p
	}
	return items
}
