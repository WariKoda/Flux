package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestCalculateTUILayoutCapsWidthAndLeavesMargins(t *testing.T) {
	if got := calculateTUILayout(200, 40, 140, 10, "kurz").Width; got != 100 {
		t.Fatalf("100 erwartet, %d", got)
	}
	if got := calculateTUILayout(80, 40, 90, 10, "kurz").Width; got != 76 {
		t.Fatalf("76 erwartet, %d", got)
	}
}

func TestFooterWrapDoesNotChangeWidth(t *testing.T) {
	short := calculateTUILayout(90, 30, 60, 8, "kurz")
	long := calculateTUILayout(90, 30, 60, 8, strings.Repeat("lang ", 40))
	if short.Width != long.Width {
		t.Fatalf("Footer änderte Breite: %d != %d", short.Width, long.Width)
	}
	if long.FooterHeight <= short.FooterHeight {
		t.Fatalf("Footer brach nicht um")
	}
}

func TestConstrainedLayoutPreservesMinimumBodyAndAllocatesWrappedFooter(t *testing.T) {
	screenHeight := 2*verticalMargin + borderHeight + searchHeight + minBodyHeight + 4
	layout := calculateTUILayout(40, screenHeight, 20, 50, strings.Repeat("footer ", 100))

	if layout.BodyHeight < minBodyHeight {
		t.Fatalf("BodyHeight = %d, mindestens %d erwartet", layout.BodyHeight, minBodyHeight)
	}
	if layout.FooterHeight != 4 {
		t.Fatalf("FooterHeight = %d, 4 erwartet", layout.FooterHeight)
	}
	if got := borderHeight + searchHeight + layout.BodyHeight + layout.FooterHeight; got != layout.WindowHeight {
		t.Fatalf("Komponentenhöhe = %d, WindowHeight = %d", got, layout.WindowHeight)
	}
}

func TestTinyTerminalLayoutReconcilesWithoutNegativeSizes(t *testing.T) {
	for screenHeight := 0; screenHeight < 2*verticalMargin+borderHeight+searchHeight+minBodyHeight; screenHeight++ {
		t.Run(fmt.Sprintf("height_%d", screenHeight), func(t *testing.T) {
			layout := calculateTUILayout(40, screenHeight, 20, 50, strings.Repeat("footer ", 100))
			allocatedFixedHeight := min(borderHeight+searchHeight, layout.WindowHeight)
			wantBodyHeight := layout.WindowHeight - allocatedFixedHeight

			if layout.BodyHeight < 0 || layout.FooterHeight < 0 || layout.WindowHeight < 0 {
				t.Fatalf("negative Größe: %+v", layout)
			}
			if got := allocatedFixedHeight + layout.BodyHeight + layout.FooterHeight; got != layout.WindowHeight {
				t.Fatalf("Komponentenhöhe = %d, WindowHeight = %d", got, layout.WindowHeight)
			}
			if layout.BodyHeight != wantBodyHeight || layout.FooterHeight != 0 {
				t.Fatalf("Body/Footer = %d/%d, %d/0 erwartet", layout.BodyHeight, layout.FooterHeight, wantBodyHeight)
			}
			if layout.WindowHeight >= borderHeight+searchHeight+minBodyHeight && layout.BodyHeight < minBodyHeight {
				t.Fatalf("BodyHeight = %d trotz Platz für Minimum", layout.BodyHeight)
			}
			if layout.Banner != nil {
				t.Fatalf("Banner = %s, hidden erwartet", bannerName(layout.Banner))
			}
		})
	}
}

func TestBannerFormSelectionAtWidthAndHeightBoundaries(t *testing.T) {
	formWidth := bannerFormWidth(largeBanner)
	windowHeight := borderHeight + searchHeight + minBodyHeight + 1
	largeScreenHeight := 2*verticalMargin + windowHeight + bannerGapHeight + bannerHeight(largeBanner)
	compactScreenHeight := 2*verticalMargin + windowHeight + bannerGapHeight + bannerHeight(compactBanner)

	tests := []struct {
		name         string
		screenWidth  int
		screenHeight int
		wantBanner   *BannerForm
	}{
		{"large at exact width and height", formWidth + 2*horizontalMargin, largeScreenHeight, &largeBanner},
		{"compact one row below large boundary", formWidth + 2*horizontalMargin, largeScreenHeight - 1, &compactBanner},
		{"hidden one row below compact boundary", formWidth + 2*horizontalMargin, compactScreenHeight - 1, nil},
		{"hidden one column below form width", formWidth + 2*horizontalMargin - 1, largeScreenHeight, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := calculateTUILayout(tt.screenWidth, tt.screenHeight, 1, minBodyHeight, "kurz")
			if layout.Banner != tt.wantBanner {
				t.Fatalf("Banner = %v, erwartet %v", bannerName(layout.Banner), bannerName(tt.wantBanner))
			}
			if layout.BodyHeight < minBodyHeight {
				t.Fatalf("BodyHeight = %d, mindestens %d erwartet", layout.BodyHeight, minBodyHeight)
			}
		})
	}
}

func bannerName(form *BannerForm) string {
	if form == nil {
		return "hidden"
	}
	return form.Name
}
