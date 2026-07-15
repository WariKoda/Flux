package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

var bannerTagPattern = regexp.MustCompile(`\[[^]]*\]`)

func stripTags(text string) string {
	return bannerTagPattern.ReplaceAllString(text, "")
}

func TestBannerCatalog(t *testing.T) {
	want := []string{
		"blurvision-rainbow3", "blurvision-monochrome",
		"single-rainbow3", "single-monochrome",
		"ansi-regular-rainbow3", "ansi-regular-monochrome",
		"banner3-rainbow3", "banner3-monochrome",
		"ansi-compact-rainbow3", "ansi-compact-monochrome",
		"none",
	}
	got := make([]string, len(banners))
	for i := range banners {
		got[i] = banners[i].Name
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Banner-Katalog: %v", got)
	}
}

func TestExactBannerForms(t *testing.T) {
	tests := []struct {
		name   string
		family BannerFamily
		rows   [][]string
	}{
		{"BlurVision", blurVisionFamily, [][]string{
			{"░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░", "░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░", "░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░", "░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░", "░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░"},
		}},
		{"single", singleFamily, [][]string{{"▓▒░ FLUX ░▒▓"}}},
		{"ANSI Regular", ansiRegularFamily, [][]string{{"███████ ██      ██    ██ ██   ██", "██      ██      ██    ██  ██ ██", "█████▌  ██      ██    ██   ███", "██      ██      ██    ██  ██ ██", "██      ███████  ██████  ██   ██"}}},
		{"Banner3", banner3Family, [][]string{{"######## ##       ##     ## ##     ##", "##       ##       ##     ##  ##   ##", "##       ##       ##     ##   ## ##", "######   ##       ##     ##    ###", "##       ##       ##     ##   ## ##", "##       ##       ##     ##  ##   ##", "##       ########  #######  ##     ##"}}},
		{"ANSI Compact", ansiCompactFamily, [][]string{{"██████ ▄▄    ▄▄ ▄▄ ▄▄ ▄▄", "██▄▄   ██    ██ ██ ▀█▄█▀", "██     ██▄▄▄ ▀███▀ ██ ██"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.family.Forms) != len(tt.rows) {
				t.Fatalf("Formanzahl: %d", len(tt.family.Forms))
			}
			for i := range tt.rows {
				if !reflect.DeepEqual(tt.family.Forms[i].Rows, tt.rows[i]) {
					t.Errorf("Form %d:\n%q\nwant:\n%q", i, tt.family.Forms[i].Rows, tt.rows[i])
				}
			}
		})
	}
	if len(noneFamily.Forms) != 0 {
		t.Fatalf("none hat %d Formen", len(noneFamily.Forms))
	}
	if got := ansiCompactFamily.Forms[0].Name; got != "ansi-compact" {
		t.Fatalf("ANSI Compact Formname = %q, ansi-compact erwartet", got)
	}
	if got := blurVisionFamily.Forms[0].Rows[4]; got != "░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░" {
		t.Fatalf("kompakte Abschlusszeile: %q", got)
	}
}

func TestLegacyBannerNamesNormalize(t *testing.T) {
	cases := map[string]string{"ansi": "blurvision-rainbow3", "monochrome": "blurvision-monochrome", "wordmark-ansi": "single-rainbow3", "wordmark-mono": "single-monochrome", "terminal-ansi": "ansi-regular-rainbow3", "terminal-mono": "ansi-regular-monochrome", "terrace-rainbow3": "blurvision-rainbow3", "terrace-monochrome": "blurvision-monochrome"}
	for in, want := range cases {
		if got, err := normalizeBannerName(in); err != nil || got != want {
			t.Errorf("%q: %q, %v", in, got, err)
		}
	}
	if _, err := normalizeBannerName("unknown"); err == nil {
		t.Fatal("unbekannter Modus muss fehlschlagen")
	}
}

