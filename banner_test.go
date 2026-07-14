package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBannerDefinitionsAndCycleOrder(t *testing.T) {
	want := []string{"wordmark-ansi", "wordmark-mono", "terminal-ansi", "terminal-mono"}
	if len(banners) != len(want) {
		t.Fatalf("%d Banner erwartet, %d erhalten", len(want), len(banners))
	}
	for i, name := range want {
		if banners[i].Name != name {
			t.Errorf("Banner %d: %q erwartet, %q erhalten", i, name, banners[i].Name)
		}
	}
	if got := nextIndex(3, len(banners)); got != 0 {
		t.Errorf("Wraparound: 0 erwartet, %d erhalten", got)
	}
}

func TestBannerAlignmentDefinitionsAndCycleOrder(t *testing.T) {
	want := []string{"left", "center", "right"}
	for i, name := range want {
		if bannerAlignments[i].Name != name {
			t.Errorf("Ausrichtung %d: %q erwartet, %q erhalten", i, name, bannerAlignments[i].Name)
		}
	}
	if got := nextIndex(2, len(bannerAlignments)); got != 0 {
		t.Errorf("Wraparound: 0 erwartet, %d erhalten", got)
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
	if err := SaveBannerName(bannerPath, "terminal-mono"); err != nil {
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
	if got, err := LoadBannerName(bannerPath); err != nil || got != "terminal-mono" {
		t.Fatalf("Banner: %q, %v", got, err)
	}
	if got, err := LoadBannerAlignmentName(alignPath); err != nil || got != "right" {
		t.Fatalf("Ausrichtung: %q, %v", got, err)
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
