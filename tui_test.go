package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestHelpTextListsCommandsAndOptions(t *testing.T) {
	got := helpText()
	for _, want := range []string{"Ctrl+E", "Ctrl+T", "Ctrl+B", "Ctrl+A", "Ctrl+H", "Esc", "Wortmarke · ANSI", "Terminal · Monochrom", "Links", "Mitte", "Rechts", "Okabe-Ito Dunkel"} {
		if !strings.Contains(got, want) {
			t.Errorf("Hilfe enthält %q nicht", want)
		}
	}
}

func TestHelpInputRules(t *testing.T) {
	state := tuiViewState{}
	if !handleHelpKey(tcell.NewEventKey(tcell.KeyCtrlH, 0, 0), &state) || !state.HelpVisible {
		t.Fatal("Ctrl+H muss Hilfe öffnen")
	}
	if !handleHelpKey(tcell.NewEventKey(tcell.KeyRune, 'x', 0), &state) || !state.HelpVisible {
		t.Fatal("Eingabe muss konsumiert werden")
	}
	if !handleHelpKey(tcell.NewEventKey(tcell.KeyEscape, 0, 0), &state) || state.HelpVisible {
		t.Fatal("Esc muss Hilfe schließen")
	}
}
