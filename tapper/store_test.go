package tapper

import (
	"path/filepath"
	"testing"
)

func TestLoadScores_ReturnsEmptyOnMissing(t *testing.T) {
	entries := loadScores(filepath.Join(t.TempDir(), "nope.json"))
	if entries != nil {
		t.Errorf("want nil, got %v", entries)
	}
}

func TestAddScore_InsertsAndSortsDescending(t *testing.T) {
	entries := addScore(nil, 50, 3)
	entries = addScore(entries, 100, 5)
	entries = addScore(entries, 75, 4)
	if entries[0].Score != 100 {
		t.Errorf("first entry score = %d, want 100", entries[0].Score)
	}
	if entries[1].Score != 75 {
		t.Errorf("second entry score = %d, want 75", entries[1].Score)
	}
	if entries[2].Score != 50 {
		t.Errorf("third entry score = %d, want 50", entries[2].Score)
	}
}

func TestAddScore_KeepsTopFive(t *testing.T) {
	var entries []ScoreEntry
	for i := 0; i < 7; i++ {
		entries = addScore(entries, i*10, i+1)
	}
	if len(entries) != 5 {
		t.Errorf("want 5 entries, got %d", len(entries))
	}
	if entries[0].Score != 60 {
		t.Errorf("top score = %d, want 60", entries[0].Score)
	}
}

func TestAddScore_RecordsWave(t *testing.T) {
	entries := addScore(nil, 42, 7)
	if entries[0].Wave != 7 {
		t.Errorf("wave = %d, want 7", entries[0].Wave)
	}
}

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	want := []ScoreEntry{{Score: 100, Wave: 3}, {Score: 50, Wave: 1}}
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
