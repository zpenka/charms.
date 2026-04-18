package game2048

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ScoreEntry struct {
	Score int `json:"score"`
}

func defaultScorePath() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, ".local", "share", "charms", "2048_scores.json")
}

func loadScores(path string) []ScoreEntry {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []ScoreEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

func saveScores(path string, entries []ScoreEntry) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func addScore(entries []ScoreEntry, score int) []ScoreEntry {
	entries = append(entries, ScoreEntry{Score: score})
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})
	if len(entries) > 5 {
		entries = entries[:5]
	}
	return entries
}

func topScoreFromPath(path string) string {
	entries := loadScores(path)
	if len(entries) == 0 {
		return "—"
	}
	return fmt.Sprintf("%d", entries[0].Score)
}

// TopScore returns the all-time best 2048 score as a display string.
func TopScore() string {
	return topScoreFromPath(defaultScorePath())
}
