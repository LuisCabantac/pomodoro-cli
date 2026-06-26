package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/LuisCabantac/pomodoro-cli/internal/preset"
	"github.com/LuisCabantac/pomodoro-cli/internal/stats"
)

type Items interface {
	preset.Preset | stats.Stat
}

type PresetList preset.PresetList
type StatList stats.StatList

type ItemsList interface {
	isItemList()
}

func (PresetList) isItemList() {}
func (StatList) isItemList()   {}

func WriteItems[T Items](fileName string, items []T) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	mainDir := filepath.Join(configDir, "pomodoro-cli")
	filePath := filepath.Join(mainDir, fileName)

	err = os.MkdirAll(mainDir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var itemsList ItemsList

	switch v := any(items).(type) {
	case []preset.Preset:
		itemsList = PresetList{
			Presets: v,
		}
	case []stats.Stat:
		itemsList = StatList{
			Stats: v,
		}
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err = encoder.Encode(itemsList); err != nil {
		return err
	}

	return nil
}

func LoadItems(fileName string) (ItemsList, error) {
	defaultPresets := preset.InitialPresets()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	mainDir := filepath.Join(configDir, "pomodoro-cli")
	filePath := filepath.Join(mainDir, fileName)

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

		var itemsList ItemsList

		if fileName == "presets.json" {
			itemsList = PresetList(preset.PresetList{
				Presets: defaultPresets,
			})
		}
		if fileName == "stats.json" {
			itemsList = StatList(stats.StatList{
				Stats: []stats.Stat{},
			})
		}

		encoder := json.NewEncoder(newFile)
		encoder.SetIndent("", "    ")

		if err = encoder.Encode(itemsList); err != nil {
			return itemsList, err
		}

		return itemsList, nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	if fileName == "presets.json" {
		var pList PresetList
		if err := decoder.Decode(&pList); err != nil {
			return nil, err
		}
		return pList, nil
	}

	if fileName == "stats.json" {
		var sList StatList
		if err := decoder.Decode(&sList); err != nil {
			return nil, err
		}
		return sList, nil
	}

	return nil, fmt.Errorf("unknown file name: %s", fileName)
}

func LoadItemsCmd() tea.Cmd {
	return func() tea.Msg {
		loadedItems, err := LoadItems("presets.json")
		if v, ok := loadedItems.(PresetList); ok {
			return preset.PresetsLoadedMsg{Presets: v.Presets, Err: err}
		}
		return preset.PresetsLoadedMsg{Presets: []preset.Preset{}, Err: errors.New("failed to load presets")}
	}
}
