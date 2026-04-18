package chess

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestScheme_DefaultIsClassic(t *testing.T) {
	m := newModel()
	if m.schemeIdx != 0 {
		t.Errorf("default schemeIdx = %d, want 0 (Classic)", m.schemeIdx)
	}
}

func TestScheme_FourSchemesAvailable(t *testing.T) {
	if len(schemes) != 4 {
		t.Errorf("want 4 schemes, got %d", len(schemes))
	}
}

func TestSchemeSelect_PressOneSelectsFirst(t *testing.T) {
	m := newModel()
	m.schemeSelect = true
	updated, _ := m.Update(key("1"))
	if updated.(model).schemeIdx != 0 {
		t.Errorf("key 1: schemeIdx = %d, want 0", updated.(model).schemeIdx)
	}
}

func TestSchemeSelect_PressFourSelectsFourth(t *testing.T) {
	m := newModel()
	m.schemeSelect = true
	updated, _ := m.Update(key("4"))
	if updated.(model).schemeIdx != 3 {
		t.Errorf("key 4: schemeIdx = %d, want 3", updated.(model).schemeIdx)
	}
}

func TestSchemeSelect_LeavesSchemeSelectAfterChoice(t *testing.T) {
	for _, k := range []string{"1", "2", "3", "4"} {
		m := newModel()
		m.schemeSelect = true
		updated, _ := m.Update(key(k))
		if updated.(model).schemeSelect {
			t.Errorf("key %q: schemeSelect should be false after selection", k)
		}
	}
}

func TestSchemeSelect_TwoPlayerRoutesToGame(t *testing.T) {
	m := newModel()
	m.schemeSelect = true
	updated, _ := m.Update(key("1"))
	got := updated.(model)
	if got.diffSelect {
		t.Error("two-player path should not go to diffSelect")
	}
	if got.message != "White's turn" {
		t.Errorf("message = %q, want \"White's turn\"", got.message)
	}
}

func TestSchemeSelect_VsComputerRoutesToDiffSelect(t *testing.T) {
	m := newModel()
	m.schemeSelect = true
	m.pendingVsComputer = true
	updated, _ := m.Update(key("1"))
	got := updated.(model)
	if !got.diffSelect {
		t.Error("vs-computer path should go to diffSelect after scheme")
	}
	if got.pendingVsComputer {
		t.Error("pendingVsComputer should be cleared")
	}
}

func TestSchemeSelect_OtherKeyIgnored(t *testing.T) {
	m := newModel()
	m.schemeSelect = true
	updated, _ := m.Update(key("x"))
	if !updated.(model).schemeSelect {
		t.Error("unrecognised key should leave schemeSelect=true")
	}
}

func TestSchemeSelect_QuitWorks(t *testing.T) {
	for _, k := range []string{"q", "ctrl+c"} {
		t.Run(k, func(t *testing.T) {
			m := newModel()
			m.schemeSelect = true
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

func TestView_SchemeSelectShowsOptions(t *testing.T) {
	m := newModel()
	m.schemeSelect = true
	v := m.View()
	for _, want := range []string{"Classic", "Ocean", "Mint", "Dusk"} {
		if !strings.Contains(v, want) {
			t.Errorf("scheme select view missing %q", want)
		}
	}
	if strings.Contains(v, "a  b  c") {
		t.Error("scheme select view should not render the board")
	}
}

func TestTimeSelect_RoutesToSchemeSelect(t *testing.T) {
	m := newModel()
	m.timeSelect = true
	updated, _ := m.Update(key("2"))
	got := updated.(model)
	if !got.schemeSelect {
		t.Error("timeSelect should route to schemeSelect after a time choice")
	}
	if got.timeSelect {
		t.Error("timeSelect should be false after choice")
	}
}
