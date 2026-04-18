package tapper

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type ScoreEntry struct {
	Score int `json:"score"`
	Wave  int `json:"wave"`
}

func defaultScorePath() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, ".local", "share", "charms", "tapper_scores.json")
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

func addScore(entries []ScoreEntry, score, wave int) []ScoreEntry {
	entries = append(entries, ScoreEntry{Score: score, Wave: wave})
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})
	if len(entries) > 5 {
		entries = entries[:5]
	}
	return entries
}
