package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// runTUI zeigt die Host-Liste an und liefert den gewählten Alias.
// Ein leerer String bedeutet: Nutzer hat ohne Auswahl beendet (q/Esc).
func runTUI(hosts []HostEntry) (string, error) {
	app := tview.NewApplication()
	app.EnableMouse(true)

	list := tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true)
	list.SetBorder(true)
	list.SetTitle(" Flux — Enter/Klick: verbinden · q/Esc: beenden ")

	var selected string
	for _, h := range hosts {
		alias := h.Alias
		list.AddItem(alias, hostDetail(h), 0, func() {
			selected = alias
			app.Stop()
		})
	}
	// Esc beendet ohne Auswahl.
	list.SetDoneFunc(app.Stop)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			}
		}
		return event
	})

	if err := app.SetRoot(list, true).Run(); err != nil {
		return "", fmt.Errorf("TUI-Fehler: %w", err)
	}
	return selected, nil
}

// hostDetail baut die Zweitzeile eines Listeneintrags: Ziel (User@HostName)
// und weitere Aliase des Blocks.
func hostDetail(h HostEntry) string {
	var parts []string
	switch {
	case h.HostName != "" && h.User != "":
		parts = append(parts, h.User+"@"+h.HostName)
	case h.HostName != "":
		parts = append(parts, h.HostName)
	case h.User != "":
		parts = append(parts, h.User+"@"+h.Alias)
	}
	if len(h.Aliases) > 1 {
		var others []string
		for _, a := range h.Aliases {
			if a != h.Alias {
				others = append(others, a)
			}
		}
		parts = append(parts, "auch: "+strings.Join(others, ", "))
	}
	return strings.Join(parts, " · ")
}
