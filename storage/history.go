package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gioui.org/app"
)

// GameResult records the outcome of one game session.
type GameResult struct {
	GameType   string    `json:"game_type"`
	Difficulty string    `json:"difficulty"`
	StartTime  time.Time `json:"start_time"`
	DurationS  int       `json:"duration_seconds"`
	Won        bool      `json:"won"`
	GaveUp     bool      `json:"gave_up"`
	HintsUsed  int       `json:"hints_used"`
	ErrorsMade int       `json:"errors_made"`
	Points     int       `json:"points"`
}

// ComputePoints calculates a score for a won game.
//
// Formula: 1_000_000 / (seconds × 1.5^hints × 1.2^errors), minimum 1.
//
// Time is the primary driver (denominator). Hints and errors multiply the
// effective time, so their penalty scales with how long the game took rather
// than being a flat additive offset.
// Returns 0 for unfinished or lost games.
func ComputePoints(won bool, durationS, hints, errors int) int {
	if !won {
		return 0
	}
	secs := float64(max(durationS, 1))
	// Each hint multiplies effective time by 1.5; each error by 1.2.
	for range hints {
		secs *= 1.5
	}
	for range errors {
		secs *= 1.2
	}
	score := int(1_000_000 / secs)
	if score < 1 {
		return 1
	}
	return score
}

// History is the persisted collection of all game results.
type History struct {
	Results []GameResult `json:"results"`
}

// Stats aggregates history data for display.
type Stats struct {
	GamesPlayed int
	GamesWon    int
	GaveUp      int
	TotalHints  int
	MinTimeSec  int // minimum winning time
	AvgTimeSec  int // average winning time
	AvgPoints   int // average score (won games only)
	MaxPoints   int // highest score ever
}

func historyPath() (string, error) {
	// app.DataDir() works on all platforms: desktop uses UserConfigDir,
	// Android uses the app's private files directory.
	dir, err := app.DataDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "GoSudoku")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

// Load reads the history from disk; returns empty history on error.
func Load() *History {
	path, err := historyPath()
	if err != nil {
		return &History{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &History{}
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return &History{}
	}
	return &h
}

// Save writes the history to disk.
func (h *History) Save() error {
	path, err := historyPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Append adds a result and saves.
func (h *History) Append(r GameResult) {
	h.Results = append(h.Results, r)
	_ = h.Save()
}

// ComputeStats returns aggregated statistics over all results.
func (h *History) ComputeStats() Stats {
	s := Stats{MinTimeSec: -1}
	totalWinTime := 0
	totalPoints := 0
	wins := 0
	for _, r := range h.Results {
		s.GamesPlayed++
		s.TotalHints += r.HintsUsed
		if r.GaveUp {
			s.GaveUp++
		}
		if r.Won {
			s.GamesWon++
			wins++
			totalWinTime += r.DurationS
			if s.MinTimeSec < 0 || r.DurationS < s.MinTimeSec {
				s.MinTimeSec = r.DurationS
			}
			pts := r.Points
			if pts == 0 {
				// recalculate for old records that predate points field
				pts = ComputePoints(true, r.DurationS, r.HintsUsed, r.ErrorsMade)
			}
			totalPoints += pts
			if pts > s.MaxPoints {
				s.MaxPoints = pts
			}
		}
	}
	if wins > 0 {
		s.AvgTimeSec = totalWinTime / wins
		s.AvgPoints = totalPoints / wins
	}
	if s.MinTimeSec < 0 {
		s.MinTimeSec = 0
	}
	return s
}

// FormatDuration converts seconds to "mm:ss".
func FormatDuration(sec int) string {
	m := sec / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
