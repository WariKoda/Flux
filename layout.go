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
	_ = naturalWidth
	width := max(1, min(maxTUIWidth, screenWidth-2*horizontalMargin))
	availableHeight := max(0, screenHeight-2*verticalMargin)
	fixedHeight := borderHeight + searchHeight
	minimumWindowHeight := fixedHeight + minBodyHeight
	bannerAndGap := 0
	var banner *BannerForm
	for i := range mode.Family.Forms {
		form := &mode.Family.Forms[i]
		formAndGap := bannerHeight(*form) + bannerGapHeight
		if bannerFormWidth(*form) <= width && formAndGap+minimumWindowHeight <= availableHeight {
			banner = form
			bannerAndGap = formAndGap
			break
		}
	}

	windowCapacity := availableHeight - bannerAndGap
	footerHeight := 0
	bodyHeight := 0
	windowHeight := windowCapacity
	if windowCapacity < minimumWindowHeight {
		allocatedFixedHeight := min(fixedHeight, windowCapacity)
		bodyHeight = windowCapacity - allocatedFixedHeight
	} else {
		bodyHeight = minBodyHeight
		remainingHeight := windowCapacity - minimumWindowHeight
		wrappedFooterHeight := wrappedLineCount(footerText, max(1, width-borderHeight))
		footerHeight = min(wrappedFooterHeight, remainingHeight)
		remainingHeight -= footerHeight
		preferredBodyHeight = max(minBodyHeight, preferredBodyHeight)
		bodyHeight += min(preferredBodyHeight-minBodyHeight, remainingHeight)
		windowHeight = fixedHeight + bodyHeight + footerHeight
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
