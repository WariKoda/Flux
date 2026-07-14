package main

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

type BannerColorMode int

const (
	bannerANSI BannerColorMode = iota
	bannerMonochrome
)

type Banner struct {
	Name, DisplayName string
	Rows              []string
	ColorMode         BannerColorMode
}

type BannerAlignment struct {
	Name, DisplayName string
	TViewAlign        int
}

var banners = []Banner{
	{"wordmark-ansi", "Wortmarke · ANSI", []string{"▓▒░ FLUX ░▒▓"}, bannerANSI},
	{"wordmark-mono", "Wortmarke · Monochrom", []string{"▓▒░ FLUX ░▒▓"}, bannerMonochrome},
	{"terminal-ansi", "Terminal · ANSI", []string{"█▀▀▀  █     █  █  ▀█▄█▀", "█▀▀   █     █  █    █", "█     █     █  █  ▄█▀█▄", "▀     ▀▀▀▀   ▀▀   ▀   ▀"}, bannerANSI},
	{"terminal-mono", "Terminal · Monochrom", []string{"█▀▀▀  █     █  █  ▀█▄█▀", "█▀▀   █     █  █    █", "█     █     █  █  ▄█▀█▄", "▀     ▀▀▀▀   ▀▀   ▀   ▀"}, bannerMonochrome},
}

var bannerAlignments = []BannerAlignment{
	{"left", "Links", tview.AlignLeft},
	{"center", "Mitte", tview.AlignCenter},
	{"right", "Rechts", tview.AlignRight},
}

func bannerIndex(name string) (int, error) {
	names := make([]string, len(banners))
	for i, banner := range banners {
		if banner.Name == name {
			return i, nil
		}
		names[i] = banner.Name
	}
	return 0, fmt.Errorf("unbekannter Banner %q (verfügbar: %s)", name, strings.Join(names, ", "))
}

func bannerAlignmentIndex(name string) (int, error) {
	names := make([]string, len(bannerAlignments))
	for i, alignment := range bannerAlignments {
		if alignment.Name == name {
			return i, nil
		}
		names[i] = alignment.Name
	}
	return 0, fmt.Errorf("unbekannte Banner-Ausrichtung %q (verfügbar: %s)", name, strings.Join(names, ", "))
}

func nextIndex(current, length int) int {
	if length <= 0 {
		panic("Zykluslänge muss positiv sein")
	}
	return (current + 1) % length
}
