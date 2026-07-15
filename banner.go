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
	bannerRainbow3 BannerColorMode = iota
	bannerMonochrome
	bannerNone
)

type BannerFamily struct {
	Name  string
	Forms []BannerForm
}

type BannerMode struct {
	Name, DisplayName string
	Family            BannerFamily
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
	{"blurvision-rainbow3", "BlurVision · Regenbogen", blurVisionFamily, bannerRainbow3},
	{"blurvision-monochrome", "BlurVision · Monochrom", blurVisionFamily, bannerMonochrome},
	{"single-rainbow3", "Single · Regenbogen", singleFamily, bannerRainbow3},
	{"single-monochrome", "Single · Monochrom", singleFamily, bannerMonochrome},
	{"ansi-regular-rainbow3", "ANSI Regular · Regenbogen", ansiRegularFamily, bannerRainbow3},
	{"ansi-regular-monochrome", "ANSI Regular · Monochrom", ansiRegularFamily, bannerMonochrome},
	{"banner3-rainbow3", "Banner3 · Regenbogen", banner3Family, bannerRainbow3},
	{"banner3-monochrome", "Banner3 · Monochrom", banner3Family, bannerMonochrome},
	{"ansi-compact-rainbow3", "ANSI Compact · Regenbogen", ansiCompactFamily, bannerRainbow3},
	{"ansi-compact-monochrome", "ANSI Compact · Monochrom", ansiCompactFamily, bannerMonochrome},
	{"none", "Kein Banner", noneFamily, bannerNone},
}

var compactBanner = BannerForm{Name: "compact", Rows: []string{
	"░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░",
	"░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░",
	"░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░",
}}

var blurVisionFamily = BannerFamily{Name: "blurvision", Forms: []BannerForm{compactBanner}}
var singleFamily = BannerFamily{Name: "single", Forms: []BannerForm{{Name: "single", Rows: []string{"▓▒░ FLUX ░▒▓"}}}}
var ansiRegularFamily = BannerFamily{Name: "ansi-regular", Forms: []BannerForm{{Name: "regular", Rows: []string{
	"███████ ██      ██    ██ ██   ██", "██      ██      ██    ██  ██ ██", "█████▌  ██      ██    ██   ███", "██      ██      ██    ██  ██ ██", "██      ███████  ██████  ██   ██",
}}}}
var banner3Family = BannerFamily{Name: "banner3", Forms: []BannerForm{{Name: "banner3", Rows: []string{
	"######## ##       ##     ## ##     ##", "##       ##       ##     ##  ##   ##", "##       ##       ##     ##   ## ##", "######   ##       ##     ##    ###", "##       ##       ##     ##   ## ##", "##       ##       ##     ##  ##   ##", "##       ########  #######  ##     ##",
}}}}
var ansiCompactFamily = BannerFamily{Name: "ansi-compact", Forms: []BannerForm{{Name: "ansi-compact", Rows: []string{
	"██████ ▄▄    ▄▄ ▄▄ ▄▄ ▄▄", "██▄▄   ██    ██ ██ ▀█▄█▀", "██     ██▄▄▄ ▀███▀ ██ ██",
}}}}
var noneFamily = BannerFamily{Name: "none"}

var bannerAlignments = []BannerAlignment{
	{"left", "Links", tview.AlignLeft},
	{"center", "Mitte", tview.AlignCenter},
	{"right", "Rechts", tview.AlignRight},
}

var bannerRainbow3Colors = []string{
	"#ff2828", "#ff7800", "#ffb400", "#ffdc00", "#dcff00", "#78ff00",
	"#00ff50", "#00ffa0", "#00c8ff", "#0078ff", "#7850ff", "#ff00c8",
}

func renderBanner(form BannerForm, mode BannerMode, theme Theme) string {
	if mode.ColorMode == bannerNone {
		return ""
	}
	return renderBannerRows(form.Rows, mode, theme)
}

func renderBannerRows(rows []string, mode BannerMode, theme Theme) string {
	rendered := make([]string, len(rows))
	for i, row := range rows {
		if mode.ColorMode == bannerMonochrome {
			rendered[i] = fmt.Sprintf("[#%06x]%s", theme.Text.Hex(), row)
			continue
		}

		position := 0
		lastColor := -1
		var output strings.Builder
		for _, r := range row {
			runeWidth := runewidth.RuneWidth(r)
			color := lastColor
			if runeWidth != 0 || lastColor < 0 {
				if mode.Family.Name == singleFamily.Name {
					color = position % len(bannerRainbow3Colors)
				} else {
					color = ((position + i + 1) / 2) % len(bannerRainbow3Colors)
				}
			}
			if color != lastColor {
				output.WriteString("[")
				output.WriteString(bannerRainbow3Colors[color])
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
	rows := strings.Split(renderBanner(form, mode, theme), "\n")
	for i, row := range form.Rows {
		rowWidth := runewidth.StringWidth(row)
		if width <= rowWidth {
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
		rows[i] = strings.Repeat(" ", padding) + rows[i]
	}
	return strings.Join(rows, "\n")
}

func normalizeBannerName(name string) (string, error) {
	if _, err := bannerIndex(name); err == nil {
		return name, nil
	}
	switch name {
	case "ansi":
		return "blurvision-rainbow3", nil
	case "monochrome":
		return "blurvision-monochrome", nil
	case "wordmark-ansi":
		return "single-rainbow3", nil
	case "wordmark-mono":
		return "single-monochrome", nil
	case "terminal-ansi":
		return "ansi-regular-rainbow3", nil
	case "terminal-mono":
		return "ansi-regular-monochrome", nil
	case "terrace-rainbow3":
		return "blurvision-rainbow3", nil
	case "terrace-monochrome":
		return "blurvision-monochrome", nil
	default:
		return "", fmt.Errorf("unbekannter Banner %q", name)
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