func TestREADMEBannerCatalogMatchesCurrentChoices(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(data)

	for _, stale := range []string{
		"**Terrace** —",
		"sechs Banner-Designs",
		"dreizehn",
		"13 Zustände",
		"siebenzeilige Form, dann die vollständige fünfzeilige Form",
	} {
		if strings.Contains(readme, stale) {
			t.Errorf("README enthält veraltete Banner-Dokumentation %q", stale)
		}
	}
	for legacy, current := range map[string]string{
		"terrace-rainbow3":   "blurvision-rainbow3",
		"terrace-monochrome": "blurvision-monochrome",
	} {
		if !strings.Contains(readme, "`"+legacy+"` wird `"+current+"`") {
			t.Errorf("README dokumentiert Legacy-Zuordnung %s -> %s nicht", legacy, current)
		}
	}
}

func TestRainbow3UsesTAAGDiagonalPhase(t *testing.T) {
	form := BannerForm{Rows: []string{"abcd", "efgh", "ijkl"}}
	got := renderBanner(form, BannerMode{ColorMode: bannerRainbow3}, Theme{})
	want := "[#ff2828]a[#ff7800]bc[#ffb400]d\n" +
		"[#ff7800]ef[#ffb400]gh\n" +
		"[#ff7800]i[#ffb400]jk[#ffdc00]l"
	if got != want {
		t.Fatalf("rainbow3:\n%s\nwant:\n%s", got, want)
	}
}

func TestRainbow3SpaceConsumesDisplayCellAndAdvancesPhase(t *testing.T) {
	form := BannerForm{Rows: []string{"a b", "a b"}}
	got := renderBanner(form, BannerMode{ColorMode: bannerRainbow3}, Theme{})
	want := "[#ff2828]a[#ff7800] b\n" +
		"[#ff7800]a [#ffb400]b"
	if got != want {
		t.Fatalf("rainbow3 space phase:\n%q\nwant:\n%q", got, want)
	}
}

func TestRainbow3WideRuneAdvancesByDisplayWidth(t *testing.T) {
	form := BannerForm{Rows: []string{"界x", "z"}}
	got := renderBanner(form, BannerMode{ColorMode: bannerRainbow3}, Theme{})
	want := "[#ff2828]界[#ff7800]x\n[#ff7800]z"
	if got != want {
		t.Fatalf("rainbow3 wide-rune phase:\n%q\nwant:\n%q", got, want)
	}
}

func TestSingleRowUsesEveryRainbow3ColorOnce(t *testing.T) {
	wantColors := []string{
		"#ff2828", "#ff7800", "#ffb400", "#ffdc00", "#dcff00", "#78ff00",
		"#00ff50", "#00ffa0", "#00c8ff", "#0078ff", "#7850ff", "#ff00c8",
	}
	if !reflect.DeepEqual(bannerRainbow3Colors, wantColors) {
		t.Fatalf("rainbow3-Palette: %v", bannerRainbow3Colors)
	}
	got := renderBanner(singleFamily.Forms[0], banners[2], Theme{})
	for _, color := range bannerRainbow3Colors {
		if strings.Count(got, "["+color+"]") != 1 {
			t.Errorf("Farbe %s nicht genau einmal: %q", color, got)
		}
	}
}

func TestNoneBannerRendersEmpty(t *testing.T) {
	if got := renderBanner(compactBanner, banners[len(banners)-1], Theme{}); got != "" {
		t.Fatalf("none: %q", got)
	}
}

func TestMonochromeBannerUsesThemeText(t *testing.T) {
	th := Theme{Text: tcell.NewHexColor(0x123456), Header: tcell.NewHexColor(0xabcdef)}
	got := renderBanner(compactBanner, banners[1], th)
	if !strings.Contains(got, "[#123456]") || strings.Contains(got, "[#abcdef]") {
		t.Fatalf("falsche Monochromfarbe: %q", got)
	}
}

func TestMonochromeBannerUsesThemeColor(t *testing.T) {
	th := Theme{Text: tcell.NewHexColor(0x123456)}
	got := renderBanner(compactBanner, banners[1], th)
	if !strings.Contains(got, "[#123456]") {
		t.Fatalf("Theme-Farbe fehlt: %q", got)
	}
	if strings.Contains(got, "[#ff5555]") {
		t.Fatalf("ANSI-Farbe in Monochrom-Banner: %q", got)
	}
}

