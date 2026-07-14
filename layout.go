package main

import (
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
)

const (
	maxTUIWidth      = 100
	horizontalMargin = 2
	verticalMargin   = 1
	minBodyHeight    = 3
	borderHeight     = 2
	searchHeight     = 1
	bannerGapHeight  = 1
)

type tuiLayout struct {
	Width        int
	WindowHeight int
	BodyHeight   int
	FooterHeight int
	Banner       *BannerForm
}

func calculateTUILayout(screenWidth, screenHeight, naturalWidth, preferredBodyHeight int, footerText string, mode BannerMode) tuiLayout {
	widthCap := max(1, min(maxTUIWidth, screenWidth-2*horizontalMargin))
	desiredWidth := max(1, naturalWidth)
	for _, form := range mode.Family.Forms {
		formWidth := bannerFormWidth(form)
		if formWidth <= widthCap {
			desiredWidth = max(desiredWidth, formWidth)
			break
		}
	}
	width := min(widthCap, desiredWidth)

	footerHeight := wrappedLineCount(footerText, max(1, width-borderHeight))
	bodyHeight := max(minBodyHeight, preferredBodyHeight)
	availableWindowHeight := max(0, screenHeight-2*verticalMargin)
	fixedHeight := borderHeight + searchHeight
	windowHeight := min(availableWindowHeight, fixedHeight+bodyHeight+footerHeight)

	overflow := fixedHeight + bodyHeight + footerHeight - windowHeight
	if overflow > 0 {
		bodyReduction := min(overflow, bodyHeight-minBodyHeight)
		bodyHeight -= bodyReduction
		overflow -= bodyReduction

		footerReduction := min(overflow, footerHeight)
		footerHeight -= footerReduction
		overflow -= footerReduction

		// Tiny terminals keep fixed border/search rows where possible, then use any
		// remaining row for the body; only they may reduce it below the minimum.
		bodyHeight -= min(overflow, bodyHeight)
	}

	remainingHeight := screenHeight - 2*verticalMargin - windowHeight
	var banner *BannerForm
	for i := range mode.Family.Forms {
		form := &mode.Family.Forms[i]
		if bannerFormWidth(*form) <= width && bannerHeight(*form)+bannerGapHeight <= remainingHeight {
			banner = form
			break
		}
	}

	return tuiLayout{
		Width:        width,
		WindowHeight: windowHeight,
		BodyHeight:   bodyHeight,
		FooterHeight: footerHeight,
		Banner:       banner,
	}
}

func wrappedLineCount(text string, width int) int {
	return max(1, len(tview.WordWrap(text, width)))
}

func bannerFormWidth(form BannerForm) int {
	width := 0
	for _, row := range form.Rows {
		width = max(width, runewidth.StringWidth(row))
	}
	return width
}
