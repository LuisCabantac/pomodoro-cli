package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/LuisCabantac/pomodoro-cli/internal/preset"
)

func LoadItemsCmd() tea.Cmd {
	return func() tea.Msg {
		var presets []preset.Preset
		defaultPresets := preset.InitialPresets()

		configDir, err := os.UserConfigDir()
		if err != nil {
			return preset.PresetsLoadedMsg{Presets: presets, Err: err}
		}

		mainDir := filepath.Join(configDir, "pomodoro")
		filePath := filepath.Join(configDir, "pomodoro", "presets.json")

		err = os.MkdirAll(mainDir, 0755)
		if err != nil {
			return preset.PresetsLoadedMsg{Presets: presets, Err: err}
		}

		file, err := os.Open(filePath)
		if err != nil {
			newFile, err := os.Create(filePath)
			if err != nil {
				return preset.PresetsLoadedMsg{Presets: presets, Err: err}
			}
			defer newFile.Close()

			presetsList := preset.PresetList{
				Presets: defaultPresets,
			}

			encoder := json.NewEncoder(newFile)
			encoder.SetIndent("", "    ")

			if err = encoder.Encode(presetsList); err != nil {
				return preset.PresetsLoadedMsg{Presets: defaultPresets, Err: err}
			}

			return preset.PresetsLoadedMsg{Presets: defaultPresets, Err: err}
		}
		defer file.Close()

		var presetsList preset.PresetList
		err = json.NewDecoder(file).Decode(&presetsList)
		if err != nil {
			return preset.PresetsLoadedMsg{Presets: presets, Err: err}
		}

		presets = presetsList.Presets

		return preset.PresetsLoadedMsg{Presets: presets, Err: nil}
	}
}
