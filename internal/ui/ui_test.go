package ui

import (
	"strings"
	"testing"

	"github.com/TechnoRed2026/redscope/internal/netmon"
	"github.com/gdamore/tcell/v2"
)

func TestPad(t *testing.T) {
	checks := map[string]string{
		pad("abcdefghij", 5): "abcd…",
		pad("ab", 5):         "ab   ",
		pad("", 4):           "-   ",
		pad("exact", 5):      "exact",
	}
	for got, want := range checks {
		if got != want {
			t.Errorf("pad → %q, want %q", got, want)
		}
	}
}

func TestPaletteUsesReadableRGB(t *testing.T) {
	for name, c := range map[string]tcell.Color{"bg": cBg, "panel": cPanel, "text": cText, "brand": cBrand} {
		if !c.IsRGB() {
			t.Fatalf("%s color must be explicit RGB, got %v", name, c)
		}
	}
	if cText.Hex() == 0 || cBg.Hex() == 0 {
		t.Fatal("palette must not use black foreground/background")
	}
}

func TestStateColor(t *testing.T) {
	cases := map[string]tcell.Color{
		"ESTABLISHED": cGood,
		"established": cGood,
		"LISTEN":      cSignal,
		"listen":      cSignal,
		"TIME_WAIT":   cMuted, // gopsutil uses underscore form
		"time_wait":   cMuted,
		"close_wait":  cMuted,
		"closed":      cMuted,
		"":            cText,
		"weird":       cText,
	}
	for in, want := range cases {
		if got := stateColor(in); got != want {
			t.Errorf("stateColor(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestStateLabel(t *testing.T) {
	if got := stateLabel("time_wait"); got != "TIME-WAIT" {
		t.Errorf("label(time_wait) = %q", got)
	}
	if got := stateLabel(""); got != "-" {
		t.Errorf("label('') = %q", got)
	}
}

func TestSearchText(t *testing.T) {
	e := netmon.Entry{Process: "curl", PID: 99, Protocol: "TCP",
		Local: "1.2.3.4:5", RemoteIP: "9.9.9.9", Host: "host.x", State: "established"}
	s := searchText(e)
	for _, want := range []string{"curl", "99", "TCP", "1.2.3.4:5", "9.9.9.9", "host.x", "established"} {
		if !strings.Contains(s, want) {
			t.Errorf("searchText missing %q in %q", want, s)
		}
	}
}
