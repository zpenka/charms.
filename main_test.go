package main

import (
	"strings"
	"testing"
)

func TestGames_SnakeIsPresent(t *testing.T) {
	found := false
	for _, g := range games {
		if g.name == "Snake" {
			found = true
		}
	}
	if !found {
		t.Error("Snake should be in the games list")
	}
}

func TestGames_2048IsPresent(t *testing.T) {
	found := false
	for _, g := range games {
		if g.name == "2048" {
			found = true
		}
	}
	if !found {
		t.Error("2048 should be in the games list")
	}
}

func TestGames_AllHaveDescriptions(t *testing.T) {
	for _, g := range games {
		if g.desc == "" {
			t.Errorf("game %q should have a description", g.name)
		}
	}
}

func TestLobbyView_ShowsAllGameNames(t *testing.T) {
	m := newLobbyModel()
	v := m.View()
	for _, g := range games {
		if !strings.Contains(v, g.name) {
			t.Errorf("lobby view should show game name %q", g.name)
		}
	}
}

func TestLobbyView_ShowsDescriptions(t *testing.T) {
	m := newLobbyModel()
	v := m.View()
	for _, g := range games {
		// check at least the first word of each description appears
		firstWord := strings.Fields(g.desc)[0]
		if !strings.Contains(v, firstWord) {
			t.Errorf("lobby view should show description for %q (looking for %q)", g.name, firstWord)
		}
	}
}
