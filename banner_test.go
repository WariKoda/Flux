package main

import "testing"

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
