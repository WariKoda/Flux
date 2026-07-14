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
	titleMain = " Flux — Tippen: suchen · Enter/Klick: verbinden · ^E: Filter · ^T: Theme · ^H: Hilfe · Esc: beenden "
	titleEdit = " Flux · Filter — Enter/Klick/Leertaste: umschalten · ^E/Esc: fertig · ^H: Hilfe "
	titleHelp = " Flux · Hilfe — ^H/Esc: zurück "
)

type tuiViewState struct {
	HelpVisible bool
}

func helpText() string {
	bannerNames := make([]string, len(banners))
	for i, banner := range banners {
		bannerNames[i] = banner.DisplayName
	}
	alignmentNames := make([]string, len(bannerAlignments))
	for i, alignment := range bannerAlignments {
		alignmentNames[i] = alignment.DisplayName
	}
	themeNames := make([]string, len(themes))
	for i, theme := range themes {
		themeNames[i] = theme.DisplayName
	}

	return strings.Join([]string{
		"Navigation",
		"Enter/Klick  Verbinden",
		"Esc          Zurück / Beenden",
		"",
		"Ansichten",
		"Ctrl+E  Filter",
		"Ctrl+H  Hilfe",
		"",
		"Banner",
		"Ctrl+B  " + strings.Join(bannerNames, " · "),
		"",
		"Ausrichtung",
		"Ctrl+A  " + strings.Join(alignmentNames, " · "),
		"",
		"Themes",
		"Ctrl+T  " + strings.Join(themeNames, " · "),
	}, "\n")
}

func handleHelpKey(event *tcell.EventKey, state *tuiViewState) bool {
	if event.Key() == tcell.KeyCtrlH {
		state.HelpVisible = !state.HelpVisible
		return true
	}
	if !state.HelpVisible {
		return false
	}
	if event.Key() == tcell.KeyEscape {
		state.HelpVisible = false
	}
	return true
}

func settingsStatus(theme Theme, banner Banner, alignment BannerAlignment) string {
	return fmt.Sprintf("Theme: %s · Banner: %s · Ausrichtung: %s", theme.DisplayName, banner.DisplayName, alignment.DisplayName)
}

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

// matchHosts liefert die Hosts, bei denen die Suchanfrage in Alias,
// HostName, User oder im effektiven su-User/Ordner vorkommt
// (Groß-/Kleinschreibung egal). Leere Anfrage: alle Hosts.
func matchHosts(hosts []HostEntry, query string) []HostEntry {
	if query == "" {
		return hosts
	}
	q := strings.ToLower(query)
	var out []HostEntry
	for _, h := range hosts {
		if hostMatches(h, q) {
			out = append(out, h)
		}
	}
	return out
}

func hostMatches(h HostEntry, lowerQuery string) bool {
	fields := append([]string{}, h.Aliases...)
	fields = append(fields, h.HostName, h.User)
	if h.RemoteCommand != "" {
		user, dir := parseRemoteTarget(h.RemoteCommand)
		fields = append(fields, user, dir)
	}
	for _, f := range fields {
		if f != "" && strings.Contains(strings.ToLower(f), lowerQuery) {
			return true
		}
	}
	return false
}

