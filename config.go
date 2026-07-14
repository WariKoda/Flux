package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// HostEntry ist ein Host-Block aus der ssh_config.
type HostEntry struct {
	// Alias ist der erste Nicht-Wildcard-Alias des Blocks und zugleich das
	// Verbindungsziel für ssh. Leer, wenn der Block nur aus Wildcards besteht.
	Alias string
	// Aliases enthält alle Aliase des Blocks in Original-Reihenfolge.
	Aliases  []string
	HostName string
	User     string
}

// ParseSSHConfig liest die ssh_config unter path und liefert alle Host-Blöcke.
// Include- und Match-Direktiven werden nicht unterstützt und führen zu einem
// Fehler, ebenso Aliase mit führendem '-' (wären als ssh-Ziel nicht nutzbar).
func ParseSSHConfig(path string) ([]HostEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ssh_config nicht lesbar: %w", err)
	}
	defer f.Close()

	var (
		entries []HostEntry
		current *HostEntry
		lineNo  int
	)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		keyword, value, err := splitDirective(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, lineNo, err)
		}

		switch strings.ToLower(keyword) {
		case "include", "match":
			return nil, fmt.Errorf("%s:%d: Direktive %q wird nicht unterstützt", path, lineNo, keyword)
		case "host":
			if current != nil {
				entries = append(entries, *current)
			}
			aliases := strings.Fields(value)
			if len(aliases) == 0 {
				return nil, fmt.Errorf("%s:%d: Host-Direktive ohne Alias", path, lineNo)
			}
			for i, a := range aliases {
				aliases[i] = unquote(a)
				if strings.HasPrefix(aliases[i], "-") {
					return nil, fmt.Errorf("%s:%d: Host-Alias %q mit führendem '-' wird nicht unterstützt", path, lineNo, aliases[i])
				}
			}
			current = &HostEntry{
				Alias:   firstNonWildcard(aliases),
				Aliases: aliases,
			}
		case "hostname":
			if current != nil {
				current.HostName = unquote(value)
			}
		case "user":
			if current != nil {
				current.User = unquote(value)
			}
		default:
			// Andere gültige ssh_config-Direktiven sind für die Anzeige irrelevant.
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ssh_config nicht lesbar: %w", err)
	}
	if current != nil {
		entries = append(entries, *current)
	}
	return entries, nil
}

// splitDirective zerlegt eine ssh_config-Zeile in Keyword und Wert.
// Erlaubte Trenner sind Whitespace oder '='.
func splitDirective(line string) (keyword, value string, err error) {
	idx := strings.IndexAny(line, " \t=")
	if idx < 0 {
		return "", "", fmt.Errorf("ungültige Zeile ohne Wert: %q", line)
	}
	keyword = line[:idx]
	value = strings.Trim(line[idx:], " \t=")
	if value == "" {
		return "", "", fmt.Errorf("Direktive %q ohne Wert", keyword)
	}
	return keyword, value, nil
}

func unquote(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return s[1 : len(s)-1]
	}
	return s
}

func isWildcard(alias string) bool {
	return strings.ContainsAny(alias, "*?")
}

func firstNonWildcard(aliases []string) string {
	for _, a := range aliases {
		if !isWildcard(a) {
			return a
		}
	}
	return ""
}

// FilterHosts entfernt reine Wildcard-Blöcke sowie Blöcke, bei denen einer der
// Aliase auf der Ausschlussliste steht.
func FilterHosts(entries []HostEntry, excludes []string) []HostEntry {
	excluded := make(map[string]bool, len(excludes))
	for _, e := range excludes {
		excluded[e] = true
	}
	var result []HostEntry
	for _, entry := range entries {
		if entry.Alias == "" {
			continue
		}
		skip := false
		for _, a := range entry.Aliases {
			if excluded[a] {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, entry)
		}
	}
	return result
}

// LoadExcludes liest die Ausschlussliste (ein Alias pro Zeile, '#'-Kommentare).
// Eine fehlende Datei ist gültig (keine Ausschlüsse); jede andere Lesestörung
// ist ein Fehler.
func LoadExcludes(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Ausschlussliste nicht lesbar: %w", err)
	}
	defer f.Close()

	var excludes []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		excludes = append(excludes, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Ausschlussliste nicht lesbar: %w", err)
	}
	return excludes, nil
}
