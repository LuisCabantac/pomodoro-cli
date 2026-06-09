package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/LuisCabantac/pomodoro-cli/internal/preset"
)

func WriteItems(presets []preset.Preset) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	mainDir := filepath.Join(configDir, "pomodoro-cli")
	filePath := filepath.Join(mainDir, "presets.json")

	err = os.MkdirAll(mainDir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	presetsList := preset.PresetList{
		Presets: presets,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err = encoder.Encode(presetsList); err != nil {
		return err
	}

	return nil
}

func LoadItems() ([]preset.Preset, error) {
	defaultPresets := preset.InitialPresets()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	mainDir := filepath.Join(configDir, "pomodoro-cli")
	filePath := filepath.Join(mainDir, "presets.json")

	err = os.MkdirAll(mainDir, 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		newFile, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer newFile.Close()

		presetsList := preset.PresetList{
			Presets: defaultPresets,
		}

		encoder := json.NewEncoder(newFile)
		encoder.SetIndent("", "    ")

		if err = encoder.Encode(presetsList); err != nil {
			return defaultPresets, err
		}

		return defaultPresets, nil
	}
	defer file.Close()

	var presetsList preset.PresetList
	err = json.NewDecoder(file).Decode(&presetsList)
	if err != nil {
		return nil, err
	}

	return presetsList.Presets, nil
}

func LoadItemsCmd() tea.Cmd {
	return func() tea.Msg {
		presets, err := LoadItems()
		return preset.PresetsLoadedMsg{Presets: presets, Err: err}
	}
}
