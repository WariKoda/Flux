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

func TestBannerModesAndForms(t *testing.T) {
	if got := []string{banners[0].Name, banners[1].Name}; !reflect.DeepEqual(got, []string{"ansi", "monochrome"}) {
		t.Fatalf("Banner-Modi: %v", got)
	}
	if len(largeBanner.Rows) != 7 || len(compactBanner.Rows) != 5 {
		t.Fatalf("Banner-Höhen falsch")
	}
	for i, row := range compactBanner.Rows {
		if strings.HasPrefix(row, " ") || strings.HasSuffix(row, " ") {
			t.Errorf("Kompaktzeile %d gepolstert: %q", i, row)
		}
	}
	if largeBanner.Rows[0] != "░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░" {
		t.Fatal("große Wortmarke verändert")
	}
	if compactBanner.Rows[2] != "░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░" {
		t.Fatal("kompakte Wortmarke verändert")
	}
}

func TestLegacyBannerNamesNormalize(t *testing.T) {
	cases := map[string]string{"wordmark-ansi": "ansi", "terminal-ansi": "ansi", "wordmark-mono": "monochrome", "terminal-mono": "monochrome", "ansi": "ansi", "monochrome": "monochrome"}
	for in, want := range cases {
		if got, err := normalizeBannerName(in); err != nil || got != want {
			t.Errorf("%q: %q, %v", in, got, err)
		}
	}
	if _, err := normalizeBannerName("unknown"); err == nil {
		t.Fatal("unbekannter Modus muss fehlschlagen")
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

func TestANSIBannerIsThemeIndependent(t *testing.T) {
	a := renderBanner(compactBanner, banners[0], Theme{Text: tcell.ColorRed})
	b := renderBanner(compactBanner, banners[0], Theme{Text: tcell.ColorBlue})
	if a != b {
		t.Fatalf("ANSI-Banner darf sich mit Theme nicht ändern")
	}
	if !strings.Contains(a, "[#ff5555]") || !strings.Contains(a, "[#8be9fd]") {
		t.Fatalf("ANSI-Palette fehlt: %q", a)
	}
}

func TestANSIBannerPreservesCombiningRunes(t *testing.T) {
	form := BannerForm{Rows: []string{"a\u0301"}}
	if got := renderBanner(form, banners[0], Theme{}); got != "[#ff5555]a\u0301" {
		t.Fatalf("kombinierte Rune nicht zusammenhängend erhalten: %q", got)
	}
}

func TestANSIBannerPreservesWideRuneColorForCombiningMark(t *testing.T) {
	form := BannerForm{Rows: []string{"界\u0301"}}
	if got := renderBanner(form, banners[0], Theme{}); got != "[#ff5555]界\u0301" {
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

func TestAlignedBannerTextUsesDisplayWidth(t *testing.T) {
	form := BannerForm{Rows: []string{"界"}}
	got := stripTags(alignedBannerText(form, banners[1], 4, bannerAlignments[1], Theme{Text: tcell.ColorGreen}))
	if got != " 界" {
		t.Fatalf("Unicode-Anzeigebreite nicht berücksichtigt: %q", got)
	}
}

func TestBannerDefinitionsAndCycleOrder(t *testing.T) {
	want := []string{"ansi", "monochrome"}
	if len(banners) != len(want) {
		t.Fatalf("%d Banner erwartet, %d erhalten", len(want), len(banners))
	}
	for i, name := range want {
		if banners[i].Name != name {
			t.Errorf("Banner %d: %q erwartet, %q erhalten", i, name, banners[i].Name)
		}
	}
	if got := nextIndex(1, len(banners)); got != 0 {
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
	if got, err := bannerIndex("monochrome"); err != nil || got != 1 {
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
	if err := SaveBannerName(bannerPath, "monochrome"); err != nil {
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
	if got, err := LoadBannerName(bannerPath); err != nil || got != "monochrome" {
		t.Fatalf("Banner: %q, %v", got, err)
	}
	if got, err := LoadBannerAlignmentName(alignPath); err != nil || got != "right" {
		t.Fatalf("Ausrichtung: %q, %v", got, err)
	}
}

func TestLegacyBannerSettingsLoadNormalizedAndSaveStableName(t *testing.T) {
	cases := map[string]string{
		"wordmark-ansi": "ansi",
		"terminal-ansi": "ansi",
		"wordmark-mono": "monochrome",
		"terminal-mono": "monochrome",
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
	path := filepath.Join(t.TempDir(), "banner")
	if err := SaveBannerName(path, "wordmark-ansi"); err == nil {
		t.Fatal("Legacy-Modus darf nicht gespeichert werden")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("ungültiger Modus hat Datei erzeugt: %v", err)
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
