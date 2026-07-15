package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestCalculateTUILayoutUsesAvailableWidthUpToCap(t *testing.T) {
	tests := []struct {
		name                            string
		screenWidth, naturalWidth, want int
	}{
		{"capped", 200, 1, 100},
		{"inside margins", 80, 1, 76},
		{"natural width ignored", 80, 140, 76},
		{"tiny", 3, 140, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateTUILayout(tt.screenWidth, 40, tt.naturalWidth, 3, "kurz", banners[len(banners)-1]).Width
			if got != tt.want {
				t.Fatalf("Width = %d, %d erwartet", got, tt.want)
			}
		})
	}
}

func TestFooterWrapIgnoresNaturalWidth(t *testing.T) {
	footer := strings.Repeat("footer ", 30)
	narrowNatural := calculateTUILayout(80, 40, 1, 3, footer, banners[len(banners)-1])
	wideNatural := calculateTUILayout(80, 40, 140, 3, footer, banners[len(banners)-1])
	if narrowNatural.FooterHeight != wideNatural.FooterHeight {
		t.Fatalf("FooterHeight = %d/%d bei gleicher Screenbreite", narrowNatural.FooterHeight, wideNatural.FooterHeight)
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
	layout := calculateTUILayout(40, screenHeight, 20, 50, strings.Repeat("footer ", 100), banners[len(banners)-1])

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
	form := blurVisionFamily.Forms[0]
	formWidth := bannerFormWidth(form)
	exactScreenHeight := 2*verticalMargin + 5 + bannerGapHeight + borderHeight + searchHeight + minBodyHeight

	tests := []struct {
		name         string
		screenWidth  int
		screenHeight int
		wantBanner   *BannerForm
	}{
		{"fixed form at exact width and height", formWidth + 2*horizontalMargin, exactScreenHeight, &banners[0].Family.Forms[0]},
		{"hidden one row below exact height", formWidth + 2*horizontalMargin, exactScreenHeight - 1, nil},
		{"hidden one column below form width", formWidth + 2*horizontalMargin - 1, exactScreenHeight, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := calculateTUILayout(tt.screenWidth, tt.screenHeight, 1, minBodyHeight, "kurz", banners[0])
			if layout.Banner != tt.wantBanner {
				t.Fatalf("Banner = %p (%s), erwartet %p (%s)", layout.Banner, bannerName(layout.Banner), tt.wantBanner, bannerName(tt.wantBanner))
			}
			if layout.BodyHeight < minBodyHeight {
				t.Fatalf("BodyHeight = %d, mindestens %d erwartet", layout.BodyHeight, minBodyHeight)
			}
			if tt.screenHeight == exactScreenHeight && tt.wantBanner != nil && layout.FooterHeight != 0 {
				t.Fatalf("FooterHeight = %d, 0 erwartet", layout.FooterHeight)
			}
		})
	}
}

func TestBannerFirstPreservesFixedFormBeforePreferredBody(t *testing.T) {
	minimumWindowHeight := borderHeight + searchHeight + minBodyHeight
	screenHeight := 2*verticalMargin + bannerHeight(blurVisionFamily.Forms[0]) + bannerGapHeight + minimumWindowHeight
	layout := calculateTUILayout(100, screenHeight, 1, 50, "kurz", banners[0])
	if layout.Banner != &banners[0].Family.Forms[0] || layout.BodyHeight != minBodyHeight || layout.FooterHeight != 0 {
		t.Fatalf("Layout = %+v, feste Form mit Body %d und Footer 0 erwartet", layout, minBodyHeight)
	}
}

func TestFooterShrinksBeforeBannerAndThenReceivesExtraRow(t *testing.T) {
	minimumWindowHeight := borderHeight + searchHeight + minBodyHeight
	exactHeight := 2*verticalMargin + bannerHeight(blurVisionFamily.Forms[0]) + bannerGapHeight + minimumWindowHeight
	footer := strings.Repeat("footer ", 100)
	exact := calculateTUILayout(100, exactHeight, 1, 50, footer, banners[0])
	if exact.Banner != &banners[0].Family.Forms[0] || exact.FooterHeight != 0 {
		t.Fatalf("exaktes Layout = %+v", exact)
	}
	extra := calculateTUILayout(100, exactHeight+1, 1, 50, footer, banners[0])
	if extra.Banner != &banners[0].Family.Forms[0] || extra.FooterHeight != 1 || extra.BodyHeight != minBodyHeight {
		t.Fatalf("Layout mit Zusatzzeile = %+v", extra)
	}
}

func TestLayoutUsesOnlySelectedBannerFamily(t *testing.T) {
	mode := banners[8]
	layout := calculateTUILayout(80, 30, 20, 3, "kurz", mode)
	if layout.Banner != &mode.Family.Forms[0] {
		t.Fatalf("Banner: %+v", layout.Banner)
	}
	if layout.Width != 76 {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := tt.mode.Family.Forms[0]
			formWidth := bannerFormWidth(form)
			windowHeight := borderHeight + searchHeight + minBodyHeight
			screenHeight := 2*verticalMargin + windowHeight + bannerGapHeight + bannerHeight(form)

			exact := calculateTUILayout(formWidth+2*horizontalMargin, screenHeight, 1, minBodyHeight, "kurz", tt.mode)
			if exact.Banner != &tt.mode.Family.Forms[0] {
				t.Fatalf("exact fit Banner = %p (%s), erwartet %p (%s)", exact.Banner, bannerName(exact.Banner), &tt.mode.Family.Forms[0], form.Name)
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
	layout := calculateTUILayout(200, 100, 20, minBodyHeight, "kurz", banners[len(banners)-1])
	if layout.Banner != nil {
		t.Fatalf("Banner = %s, hidden erwartet", bannerName(layout.Banner))
	}
	if layout.Width != 100 {
		t.Fatalf("Width = %d, 100 erwartet", layout.Width)
	}
}

func TestNoneBannerDoesNotFillUnusedHeight(t *testing.T) {
	footer := "kurz"
	layout := calculateTUILayout(80, 100, 1, 5, footer, banners[len(banners)-1])
	want := borderHeight + searchHeight + 5 + wrappedLineCount(footer, layout.Width-borderHeight)
	if layout.WindowHeight != want {
		t.Fatalf("WindowHeight = %d, %d erwartet", layout.WindowHeight, want)
	}
}

func TestLayoutHeightReconciliationAcrossCatalog(t *testing.T) {
	for screenHeight := 0; screenHeight <= 40; screenHeight++ {
		for modeIndex, mode := range banners {
			layout := calculateTUILayout(100, screenHeight, 1, 50, strings.Repeat("footer ", 100), mode)
			if layout.BodyHeight < 0 || layout.FooterHeight < 0 || layout.WindowHeight < 0 {
				t.Fatalf("height %d mode %d: negative Größe: %+v", screenHeight, modeIndex, layout)
			}
			bannerAndGap := 0
			if layout.Banner != nil {
				bannerAndGap = bannerHeight(*layout.Banner) + bannerGapHeight
			}
			if got := 2*verticalMargin + bannerAndGap + layout.WindowHeight; got > screenHeight && screenHeight >= 2*verticalMargin {
				t.Fatalf("height %d mode %d: Layout %d Zeilen größer als Screen: %+v", screenHeight, modeIndex, got, layout)
			}
		}
	}
}

func bannerName(form *BannerForm) string {
	if form == nil {
		return "hidden"
	}
	return form.Name
}
