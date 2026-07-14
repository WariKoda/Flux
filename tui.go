package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// hostGroup fasst Hosts mit gleichem Server (HostName) zusammen.
type hostGroup struct {
	Server string
	Hosts  []HostEntry
}

// groupHosts gruppiert nach HostName (Fallback: Alias) und erhält die
// Reihenfolge des ersten Auftretens.
func groupHosts(hosts []HostEntry) []hostGroup {
	index := map[string]int{}
	var groups []hostGroup
	for _, h := range hosts {
		server := h.HostName
		if server == "" {
			server = h.Alias
		}
		i, ok := index[server]
		if !ok {
			i = len(groups)
			index[server] = i
			groups = append(groups, hostGroup{Server: server})
		}
		groups[i].Hosts = append(groups[i].Hosts, h)
	}
	return groups
}

const (
	modeMain = iota // Hauptansicht: verbinden
	modeEdit        // Filter-Sub-UI: Einträge ein-/ausblenden
)

const (
	titleMain = " Flux — Enter/Klick: verbinden · e: Filter · t: Theme · q/Esc: beenden "
	titleEdit = " Flux · Filter — Enter/Klick/Leertaste: umschalten · e/Esc: fertig "
)

// hostShortDetail liefert die Kurzinfo-Spalte eines Hosts (effektiver su-User).
func hostShortDetail(h HostEntry) string {
	if h.RemoteCommand == "" {
		return ""
	}
	if user, _ := parseRemoteTarget(h.RemoteCommand); user != "" {
		return "→ " + user
	}
	return ""
}

// runTUI zeigt die nach Servern gruppierte Host-Tabelle als zentriertes
// Fenster an und liefert den gewählten Alias. Ein leerer String bedeutet:
// Nutzer hat ohne Auswahl beendet (q/Esc). 't' wechselt das Theme
// (persistiert unter themePath), 'e' öffnet die Filter-Sub-UI
// (persistiert unter excludePath). Sind anfangs alle Hosts ausgeschlossen,
// startet Flux direkt in der Filter-Sub-UI.
func runTUI(entries []HostEntry, excludes []string, excludePath, themeName, themePath string) (string, error) {
	themeIdx, err := themeIndex(themeName)
	if err != nil {
		return "", err
	}
	excludeState := NewExcludeState(entries, excludes)

	app := tview.NewApplication()
	app.EnableMouse(true)

	table := tview.NewTable()
	table.SetSelectable(true, false)
	table.SetBorder(true)

	footer := tview.NewTextView()

	var (
		selected string
		saveErr  error
	)
	connect := func(alias string) {
		selected = alias
		app.Stop()
	}

	mode := modeMain
	if len(excludeState.Visible(entries)) == 0 {
		mode = modeEdit
	}

	rowHosts := map[int]HostEntry{}

	// fillTable baut den Tabelleninhalt für den aktiven Modus neu auf.
	// Vorwärtsdeklaration, weil die Klick-Callbacks toggle() brauchen und
	// toggle() wiederum fillTable.
	var fillTable func(th Theme)

	toggle := func(h HostEntry) {
		excludeState.Excluded[h.Alias] = !excludeState.Excluded[h.Alias]
		if err := SaveExcludes(excludePath, excludeState.List(entries)); err != nil {
			saveErr = err
			app.Stop()
			return
		}
		row, _ := table.GetSelection()
		fillTable(themes[themeIdx])
		table.Select(row, 0)
	}

	fillTable = func(th Theme) {
		table.Clear()
		for k := range rowHosts {
			delete(rowHosts, k)
		}
		hosts := entries
		if mode == modeMain {
			hosts = excludeState.Visible(entries)
		}
		row := 0
		for _, g := range groupHosts(hosts) {
			table.SetCell(row, 0, tview.NewTableCell("─ "+g.Server+" ").
				SetTextColor(th.Header).
				SetSelectable(false))
			table.SetCell(row, 1, tview.NewTableCell("").
				SetTextColor(th.Header).
				SetSelectable(false).
				SetExpansion(1))
			row++
			for _, h := range g.Hosts {
				host := h
				var label string
				var click func() bool
				if mode == modeEdit {
					marker := "[x]"
					if excludeState.Excluded[host.Alias] {
						marker = "[ ]"
					}
					label = " " + marker + " " + host.Alias + "  "
					click = func() bool {
						toggle(host)
						return true
					}
				} else {
					label = "  " + host.Alias + "  "
					click = func() bool {
						connect(host.Alias)
						return true
					}
				}
				table.SetCell(row, 0, tview.NewTableCell(label).
					SetTextColor(th.Text).
					SetClickedFunc(click))
				table.SetCell(row, 1, tview.NewTableCell(hostShortDetail(host)).
					SetTextColor(th.Detail).
					SetExpansion(1).
					SetClickedFunc(click))
				rowHosts[row] = host
				row++
			}
		}
	}

	footerText := func(row int) string {
		text := ""
		if h, ok := rowHosts[row]; ok {
			text = hostDetail(h)
			if mode == modeEdit && excludeState.Excluded[h.Alias] {
				text = "ausgeblendet · " + text
			}
		} else if mode == modeMain {
			text = "alle Hosts ausgeblendet — 'e' öffnet den Filter"
		}
		return fmt.Sprintf(" %s │ Theme: %s", text, themes[themeIdx].DisplayName)
	}

	applyTheme := func(th Theme) {
		table.SetBackgroundColor(th.Background)
		table.SetBorderColor(th.Border)
		table.SetTitleColor(th.Title)
		table.SetSelectedStyle(tcell.StyleDefault.
			Foreground(th.SelectedFg).
			Background(th.SelectedBg))
		footer.SetBackgroundColor(th.Background)
		footer.SetTextColor(th.Detail)
		fillTable(th)
	}

	setMode := func(m int) {
		mode = m
		if mode == modeEdit {
			table.SetTitle(titleEdit)
		} else {
			table.SetTitle(titleMain)
		}
		fillTable(themes[themeIdx])
		table.Select(1, 0)
		row, _ := table.GetSelection()
		footer.SetText(footerText(row))
	}

	applyTheme(themes[themeIdx])

	table.SetSelectionChangedFunc(func(row, column int) {
		footer.SetText(footerText(row))
	})
	table.SetSelectedFunc(func(row, column int) {
		h, ok := rowHosts[row]
		if !ok {
			return
		}
		if mode == modeEdit {
			toggle(h)
			footer.SetText(footerText(row))
		} else {
			connect(h.Alias)
		}
	})
	setMode(mode)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if mode == modeEdit {
				setMode(modeMain)
			} else {
				app.Stop()
			}
			return nil
		}
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'e':
				if mode == modeEdit {
					setMode(modeMain)
				} else {
					setMode(modeEdit)
				}
				return nil
			case ' ':
				if mode == modeEdit {
					row, _ := table.GetSelection()
					if h, ok := rowHosts[row]; ok {
						toggle(h)
						footer.SetText(footerText(row))
					}
					return nil
				}
			case 't':
				themeIdx = (themeIdx + 1) % len(themes)
				if err := SaveThemeName(themePath, themes[themeIdx].Name); err != nil {
					saveErr = err
					app.Stop()
					return nil
				}
				applyTheme(themes[themeIdx])
				row, _ := table.GetSelection()
				footer.SetText(footerText(row))
				return nil
			}
		}
		return event
	})

	// Fenstermaße aus der Filter-Ansicht ableiten (Obermenge aller Zeilen).
	// Auch die Fußzeile zählt mit, damit die Detailanzeige nie abgeschnitten
	// wird.
	longestTheme := ""
	for _, th := range themes {
		if len([]rune(th.DisplayName)) > len([]rune(longestTheme)) {
			longestTheme = th.DisplayName
		}
	}
	maxLine := len([]rune(titleMain))
	if l := len([]rune(titleEdit)); l > maxLine {
		maxLine = l
	}
	totalRows := 0
	footerMax := 0
	for _, g := range groupHosts(entries) {
		if l := len([]rune("─ " + g.Server + " ")); l > maxLine {
			maxLine = l
		}
		totalRows++
		for _, h := range g.Hosts {
			line := " [x] " + h.Alias + "  " + hostShortDetail(h)
			if l := len([]rune(line)); l > maxLine {
				maxLine = l
			}
			footerLine := " ausgeblendet · " + hostDetail(h) + " │ Theme: " + longestTheme
			if l := len([]rune(footerLine)); l > footerMax {
				footerMax = l
			}
			totalRows++
		}
	}
	width := maxLine + 4
	if footerMax+1 > width {
		width = footerMax + 1
	}
	height := totalRows + 3 // Rahmen (2) + Fußzeile (1)

	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, totalRows+2, 0, true).
		AddItem(footer, 1, 0, false)
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(content, height, 0, true).
		AddItem(nil, 0, 1, false)
	outer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(inner, width, 0, true).
		AddItem(nil, 0, 1, false)

	if err := app.SetRoot(outer, true).Run(); err != nil {
		return "", fmt.Errorf("TUI-Fehler: %w", err)
	}
	if saveErr != nil {
		return "", saveErr
	}
	return selected, nil
}