func TestRainbow3BannerIsThemeIndependent(t *testing.T) {
	a := renderBanner(compactBanner, banners[0], Theme{Text: tcell.ColorRed})
	b := renderBanner(compactBanner, banners[0], Theme{Text: tcell.ColorBlue})
	if a != b {
		t.Fatalf("ANSI-Banner darf sich mit Theme nicht ändern")
	}
	if !strings.Contains(a, "[#ff2828]") || !strings.Contains(a, "[#ff7800]") {
		t.Fatalf("rainbow3-Palette fehlt: %q", a)
	}
}

func TestRainbow3BannerPreservesCombiningRunes(t *testing.T) {
	form := BannerForm{Rows: []string{"a\u0301"}}
	if got := renderBanner(form, banners[0], Theme{}); got != "[#ff2828]a\u0301" {
		t.Fatalf("kombinierte Rune nicht zusammenhängend erhalten: %q", got)
	}
}

func TestRainbow3BannerPreservesWideRuneColorForCombiningMark(t *testing.T) {
	form := BannerForm{Rows: []string{"界\u0301"}}
	if got := renderBanner(form, banners[0], Theme{}); got != "[#ff2828]界\u0301" {
		t.Fatalf("kombinierte Rune nach breiter Basis wechselte die Farbe: %q", got)
	}
}

func TestBannerVisibilityRequiresWholeBannerAndGap(t *testing.T) {
	form := BannerForm{Rows: []string{"1", "2", "3", "4"}}
	if !bannerVisible(15, 10, form) {
		t.Fatal("4 Zeilen + Abstand müssen exakt passen")
	}
	if bannerVisible(14, 10, form) {
		t.Fatal("eine Zeile zu wenig muss Banner ausblenden")
	}
}

func TestAlignedBannerText(t *testing.T) {
	form := BannerForm{Rows: []string{"FLUX"}}
	th := Theme{Text: tcell.ColorGreen}
	if got := stripTags(alignedBannerText(form, banners[1], 8, bannerAlignments[0], th)); got != "FLUX" {
		t.Errorf("links: %q", got)
	}
	if got := stripTags(alignedBannerText(form, banners[1], 8, bannerAlignments[1], th)); got != "  FLUX" {
		t.Errorf("mitte: %q", got)
	}
	if got := stripTags(alignedBannerText(form, banners[1], 8, bannerAlignments[2], th)); got != "    FLUX" {
		t.Errorf("rechts: %q", got)
	}
}

