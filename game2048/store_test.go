package game2048

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
	entries := addScore(nil, 100)
	entries = addScore(entries, 300)
	entries = addScore(entries, 200)
	if entries[0].Score != 300 {
		t.Errorf("first = %d, want 300", entries[0].Score)
	}
}

func TestAddScore_KeepsTopFive(t *testing.T) {
	var entries []ScoreEntry
	for i := 0; i < 7; i++ {
		entries = addScore(entries, i*100)
	}
	if len(entries) != 5 {
		t.Errorf("want 5 entries, got %d", len(entries))
	}
}

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	want := []ScoreEntry{{Score: 512}, {Score: 256}}
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

func TestTopScore_ReturnsEmDashWhenEmpty(t *testing.T) {
	got := topScoreFromPath(filepath.Join(t.TempDir(), "nope.json"))
	if got != "—" {
		t.Errorf("want —, got %q", got)
	}
}

func TestTopScore_ReturnsHighestScore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	saveScores(path, []ScoreEntry{{Score: 1024}, {Score: 256}})
	got := topScoreFromPath(path)
	if got != "1024" {
		t.Errorf("want 1024, got %q", got)
	}
}
