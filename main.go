// Flux ist ein Terminal-UI, das die SSH-Login-Ziele aus ~/.ssh/config anzeigt
// und den eigenen Prozess nach der Auswahl durch `ssh -- <alias>` ersetzt.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "flux: "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Home-Verzeichnis nicht bestimmbar: %w", err)
	}
	configPath := filepath.Join(home, ".ssh", "config")
	entries, err := ParseSSHConfig(configPath)
	if err != nil {
		return err
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("Config-Verzeichnis nicht bestimmbar: %w", err)
	}
	excludePath := filepath.Join(configDir, "flux", "exclude")
	excludes, err := LoadExcludes(excludePath)
	if err != nil {
		return err
	}
	themePath := filepath.Join(configDir, "flux", "theme")
	themeName, err := LoadThemeName(themePath)
	if err != nil {
		return err
	}
	bannerPath := filepath.Join(configDir, "flux", "banner")
	bannerName, err := LoadBannerName(bannerPath)
	if err != nil {
		return err
	}
	alignmentPath := filepath.Join(configDir, "flux", "banner-alignment")
	alignmentName, err := LoadBannerAlignmentName(alignmentPath)
	if err != nil {
		return err
	}

	// Wildcard-Blöcke sind nie anzeigbar; Ausschlüsse verwaltet die
	// Filter-Sub-UI in runTUI selbst.
	candidates := FilterHosts(entries, nil)
	if len(candidates) == 0 {
		return fmt.Errorf("keine anzeigbaren Hosts in %s (nur Wildcard-Blöcke)", configPath)
	}

	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh nicht im PATH: %w", err)
	}

	selected, err := runTUI(candidates, excludes, excludePath, themeName, themePath, bannerName, bannerPath, alignmentName, alignmentPath)
	if err != nil {
		return err
	}
	if selected == "" {
		// Nutzer hat bewusst ohne Auswahl beendet.
		return nil
	}

	// '--' beendet die Optionsverarbeitung von ssh, damit ein Alias nie als
	// Option interpretiert werden kann. Exec ersetzt den Flux-Prozess; bei
	// Erfolg kehrt der Aufruf nicht zurück.
	if err := syscall.Exec(sshPath, []string{"ssh", "--", selected}, os.Environ()); err != nil {
		return fmt.Errorf("ssh konnte nicht gestartet werden: %w", err)
	}
	return nil
}