func TestAlignedBannerBlock(t *testing.T) {
	form := BannerForm{Rows: []string{"123456", "12", "1234"}}
	theme := Theme{Text: tcell.ColorGreen}
	tests := []struct {
		name      string
		alignment BannerAlignment
		want      string
	}{
		{name: "left", alignment: bannerAlignments[0], want: "123456\n12\n1234"},
		{name: "center", alignment: bannerAlignments[1], want: "  123456\n  12\n  1234"},
		{name: "right", alignment: bannerAlignments[2], want: "    123456\n    12\n    1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripTags(alignedBannerText(form, banners[1], 10, tt.alignment, theme))
			if got != tt.want {
				t.Fatalf("Blockausrichtung:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestAlignedBannerTextUsesDisplayWidth(t *testing.T) {
	form := BannerForm{Rows: []string{"界"}}
	got := stripTags(alignedBannerText(form, banners[1], 4, bannerAlignments[1], Theme{Text: tcell.ColorGreen}))
	if got != " 界" {
		t.Fatalf("Unicode-Anzeigebreite nicht berücksichtigt: %q", got)
	}
}

func TestAlignedBannerPaddingDoesNotChangeRainbowPhase(t *testing.T) {
	form := BannerForm{Rows: []string{"abcd"}}
	for _, alignment := range bannerAlignments[1:] {
		got := alignedBannerText(form, banners[0], 8, alignment, Theme{})
		if !strings.Contains(got, " [#ff2828]a") {
			t.Errorf("%s: Padding/Farbe falsch: %q", alignment.Name, got)
		}
	}
}

func TestBannerDefinitionsAndCycleOrder(t *testing.T) {
	want := []string{"blurvision-rainbow3", "blurvision-monochrome", "single-rainbow3", "single-monochrome", "ansi-regular-rainbow3", "ansi-regular-monochrome", "banner3-rainbow3", "banner3-monochrome", "ansi-compact-rainbow3", "ansi-compact-monochrome", "none"}
	if len(banners) != len(want) {
		t.Fatalf("%d Banner erwartet, %d erhalten", len(want), len(banners))
	}
	for i, name := range want {
		if banners[i].Name != name {
			t.Errorf("Banner %d: %q erwartet, %q erhalten", i, name, banners[i].Name)
		}
	}
	if got := nextIndex(10, len(banners)); got != 0 {
		t.Errorf("Wraparound: 0 erwartet, %d erhalten", got)
	}
}

func TestBannerAlignmentDefinitionsAndCycleOrder(t *testing.T) {
	want := []string{"left", "center", "right"}
	if len(bannerAlignments) != len(want) {
		t.Fatalf("%d Ausrichtungen erwartet, %d erhalten", len(want), len(bannerAlignments))
	}
	for i, name := range want {
		if bannerAlignments[i].Name != name {
			t.Errorf("Ausrichtung %d: %q erwartet, %q erhalten", i, name, bannerAlignments[i].Name)
		}
	}
	if got := nextIndex(2, len(bannerAlignments)); got != 0 {
		t.Errorf("Wraparound: 0 erwartet, %d erhalten", got)
	}
}

func TestBannerAndAlignmentLookup(t *testing.T) {
	if got, err := bannerIndex("ansi-compact-monochrome"); err != nil || got != 9 {
		t.Fatalf("Banner-Lookup: Index %d, Fehler %v", got, err)
	}
	if got, err := bannerAlignmentIndex("center"); err != nil || got != 1 {
		t.Fatalf("Ausrichtungs-Lookup: Index %d, Fehler %v", got, err)
	}
}

func TestNextIndexPanicsForNonPositiveLength(t *testing.T) {
	for _, length := range []int{0, -1} {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("nextIndex muss für Länge %d panic auslösen", length)
				}
			}()
			nextIndex(0, length)
		})
	}
}

func TestUnknownBannerAndAlignmentFail(t *testing.T) {
	if _, err := bannerIndex("unknown"); err == nil {
		t.Fatal("unbekannter Banner muss fehlschlagen")
	}
	if _, err := bannerAlignmentIndex("unknown"); err == nil {
		t.Fatal("unbekannte Ausrichtung muss fehlschlagen")
	}
}

func TestBannerSettingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	bannerPath := filepath.Join(dir, "nested", "banner")
	alignPath := filepath.Join(dir, "nested", "banner-alignment")
	if err := SaveBannerName(bannerPath, "banner3-monochrome"); err != nil {
		t.Fatal(err)
	}
	if err := SaveBannerAlignmentName(alignPath, "right"); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Dir(bannerPath)); err != nil {
		t.Fatal(err)
	} else if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("Verzeichnismodus: %04o, erwartet 0700", got)
	}
	for _, path := range []string{bannerPath, alignPath} {
		if info, err := os.Stat(path); err != nil {
			t.Fatal(err)
		} else if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("Dateimodus %s: %04o, erwartet 0600", path, got)
		}
	}
	if got, err := LoadBannerName(bannerPath); err != nil || got != "banner3-monochrome" {
		t.Fatalf("Banner: %q, %v", got, err)
	}
	if got, err := LoadBannerAlignmentName(alignPath); err != nil || got != "right" {
		t.Fatalf("Ausrichtung: %q, %v", got, err)
	}
}

