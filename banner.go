package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
)

type BannerColorMode int

const (
	bannerANSI BannerColorMode = iota
	bannerMonochrome
)

type BannerMode struct {
	Name, DisplayName string
	ColorMode         BannerColorMode
}

type BannerForm struct {
	Name string
	Rows []string
}

type BannerAlignment struct {
	Name, DisplayName string
	TViewAlign        int
}

var banners = []BannerMode{
	{"ansi", "ANSI", bannerANSI},
	{"monochrome", "Monochrom", bannerMonochrome},
}

var largeBanner = BannerForm{Name: "large", Rows: []string{
	"░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░",
}}

var compactBanner = BannerForm{Name: "compact", Rows: []string{
	"░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
}}

var bannerAlignments = []BannerAlignment{
	{"left", "Links", tview.AlignLeft},
	{"center", "Mitte", tview.AlignCenter},
	{"right", "Rechts", tview.AlignRight},
}

var bannerANSIColors = []string{"#ff5555", "#f1fa8c", "#50fa7b", "#8be9fd", "#6272a4", "#bd93f9", "#ff79c6", "#ffb86c"}

func renderBanner(form BannerForm, mode BannerMode, theme Theme) string {
	return renderBannerRows(form.Rows, mode.ColorMode, theme)
}

func renderBannerRows(rows []string, colorMode BannerColorMode, theme Theme) string {
	rendered := make([]string, len(rows))
	for i, row := range rows {
		if colorMode == bannerMonochrome {
			rendered[i] = fmt.Sprintf("[#%06x]%s", theme.Text.Hex(), row)
			continue
		}

		rowWidth := runewidth.StringWidth(row)
		position := 0
		lastColor := -1
		var output strings.Builder
		for _, r := range row {
			runeWidth := runewidth.RuneWidth(r)
			colorPosition := position
			color := lastColor
			if runeWidth != 0 || lastColor < 0 {
				color = min(colorPosition*len(bannerANSIColors)/max(1, rowWidth), len(bannerANSIColors)-1)
			}
			if color != lastColor {
				output.WriteString("[")
				output.WriteString(bannerANSIColors[color])
				output.WriteString("]")
				lastColor = color
			}
			output.WriteRune(r)
			position += runeWidth
		}
		rendered[i] = output.String()
	}
	return strings.Join(rendered, "\n")
}

func bannerHeight(form BannerForm) int {
	return len(form.Rows)
}

func bannerVisible(screenHeight, tuiHeight int, form BannerForm) bool {
	return screenHeight >= tuiHeight+bannerHeight(form)+1
}

func alignedBannerText(form BannerForm, mode BannerMode, width int, alignment BannerAlignment, theme Theme) string {
	rows := make([]string, len(form.Rows))
	for i, row := range form.Rows {
		rowWidth := runewidth.StringWidth(row)
		if width <= rowWidth {
			rows[i] = row
			continue
		}

		padding := width - rowWidth
		switch alignment.TViewAlign {
		case tview.AlignCenter:
			padding /= 2
		case tview.AlignRight:
		default:
			padding = 0
		}
		rows[i] = strings.Repeat(" ", padding) + row
	}
	return renderBannerRows(rows, mode.ColorMode, theme)
}

func normalizeBannerName(name string) (string, error) {
	switch name {
	case "ansi", "wordmark-ansi", "terminal-ansi":
		return "ansi", nil
	case "monochrome", "wordmark-mono", "terminal-mono":
		return "monochrome", nil
	default:
		return "", fmt.Errorf("unbekannter Banner %q (verfügbar: ansi, monochrome)", name)
	}
}

func bannerIndex(name string) (int, error) {
	names := make([]string, len(banners))
	for i, banner := range banners {
		if banner.Name == name {
			return i, nil
		}
		names[i] = banner.Name
	}
	return 0, fmt.Errorf("unbekannter Banner %q (verfügbar: %s)", name, strings.Join(names, ", "))
}

func bannerAlignmentIndex(name string) (int, error) {
	names := make([]string, len(bannerAlignments))
	for i, alignment := range bannerAlignments {
		if alignment.Name == name {
			return i, nil
		}
		names[i] = alignment.Name
	}
	return 0, fmt.Errorf("unbekannte Banner-Ausrichtung %q (verfügbar: %s)", name, strings.Join(names, ", "))
}

func nextIndex(current, length int) int {
	if length <= 0 {
		panic("Zykluslänge muss positiv sein")
	}
	return (current + 1) % length
}

func loadChoice(path, errorContext, defaultName string, validate func(string) error) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultName, nil
		}
		return "", fmt.Errorf("%s: %w", errorContext, err)
	}

	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", fmt.Errorf("%s: Datei ist leer", errorContext)
	}
	if err := validate(name); err != nil {
		return "", fmt.Errorf("%s: %w", errorContext, err)
	}
	return name, nil
}

func saveChoice(path, errorContext, name string, validate func(string) error) error {
	if err := validate(name); err != nil {
		return fmt.Errorf("%s: %w", errorContext, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("%s: %w", errorContext, err)
	}
	if err := os.WriteFile(path, []byte(name+"\n"), 0o600); err != nil {
		return fmt.Errorf("%s: %w", errorContext, err)
	}
	return nil
}

func LoadBannerName(path string) (string, error) {
	name, err := loadChoice(path, "Banner-Datei nicht lesbar", banners[0].Name, func(name string) error {
		_, err := normalizeBannerName(name)
		return err
	})
	if err != nil {
		return "", err
	}
	return normalizeBannerName(name)
}

func SaveBannerName(path, name string) error {
	return saveChoice(path, "Banner nicht speicherbar", name, func(name string) error {
		_, err := bannerIndex(name)
		return err
	})
}

func LoadBannerAlignmentName(path string) (string, error) {
	return loadChoice(path, "Banner-Ausrichtungsdatei nicht lesbar", bannerAlignments[0].Name, func(name string) error {
		_, err := bannerAlignmentIndex(name)
		return err
	})
}

func SaveBannerAlignmentName(path, name string) error {
	return saveChoice(path, "Banner-Ausrichtung nicht speicherbar", name, func(name string) error {
		_, err := bannerAlignmentIndex(name)
		return err
	})
}
