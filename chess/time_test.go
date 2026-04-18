package chess

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTimeSelect_PressOneSetsBullet(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	updated, _ := m.Update(key("1"))
	got := updated.(model)
	if got.whiteTime != 1*time.Minute {
		t.Errorf("bullet: whiteTime = %v, want 1m", got.whiteTime)
	}
	if got.blackTime != 1*time.Minute {
		t.Errorf("bullet: blackTime = %v, want 1m", got.blackTime)
	}
}

func TestTimeSelect_PressTwoSetsBlitz(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	updated, _ := m.Update(key("2"))
	got := updated.(model)
	if got.whiteTime != 5*time.Minute {
		t.Errorf("blitz: whiteTime = %v, want 5m", got.whiteTime)
	}
}

func TestTimeSelect_PressThreeSetsRapid(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	updated, _ := m.Update(key("3"))
	got := updated.(model)
	if got.whiteTime != 10*time.Minute {
		t.Errorf("rapid: whiteTime = %v, want 10m", got.whiteTime)
	}
}

func TestTimeSelect_PressFourSetsClassical(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	updated, _ := m.Update(key("4"))
	got := updated.(model)
	if got.whiteTime != 30*time.Minute {
		t.Errorf("classical: whiteTime = %v, want 30m", got.whiteTime)
	}
}

func TestTimeSelect_LeavesTimeSelectAfterChoice(t *testing.T) {
	for _, k := range []string{"1", "2", "3", "4"} {
		m := newModel()
		m.timeSelect = true
		updated, _ := m.Update(key(k))
		if updated.(model).timeSelect {
			t.Errorf("key %q: timeSelect should be false after selection", k)
		}
	}
}

func TestTimeSelect_TwoPlayerRoutesToGame(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	updated, _ := m.Update(key("2"))
	got := updated.(model)
	if got.diffSelect {
		t.Error("two-player path should not go to diffSelect")
	}
	if got.colorSelect {
		t.Error("two-player path should not go to colorSelect")
	}
	if got.message != "White's turn" {
		t.Errorf("message = %q, want \"White's turn\"", got.message)
	}
}

func TestTimeSelect_VsComputerRoutesToDiffSelect(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	m.pendingVsComputer = true
	updated, _ := m.Update(key("2"))
	got := updated.(model)
	if !got.diffSelect {
		t.Error("vs-computer path should go to diffSelect")
	}
	if got.pendingVsComputer {
		t.Error("pendingVsComputer should be cleared after routing")
	}
}

func TestTimeSelect_OtherKeyIgnored(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	updated, _ := m.Update(key("x"))
	if !updated.(model).timeSelect {
		t.Error("unrecognised key should leave timeSelect=true")
	}
}

func TestTimeSelect_QuitWorks(t *testing.T) {
	for _, k := range []string{"q", "ctrl+c"} {
		t.Run(k, func(t *testing.T) {
			m := newModel()
			m.timeSelect = true
			_, cmd := m.Update(key(k))
			if cmd == nil {
				t.Fatalf("key %q should return quit command", k)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Errorf("expected QuitMsg for key %q", k)
			}
		})
	}
}

func TestView_TimeSelectShowsOptions(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	v := m.View()
	for _, want := range []string{"Bullet", "Blitz", "Rapid", "Classical"} {
		if !strings.Contains(v, want) {
			t.Errorf("time select view missing %q", want)
		}
	}
	if strings.Contains(v, "a  b  c") {
		t.Error("time select view should not render the board")
	}
}

func TestModeSelect_PressOneSetsTimeSelect(t *testing.T) {
	m := newModel()
	m.modeSelect = true
	updated, _ := m.Update(key("1"))
	got := updated.(model)
	if got.modeSelect {
		t.Error("modeSelect should be false after pressing 1")
	}
	if !got.timeSelect {
		t.Error("pressing 1 should go to timeSelect")
	}
	if got.vsComputer {
		t.Error("vsComputer should still be false")
	}
}

func TestModeSelect_PressTwoSetsTimeSelect(t *testing.T) {
	m := newModel()
	m.modeSelect = true
	updated, _ := m.Update(key("2"))
	got := updated.(model)
	if got.modeSelect {
		t.Error("modeSelect should be false after pressing 2")
	}
	if !got.timeSelect {
		t.Error("pressing 2 should go to timeSelect")
	}
	if got.diffSelect {
		t.Error("should not jump straight to diffSelect — time select comes first")
	}
	if !got.pendingVsComputer {
		t.Error("pendingVsComputer should be set so we remember to go to diffSelect after time")
	}
}
