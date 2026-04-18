package snake

import (
	"path/filepath"
	"testing"
)

func TestLoadScores_ReturnsNilOnMissing(t *testing.T) {
	got := loadScores(filepath.Join(t.TempDir(), "nope.json"))
	if got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestAddScore_SortsDescending(t *testing.T) {
	entries := addScore(nil, 30)
	entries = addScore(entries, 10)
	entries = addScore(entries, 20)
	if entries[0].Score != 30 {
		t.Errorf("first = %d, want 30", entries[0].Score)
	}
	if entries[2].Score != 10 {
		t.Errorf("third = %d, want 10", entries[2].Score)
	}
}

func TestAddScore_KeepsTopFive(t *testing.T) {
	var entries []ScoreEntry
	for i := 0; i < 7; i++ {
		entries = addScore(entries, i*10)
	}
	if len(entries) != 5 {
		t.Errorf("want 5 entries, got %d", len(entries))
	}
	if entries[0].Score != 60 {
		t.Errorf("top score = %d, want 60", entries[0].Score)
	}
}

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	want := []ScoreEntry{{Score: 42}, {Score: 17}}
	if err := saveScores(path, want); err != nil {
		t.Fatalf("saveScores: %v", err)
	}
	got := loadScores(path)
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestTopScore_ReturnsEmptyStringWhenNoScores(t *testing.T) {
	got := topScoreFromPath(filepath.Join(t.TempDir(), "nope.json"))
	if got != "—" {
		t.Errorf("want —, got %q", got)
	}
}

func TestTopScore_ReturnsHighestScore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	saveScores(path, []ScoreEntry{{Score: 99}, {Score: 42}})
	got := topScoreFromPath(path)
	if got != "99" {
		t.Errorf("want 99, got %q", got)
	}
}
