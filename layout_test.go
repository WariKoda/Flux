package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestCalculateTUILayoutCapsWidthAndLeavesMargins(t *testing.T) {
	if got := calculateTUILayout(200, 40, 140, 10, "kurz", banners[0]).Width; got != 100 {
		t.Fatalf("100 erwartet, %d", got)
	}
	if got := calculateTUILayout(80, 40, 90, 10, "kurz", banners[0]).Width; got != 76 {
		t.Fatalf("76 erwartet, %d", got)
	}
}

func TestFooterWrapDoesNotChangeWidth(t *testing.T) {
	short := calculateTUILayout(90, 30, 60, 8, "kurz", banners[0])
	long := calculateTUILayout(90, 30, 60, 8, strings.Repeat("lang ", 40), banners[0])
	if short.Width != long.Width {
		t.Fatalf("Footer änderte Breite: %d != %d", short.Width, long.Width)
	}
	if long.FooterHeight <= short.FooterHeight {
		t.Fatalf("Footer brach nicht um")
	}
}

func TestConstrainedLayoutPreservesMinimumBodyAndAllocatesWrappedFooter(t *testing.T) {
	screenHeight := 2*verticalMargin + borderHeight + searchHeight + minBodyHeight + 4
	layout := calculateTUILayout(40, screenHeight, 20, 50, strings.Repeat("footer ", 100), banners[12])

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
			layout := calculateTUILayout(40, screenHeight, 20, 50, strings.Repeat("footer ", 100), banners[0])
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
		wantBanner   string
	}{
		{"large at exact width and height", formWidth + 2*horizontalMargin, largeScreenHeight, largeBanner.Name},
		{"compact one row below large boundary", formWidth + 2*horizontalMargin, largeScreenHeight - 1, compactBanner.Name},
		{"hidden one row below compact boundary", formWidth + 2*horizontalMargin, compactScreenHeight - 1, "hidden"},
		{"hidden one column below form width", formWidth + 2*horizontalMargin - 1, largeScreenHeight, "hidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := calculateTUILayout(tt.screenWidth, tt.screenHeight, 1, minBodyHeight, "kurz", banners[0])
			if got := bannerName(layout.Banner); got != tt.wantBanner {
				t.Fatalf("Banner = %v, erwartet %v", got, tt.wantBanner)
			}
			if layout.BodyHeight < minBodyHeight {
				t.Fatalf("BodyHeight = %d, mindestens %d erwartet", layout.BodyHeight, minBodyHeight)
			}
		})
	}
}

func TestLayoutUsesOnlySelectedBannerFamily(t *testing.T) {
	mode := banners[8]
	layout := calculateTUILayout(80, 30, 20, 3, "kurz", mode)
	if layout.Banner == nil || layout.Banner.Name != "compact" {
		t.Fatalf("Banner: %+v", layout.Banner)
	}
	if layout.Width != max(20, bannerFormWidth(mode.Family.Forms[0])) {
		t.Fatalf("Width: %d", layout.Width)
	}
}

func TestSelectedBannerFamilyBoundaries(t *testing.T) {
	tests := []struct {
		name string
		mode BannerMode
	}{
		{"single", banners[2]},
		{"ansi regular", banners[4]},
		{"banner3", banners[6]},
		{"ansi compact", banners[8]},
		{"terrace", banners[10]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := tt.mode.Family.Forms[0]
			formWidth := bannerFormWidth(form)
			windowHeight := borderHeight + searchHeight + minBodyHeight + wrappedLineCount("kurz", max(1, formWidth-borderHeight))
			screenHeight := 2*verticalMargin + windowHeight + bannerGapHeight + bannerHeight(form)

			exact := calculateTUILayout(formWidth+2*horizontalMargin, screenHeight, 1, minBodyHeight, "kurz", tt.mode)
			if exact.Banner == nil || exact.Banner.Name != form.Name {
				t.Fatalf("exact fit Banner = %v, erwartet %s", bannerName(exact.Banner), form.Name)
			}
			if got := calculateTUILayout(formWidth+2*horizontalMargin-1, screenHeight, 1, minBodyHeight, "kurz", tt.mode).Banner; got != nil {
				t.Fatalf("one column below Banner = %s, hidden erwartet", bannerName(got))
			}
			if got := calculateTUILayout(formWidth+2*horizontalMargin, screenHeight-1, 1, minBodyHeight, "kurz", tt.mode).Banner; got != nil {
				t.Fatalf("one row below Banner = %s, hidden erwartet", bannerName(got))
			}
		})
	}
}

func TestNoneBannerNeverSelectsOrWidens(t *testing.T) {
	layout := calculateTUILayout(200, 100, 20, minBodyHeight, "kurz", banners[12])
	if layout.Banner != nil {
		t.Fatalf("Banner = %s, hidden erwartet", bannerName(layout.Banner))
	}
	if layout.Width != 20 {
		t.Fatalf("Width = %d, 20 erwartet", layout.Width)
	}
}

func bannerName(form *BannerForm) string {
	if form == nil {
		return "hidden"
	}
	return form.Name
}
