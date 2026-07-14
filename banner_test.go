package main

import (
	"os"
	"path/filepath"
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
		name  string
		value string
		load  func(string) (string, error)
	}{
		{"leerer Banner", " \n", LoadBannerName},
		{"unbekannter Banner", "wat\n", LoadBannerName},
		{"leere Ausrichtung", "\n", LoadBannerAlignmentName},
		{"unbekannte Ausrichtung", "diagonal\n", LoadBannerAlignmentName},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "setting")
			if err := os.WriteFile(path, []byte(tc.value), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := tc.load(path); err == nil {
				t.Fatal("Fehler erwartet")
			}
		})
	}
}
