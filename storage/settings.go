package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gioui.org/app"
)

// Settings holds the last-used game configuration.
type Settings struct {
	GameType      int `json:"game_type"`
	Difficulty    int `json:"difficulty"`
	MaxHintsSel   int `json:"max_hints_sel"`
	MaxErrorsSel  int `json:"max_errors_sel"`
	Lang          int `json:"lang"`
}

func settingsPath() (string, error) {
	dir, err := app.DataDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "GoSudoku")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// LoadSettings reads settings from disk; returns defaults on any error.
func LoadSettings() Settings {
	path, err := settingsPath()
	if err != nil {
		return Settings{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}
	}
	return s
}

// Save writes settings to disk.
func (s Settings) Save() {
	path, err := settingsPath()
	if err != nil {
		return
	}
	data, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(path, data, 0o644)
}