// runTUI zeigt die nach Servern gruppierte Host-Tabelle als zentriertes
// Fenster an und liefert den gewählten Alias. Ein leerer String bedeutet:
// Nutzer hat ohne Auswahl beendet (Esc). Tippen filtert sofort; '^T'
// wechselt das Theme (persistiert unter themePath), '^B' den Banner und '^A'
// dessen Ausrichtung. '^H' öffnet die Hilfe, '^E' die Filter-Sub-UI
// (persistiert unter excludePath). Sind anfangs alle Hosts ausgeschlossen,
// startet Flux direkt in der Filter-Sub-UI.
func runTUI(entries []HostEntry, excludes []string, excludePath, themeName, themePath, bannerName, bannerPath, alignmentName, alignmentPath string) (string, error) {
	themeIdx, err := themeIndex(themeName)
	if err != nil {
		return "", err
	}
	bannerIdx, err := bannerIndex(bannerName)
	if err != nil {
		return "", err
	}
	alignmentIdx, err := bannerAlignmentIndex(alignmentName)
	if err != nil {
		return "", err
	}
	excludeState := NewExcludeState(entries, excludes)

	app := tview.NewApplication()
	app.EnableMouse(true)

	table := tview.NewTable()
	table.SetSelectable(true, false)
	table.SetBorder(true)

	searchBar := tview.NewTextView()
	footer := tview.NewTextView()
	bannerView := tview.NewTextView().SetDynamicColors(true)
	helpView := tview.NewTextView().SetScrollable(true)
	helpView.SetBorder(true).SetTitle(titleHelp)
	helpView.SetText(helpText())
	bodyPages := tview.NewPages().
		AddPage("hosts", table, true, true).
		AddPage("help", helpView, true, false)
	viewState := tuiViewState{}

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
	query := ""

	rowHosts := map[int]HostEntry{}
	firstHostRow := -1

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
		firstHostRow = -1
		hosts := entries
		if mode == modeMain {
			hosts = excludeState.Visible(entries)
		}
		hosts = matchHosts(hosts, query)
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
				if firstHostRow < 0 {
					firstHostRow = row
				}
				rowHosts[row] = host
				row++
			}
		}
	}

	updateSearchBar := func() {
		if query == "" {
			searchBar.SetText(" Suche: (tippen zum Filtern)")
		} else {
			searchBar.SetText(" Suche: " + query + "▌")
		}
	}

	footerText := func(row int) string {
		text := ""
		if h, ok := rowHosts[row]; ok {
			text = hostDetail(h)
			if mode == modeEdit && excludeState.Excluded[h.Alias] {
				text = "ausgeblendet · " + text
			}
		} else if len(rowHosts) == 0 {
			if query != "" {
				text = "keine Treffer für \"" + query + "\""
			} else if mode == modeMain {
				text = "alle Hosts ausgeblendet — ^E öffnet den Filter"
			}
		}
		return fmt.Sprintf(" %s │ %s", text, settingsStatus(themes[themeIdx], banners[bannerIdx], bannerAlignments[alignmentIdx]))
	}

	refreshBanner := func() {
		th := themes[themeIdx]
		bannerView.SetText(renderBanner(banners[bannerIdx], th))
		bannerView.SetTextAlign(bannerAlignments[alignmentIdx].TViewAlign)
		bannerView.SetBackgroundColor(th.Background)
		row, _ := table.GetSelection()
		footer.SetText(footerText(row))
	}

	applyTheme := func(th Theme) {
		table.SetBackgroundColor(th.Background)
		table.SetBorderColor(th.Border)
		table.SetTitleColor(th.Title)
		table.SetSelectedStyle(tcell.StyleDefault.
			Foreground(th.SelectedFg).
			Background(th.SelectedBg))
		searchBar.SetBackgroundColor(th.Background)
		searchBar.SetTextColor(th.Text)
		footer.SetBackgroundColor(th.Background)
		footer.SetTextColor(th.Detail)
		helpView.SetBackgroundColor(th.Background)
		helpView.SetTextColor(th.Text)
		helpView.SetBorderColor(th.Border)
		helpView.SetTitleColor(th.Title)
		fillTable(th)
	}

	// refresh baut die Tabelle nach Such-/Modus-Änderungen neu auf, wählt
	// den ersten Treffer vor und aktualisiert Such- und Fußzeile.
	refresh := func() {
		fillTable(themes[themeIdx])
		if firstHostRow >= 0 {
			table.Select(firstHostRow, 0)
		}
		updateSearchBar()
		row, _ := table.GetSelection()
		footer.SetText(footerText(row))
	}

	setMode := func(m int) {
		mode = m
		if mode == modeEdit {
			table.SetTitle(titleEdit)
		} else {
			table.SetTitle(titleMain)
		}
		refresh()
	}

	applyTheme(themes[themeIdx])
	refreshBanner()

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

	// Fenstermaße aus der Filter-Ansicht ableiten (Obermenge aller Zeilen).
	// Auch die Fußzeile zählt mit, damit die Detailanzeige nie abgeschnitten
	// wird.
	longestSettings := ""
	for _, th := range themes {
		for _, banner := range banners {
			for _, alignment := range bannerAlignments {
				status := settingsStatus(th, banner, alignment)
				if len([]rune(status)) > len([]rune(longestSettings)) {
					longestSettings = status
				}
			}
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
			footerLine := " ausgeblendet · " + hostDetail(h) + " │ " + longestSettings
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
	height := totalRows + 4 // Rahmen (2) + Suchzeile (1) + Fußzeile (1)

	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(searchBar, 1, 0, false).
		AddItem(bodyPages, totalRows+2, 0, true).
		AddItem(footer, 1, 0, false)
	bannerGap := tview.NewBox()
	bannerSize := bannerHeight(banners[bannerIdx]) + 1
	bannerStack := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(bannerView, bannerHeight(banners[bannerIdx]), 0, false).
		AddItem(bannerGap, 1, 0, false).
		AddItem(content, height, 0, true)
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(bannerStack, height+bannerSize, 0, true).
		AddItem(nil, 0, 1, false)
	outer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(inner, width, 0, true).
		AddItem(nil, 0, 1, false)

	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		_, screenHeight := screen.Size()
		size := 0
		if bannerVisible(screenHeight, height, banners[bannerIdx]) {
			size = bannerHeight(banners[bannerIdx]) + 1
		}
		bannerStack.ResizeItem(bannerView, max(0, size-1), 0)
		gapSize := 0
		if size > 0 {
			gapSize = 1
		}
		bannerStack.ResizeItem(bannerGap, gapSize, 0)
		inner.ResizeItem(bannerStack, height+size, 0)
		return false
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		wasHelpVisible := viewState.HelpVisible
		if handleHelpKey(event, &viewState) {
			if viewState.HelpVisible != wasHelpVisible {
				if viewState.HelpVisible {
					helpView.SetTitle(titleHelp)
					helpView.ScrollToBeginning()
					bodyPages.SwitchToPage("help")
					app.SetFocus(helpView)
				} else {
					bodyPages.SwitchToPage("hosts")
					app.SetFocus(table)
				}
				return nil
			}
			return event
		}

		switch event.Key() {
		case tcell.KeyEscape:
			if query != "" {
				query = ""
				refresh()
			} else if mode == modeEdit {
				setMode(modeMain)
			} else {
				app.Stop()
			}
			return nil
		case tcell.KeyCtrlE:
			if mode == modeEdit {
				setMode(modeMain)
			} else {
				setMode(modeEdit)
			}
			return nil
		case tcell.KeyCtrlT:
			themeIdx = nextIndex(themeIdx, len(themes))
			if err := SaveThemeName(themePath, themes[themeIdx].Name); err != nil {
				saveErr = err
				app.Stop()
				return nil
			}
			applyTheme(themes[themeIdx])
			refreshBanner()
			updateSearchBar()
			return nil
		case tcell.KeyCtrlB:
			bannerIdx = nextIndex(bannerIdx, len(banners))
			if err := SaveBannerName(bannerPath, banners[bannerIdx].Name); err != nil {
				saveErr = err
				app.Stop()
				return nil
			}
			refreshBanner()
			return nil
		case tcell.KeyCtrlA:
			alignmentIdx = nextIndex(alignmentIdx, len(bannerAlignments))
			if err := SaveBannerAlignmentName(alignmentPath, bannerAlignments[alignmentIdx].Name); err != nil {
				saveErr = err
				app.Stop()
				return nil
			}
			refreshBanner()
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if query != "" {
				runes := []rune(query)
				query = string(runes[:len(runes)-1])
				refresh()
			}
			return nil
		case tcell.KeyRune:
			r := event.Rune()
			if mode == modeEdit && r == ' ' {
				row, _ := table.GetSelection()
				if h, ok := rowHosts[row]; ok {
					toggle(h)
					footer.SetText(footerText(row))
				}
				return nil
			}
			query += string(r)
			refresh()
			return nil
		}
		return event
	})

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