func TestLegacyBannerSettingsLoadNormalizedAndSaveStableName(t *testing.T) {
	cases := map[string]string{
		"ansi": "blurvision-rainbow3", "monochrome": "blurvision-monochrome",
		"wordmark-ansi": "single-rainbow3", "wordmark-mono": "single-monochrome",
		"terminal-ansi": "ansi-regular-rainbow3", "terminal-mono": "ansi-regular-monochrome",
		"terrace-rainbow3": "blurvision-rainbow3", "terrace-monochrome": "blurvision-monochrome",
	}
	for legacy, want := range cases {
		t.Run(legacy, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "banner")
			if err := os.WriteFile(path, []byte("  "+legacy+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := LoadBannerName(path)
			if err != nil || got != want {
				t.Fatalf("LoadBannerName: %q, %v", got, err)
			}
			if err := SaveBannerName(path, got); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != want+"\n" {
				t.Fatalf("gespeicherter Modus: %q", data)
			}
		})
	}
}

func TestSaveBannerNameRejectsLegacyNames(t *testing.T) {
	for _, name := range []string{"wordmark-ansi", "terrace-rainbow3", "terrace-monochrome"} {
		path := filepath.Join(t.TempDir(), "banner")
		if err := SaveBannerName(path, name); err == nil {
			t.Errorf("Legacy-Modus %q darf nicht gespeichert werden", name)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("ungültiger Modus %q hat Datei erzeugt: %v", name, err)
		}
	}
}

func TestBannerSettingsRoundTripsNone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "banner")
	if err := SaveBannerName(path, "none"); err != nil {
		t.Fatal(err)
	}
	if got, err := LoadBannerName(path); err != nil || got != "none" {
		t.Fatalf("none: %q, %v", got, err)
	}
}

func TestMissingBannerSettingsUseDefaults(t *testing.T) {
	dir := t.TempDir()
	if got, err := LoadBannerName(filepath.Join(dir, "banner")); err != nil || got != banners[0].Name {
		t.Fatalf("Banner: %q, %v", got, err)
	}
	if got, err := LoadBannerAlignmentName(filepath.Join(dir, "align")); err != nil || got != bannerAlignments[0].Name {
		t.Fatalf("Ausrichtung: %q, %v", got, err)
	}
}

func TestInvalidBannerSettingsFail(t *testing.T) {
	for _, tc := range []struct {
		name        string
		value       string
		load        func(string) (string, error)
		errorPrefix string
	}{
		{"leerer Banner", " \n", LoadBannerName, "Banner-Datei nicht lesbar:"},
		{"unbekannter Banner", "wat\n", LoadBannerName, "Banner-Datei nicht lesbar:"},
		{"leere Ausrichtung", "\n", LoadBannerAlignmentName, "Banner-Ausrichtungsdatei nicht lesbar:"},
		{"unbekannte Ausrichtung", "diagonal\n", LoadBannerAlignmentName, "Banner-Ausrichtungsdatei nicht lesbar:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "setting")
			if err := os.WriteFile(path, []byte(tc.value), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := tc.load(path); err == nil {
				t.Fatal("Fehler erwartet")
			} else if !strings.HasPrefix(err.Error(), tc.errorPrefix) {
				t.Fatalf("Fehler %q beginnt nicht mit %q", err, tc.errorPrefix)
			}
		})
	}
}

func TestUnreadableBannerSettingsUseDistinctErrorContexts(t *testing.T) {
	for _, tc := range []struct {
		name        string
		load        func(string) (string, error)
		errorPrefix string
	}{
		{"Banner", LoadBannerName, "Banner-Datei nicht lesbar:"},
		{"Ausrichtung", LoadBannerAlignmentName, "Banner-Ausrichtungsdatei nicht lesbar:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := t.TempDir()
			if _, err := tc.load(path); err == nil {
				t.Fatal("Fehler erwartet")
			} else if !strings.HasPrefix(err.Error(), tc.errorPrefix) {
				t.Fatalf("Fehler %q beginnt nicht mit %q", err, tc.errorPrefix)
			}
		})
	}
}
