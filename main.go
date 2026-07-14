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
	excludes, err := LoadExcludes(filepath.Join(configDir, "flux", "exclude"))
	if err != nil {
		return err
	}

	hosts := FilterHosts(entries, excludes)
	if len(hosts) == 0 {
		return fmt.Errorf("keine anzeigbaren Hosts in %s (nach Wildcard- und Ausschlussfilter)", configPath)
	}

	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh nicht im PATH: %w", err)
	}

	selected, err := runTUI(hosts)
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
