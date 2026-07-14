package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestMainTitlesAdvertiseHelp(t *testing.T) {
	if !strings.Contains(titleMain, "^O: Optionen") || !strings.Contains(titleEdit, "^O: Optionen") {
		t.Fatal("Optionen-Shortcut fehlt im Titel")
	}
}

func TestFooterSettingsText(t *testing.T) {
	got := settingsStatus(themes[0], banners[0], bannerAlignments[2])
	for _, want := range []string{"Theme: Dunkel", "Banner: ANSI", "Ausrichtung: Rechts"} {
		if !strings.Contains(got, want) {
			t.Errorf("Status enthält %q nicht: %q", want, got)
		}
	}
	for _, old := range []string{"Wortmarke", "Terminal"} {
		if strings.Contains(got, old) {
			t.Errorf("Status enthält alte Banner-Auswahl %q: %q", old, got)
		}
	}
}

func TestNaturalTableWidthDependsOnlyOnGroupAndHostRows(t *testing.T) {
	entries := []HostEntry{{
		Alias:         "kurz",
		Aliases:       []string{"kurz", "zweiter-alias"},
		HostName:      "srv",
		User:          "benutzer-mit-sehr-langem-footer-detail",
		Port:          "2222",
		RemoteCommand: "su -l zielbenutzer -c 'cd /ein/sehr/langer/footer/pfad && exec bash'",
	}}

	got := naturalTableWidth(entries)
	want := len([]rune(" [x] kurz  → zielbenutzer")) + 4
	if got != want {
		t.Fatalf("natürliche Tabellenbreite = %d, erwartet %d", got, want)
	}
	if got >= len([]rune(titleMain)) || got >= len([]rune(hostDetail(entries[0]))) {
		t.Fatalf("Breite %d wurde von Titel oder Footer beeinflusst", got)
	}
}

func TestHelpTextListsCommandsAndOptions(t *testing.T) {
	got := helpText()
	for _, want := range []string{
		"Tippen", "Suche", "Backspace", "Pfeil ↑/↓", "Home/End", "Enter",
		"Linksklick", "Mausrad", "Ctrl+E", "Ctrl+T", "Ctrl+B", "Ctrl+A",
		"Ctrl+O", "Esc", "ANSI", "Monochrom",
		"Links", "Mitte", "Rechts", "Okabe-Ito Dunkel",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Hilfe enthält %q nicht", want)
		}
	}
	for _, old := range []string{"Wortmarke", "Terminal"} {
		if strings.Contains(got, old) {
			t.Errorf("Hilfe enthält alte Banner-Auswahl %q", old)
		}
	}
}

func TestHelpInputRules(t *testing.T) {
	state := tuiViewState{}
	for _, key := range []tcell.Key{tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyEscape} {
		if handleHelpKey(tcell.NewEventKey(key, 0, 0), &state) {
			t.Fatalf("%v darf bei geschlossenen Optionen nicht konsumiert werden", key)
		}
	}
	if handleHelpKey(tcell.NewEventKey(tcell.KeyRune, 'x', 0), &state) {
		t.Fatal("Unbeteiligte Eingabe darf bei geschlossenen Optionen nicht konsumiert werden")
	}
	if !handleHelpKey(tcell.NewEventKey(tcell.KeyCtrlO, 0, 0), &state) || !state.HelpVisible {
		t.Fatal("Ctrl+O muss Optionen öffnen")
	}
	if !handleHelpKey(tcell.NewEventKey(tcell.KeyRune, 'x', 0), &state) || !state.HelpVisible {
		t.Fatal("Eingabe muss konsumiert werden")
	}
	if !handleHelpKey(tcell.NewEventKey(tcell.KeyCtrlO, 0, 0), &state) || state.HelpVisible {
		t.Fatal("Ctrl+O muss Optionen schließen")
	}
	state.HelpVisible = true
	if !handleHelpKey(tcell.NewEventKey(tcell.KeyEscape, 0, 0), &state) || state.HelpVisible {
		t.Fatal("Esc muss Optionen schließen")
	}
}
