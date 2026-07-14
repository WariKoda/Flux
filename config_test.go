package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Alle Fixtures sind synthetisch — keine realen Hostnamen.

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseSSHConfigBasic(t *testing.T) {
	path := writeConfig(t, `
# Kommentar
Host alpha
    HostName alpha.example.test
    User alice

Host beta beta.alias
    User bob

Host *
    Compression yes
`)
	entries, err := ParseSSHConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("3 Host-Blöcke erwartet, %d erhalten", len(entries))
	}
	if entries[0].Alias != "alpha" || entries[0].HostName != "alpha.example.test" || entries[0].User != "alice" {
		t.Errorf("Block alpha falsch geparst: %+v", entries[0])
	}
	if entries[1].Alias != "beta" || len(entries[1].Aliases) != 2 || entries[1].Aliases[1] != "beta.alias" {
		t.Errorf("Mehrfach-Alias falsch geparst: %+v", entries[1])
	}
	if entries[2].Alias != "" {
		t.Errorf("Wildcard-Block darf kein Verbindungsziel haben: %+v", entries[2])
	}
}

func TestParseSSHConfigEqualsSeparatorAndQuotes(t *testing.T) {
	path := writeConfig(t, `
Host gamma
    HostName=gamma.example.test
    User "carol"
`)
	entries, err := ParseSSHConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].HostName != "gamma.example.test" || entries[0].User != "carol" {
		t.Errorf("'='-Trenner oder Quotes falsch geparst: %+v", entries[0])
	}
}

func TestParseSSHConfigIncludeFails(t *testing.T) {
	path := writeConfig(t, "Include ~/.ssh/other\nHost alpha\n    User alice\n")
	if _, err := ParseSSHConfig(path); err == nil {
		t.Fatal("Include muss einen Fehler liefern")
	}
}

func TestParseSSHConfigMatchFails(t *testing.T) {
	path := writeConfig(t, "Match host alpha\n    User alice\n")
	if _, err := ParseSSHConfig(path); err == nil {
		t.Fatal("Match muss einen Fehler liefern")
	}
}

func TestParseSSHConfigLeadingDashFails(t *testing.T) {
	path := writeConfig(t, "Host -evil\n    User alice\n")
	if _, err := ParseSSHConfig(path); err == nil {
		t.Fatal("Alias mit führendem '-' muss einen Fehler liefern")
	}
}

func TestParseSSHConfigInvalidLineFails(t *testing.T) {
	path := writeConfig(t, "Host alpha\nKaputteZeileOhneWert\n")
	if _, err := ParseSSHConfig(path); err == nil {
		t.Fatal("Zeile ohne Wert muss einen Fehler liefern")
	}
}

func TestParseSSHConfigMissingFileFails(t *testing.T) {
	if _, err := ParseSSHConfig(filepath.Join(t.TempDir(), "gibt-es-nicht")); err == nil {
		t.Fatal("fehlende Datei muss einen Fehler liefern")
	}
}

func TestFilterHosts(t *testing.T) {
	entries := []HostEntry{
		{Alias: "alpha", Aliases: []string{"alpha"}},
		{Alias: "", Aliases: []string{"*"}},
		{Alias: "web1", Aliases: []string{"web-*", "web1"}},
		{Alias: "git.example.test", Aliases: []string{"git.example.test"}},
		{Alias: "delta", Aliases: []string{"delta", "delta.alias"}},
	}
	got := FilterHosts(entries, []string{"git.example.test", "delta.alias"})
	if len(got) != 2 {
		t.Fatalf("2 Hosts erwartet, %d erhalten: %+v", len(got), got)
	}
	if got[0].Alias != "alpha" || got[1].Alias != "web1" {
		t.Errorf("falsche Filterung: %+v", got)
	}
}

func TestLoadExcludes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exclude")
	content := "# Kommentar\n\ngit.example.test\n  spaced.example.test  \n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	excludes, err := LoadExcludes(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(excludes) != 2 || excludes[0] != "git.example.test" || excludes[1] != "spaced.example.test" {
		t.Errorf("Ausschlussliste falsch geparst: %+v", excludes)
	}
}

func TestLoadExcludesMissingFileIsEmpty(t *testing.T) {
	excludes, err := LoadExcludes(filepath.Join(t.TempDir(), "gibt-es-nicht"))
	if err != nil {
		t.Fatal(err)
	}
	if excludes != nil {
		t.Errorf("fehlende Datei muss leere Liste liefern, erhalten: %+v", excludes)
	}
}

func TestHostDetail(t *testing.T) {
	cases := []struct {
		name string
		in   HostEntry
		want string
	}{
		{"user und hostname", HostEntry{Alias: "alpha", Aliases: []string{"alpha"}, HostName: "alpha.example.test", User: "alice"}, "alice@alpha.example.test"},
		{"nur hostname", HostEntry{Alias: "alpha", Aliases: []string{"alpha"}, HostName: "alpha.example.test"}, "alpha.example.test"},
		{"nur user", HostEntry{Alias: "alpha", Aliases: []string{"alpha"}, User: "alice"}, "alice@alpha"},
		{"mit weiteren aliasen", HostEntry{Alias: "beta", Aliases: []string{"beta", "beta.alias"}}, "auch: beta.alias"},
	}
	for _, c := range cases {
		if got := hostDetail(c.in); got != c.want {
			t.Errorf("%s: %q erwartet, %q erhalten", c.name, c.want, got)
		}
	}
}
