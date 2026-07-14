package main

import (
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
