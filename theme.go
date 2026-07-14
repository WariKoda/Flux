package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Theme bündelt die Farben der Oberfläche.
type Theme struct {
	Name        string // Kennung für die Theme-Datei
	DisplayName string // Anzeigename in der Fußzeile
	Background  tcell.Color
	Text        tcell.Color
	Header      tcell.Color // Server-Überschriften
	Detail      tcell.Color // Zweitspalte und Fußzeile
	Border      tcell.Color
	Title       tcell.Color
	SelectedFg  tcell.Color
	SelectedBg  tcell.Color
}

// Okabe-Ito-Palette (Color Universal Design): für Protanopie, Deuteranopie
// und Tritanopie unterscheidbar, mit distinkter Helligkeit für Graustufen.
var (
	okabeOrange     = tcell.NewHexColor(0xE69F00)
	okabeSkyBlue    = tcell.NewHexColor(0x56B4E9)
	okabeYellow     = tcell.NewHexColor(0xF0E442)
	okabeBlue       = tcell.NewHexColor(0x0072B2)
	okabeVermillion = tcell.NewHexColor(0xD55E00)
)

// themes definiert die verfügbaren Themes; die Reihenfolge ist zugleich die
// Zyklus-Reihenfolge des Shortcuts 't'. themes[0] ist der Default.
var themes = []Theme{
	{
		Name:        "dark",
		DisplayName: "Dunkel",
		Background:  tcell.ColorBlack,
		Text:        tcell.ColorWhite,
		Header:      tcell.ColorAqua,
		Detail:      tcell.ColorSilver,
		Border:      tcell.ColorTeal,
		Title:       tcell.ColorWhite,
		SelectedFg:  tcell.ColorBlack,
		SelectedBg:  tcell.ColorAqua,
	},
	{
		Name:        "light",
		DisplayName: "Hell",
		Background:  tcell.ColorWhite,
		Text:        tcell.ColorBlack,
		Header:      tcell.ColorNavy,
		Detail:      tcell.ColorDarkSlateGray,
		Border:      tcell.ColorNavy,
		Title:       tcell.ColorBlack,
		SelectedFg:  tcell.ColorWhite,
		SelectedBg:  tcell.ColorNavy,
	},
	{
		Name:        "matrix",
		DisplayName: "Matrix",
		Background:  tcell.ColorBlack,
		Text:        tcell.ColorGreen,
		Header:      tcell.ColorLightYellow,
		Detail:      tcell.ColorGreen,
		Border:      tcell.ColorGreen,
		Title:       tcell.ColorGreen,
		SelectedFg:  tcell.ColorBlack,
		SelectedBg:  tcell.ColorGreen,
	},
	// Farbenblind-freundliche Themes auf Basis der Okabe-Ito-Palette:
	// Unterscheidung läuft über die Blau/Orange-Achse (in allen häufigen
	// Farbfehlsichtigkeiten trennbar), nie über Rot/Grün.
	{
		Name:        "cb-dark",
		DisplayName: "Okabe-Ito Dunkel",
		Background:  tcell.ColorBlack,
		Text:        tcell.ColorWhite,
		Header:      okabeSkyBlue,
		Detail:      okabeYellow,
		Border:      okabeBlue,
		Title:       tcell.ColorWhite,
		SelectedFg:  tcell.ColorBlack,
		SelectedBg:  okabeOrange,
	},
	{
		Name:        "cb-light",
		DisplayName: "Okabe-Ito Hell",
		Background:  tcell.ColorWhite,
		Text:        tcell.ColorBlack,
		Header:      okabeBlue,
		Detail:      okabeVermillion,
		Border:      okabeBlue,
		Title:       tcell.ColorBlack,
		SelectedFg:  tcell.ColorBlack,
		SelectedBg:  okabeOrange,
	},
}

// themeIndex liefert die Position des Themes mit dem gegebenen Namen.
// Ein unbekannter Name ist ein Fehler (z. B. kaputte Theme-Datei).
func themeIndex(name string) (int, error) {
	names := make([]string, len(themes))
	for i, t := range themes {
		if t.Name == name {
			return i, nil
		}
		names[i] = t.Name
	}
	return 0, fmt.Errorf("unbekanntes Theme %q (verfügbar: %s)", name, strings.Join(names, ", "))
}

// LoadThemeName liest den gespeicherten Theme-Namen. Eine fehlende Datei ist
// gültig (Default-Theme); jede andere Lesestörung und eine leere Datei sind
// Fehler.
func LoadThemeName(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return themes[0].Name, nil
		}
		return "", fmt.Errorf("Theme-Datei nicht lesbar: %w", err)
	}
	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", fmt.Errorf("Theme-Datei %s ist leer", path)
	}
	return name, nil
}

// SaveThemeName persistiert den Theme-Namen für den nächsten Start.
func SaveThemeName(path, name string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("Theme-Verzeichnis nicht anlegbar: %w", err)
	}
	if err := os.WriteFile(path, []byte(name+"\n"), 0o600); err != nil {
		return fmt.Errorf("Theme nicht speicherbar: %w", err)
	}
	return nil
}