// hostDetail baut die Detailzeile eines Hosts für die Fußzeile: Login-Ziel
// (User@HostName:Port), das effektive Ziel aus einem RemoteCommand
// (su-User und Ordner) sowie weitere Aliase des Blocks.
func hostDetail(h HostEntry) string {
	var parts []string

	target := ""
	switch {
	case h.HostName != "" && h.User != "":
		target = h.User + "@" + h.HostName
	case h.HostName != "":
		target = h.HostName
	case h.User != "":
		target = h.User + "@" + h.Alias
	}
	if target != "" && h.Port != "" {
		target += ":" + h.Port
	}
	if target != "" {
		parts = append(parts, target)
	}

	if h.RemoteCommand != "" {
		user, dir := parseRemoteTarget(h.RemoteCommand)
		switch {
		case user != "" && dir != "":
			parts = append(parts, "→ "+user+" · "+dir)
		case user != "":
			parts = append(parts, "→ "+user)
		case dir != "":
			parts = append(parts, "→ "+dir)
		default:
			// Unbekanntes Kommando transparent zeigen statt still verwerfen.
			parts = append(parts, "→ "+h.RemoteCommand)
		}
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
	return strings.Join(parts, " ")
}

var (
	remoteUserRe = regexp.MustCompile(`(?:^|\s)-l\s+(\S+)`)
	remoteDirRe  = regexp.MustCompile(`\bcd\s+([^&;']+)`)
)

// parseRemoteTarget extrahiert aus einem RemoteCommand im su-Stil
// (`su … -l <user> -c 'cd <ordner> && …'`) den effektiven User und Ordner.
// Nicht erkannte Bestandteile bleiben leer; die Anzeige fällt dann auf das
// rohe Kommando zurück.
func parseRemoteTarget(cmd string) (user, dir string) {
	if m := remoteUserRe.FindStringSubmatch(cmd); m != nil {
		user = m[1]
	}
	if m := remoteDirRe.FindStringSubmatch(cmd); m != nil {
		dir = strings.TrimSpace(m[1])
	}
	return user, dir
}
