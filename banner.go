package main

import (
	"fmt"
	"os"
	"path/filepath"
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

func loadChoice(path, errorContext, defaultName string, validate func(string) error) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultName, nil
		}
		return "", fmt.Errorf("%s: %w", errorContext, err)
	}

	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", fmt.Errorf("%s: Datei ist leer", errorContext)
	}
	if err := validate(name); err != nil {
		return "", fmt.Errorf("%s: %w", errorContext, err)
	}
	return name, nil
}

func saveChoice(path, errorContext, name string, validate func(string) error) error {
	if err := validate(name); err != nil {
		return fmt.Errorf("%s: %w", errorContext, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("%s: %w", errorContext, err)
	}
	if err := os.WriteFile(path, []byte(name+"\n"), 0o600); err != nil {
		return fmt.Errorf("%s: %w", errorContext, err)
	}
	return nil
}

func LoadBannerName(path string) (string, error) {
	return loadChoice(path, "Banner-Datei nicht lesbar", banners[0].Name, func(name string) error {
		_, err := bannerIndex(name)
		return err
	})
}

func SaveBannerName(path, name string) error {
	return saveChoice(path, "Banner nicht speicherbar", name, func(name string) error {
		_, err := bannerIndex(name)
		return err
	})
}

func LoadBannerAlignmentName(path string) (string, error) {
	return loadChoice(path, "Banner-Ausrichtungsdatei nicht lesbar", bannerAlignments[0].Name, func(name string) error {
		_, err := bannerAlignmentIndex(name)
		return err
	})
}

func SaveBannerAlignmentName(path, name string) error {
	return saveChoice(path, "Banner-Ausrichtung nicht speicherbar", name, func(name string) error {
		_, err := bannerAlignmentIndex(name)
		return err
	})
}
