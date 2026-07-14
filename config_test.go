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

func TestParseSSHConfigPortAndRemoteCommand(t *testing.T) {
	path := writeConfig(t, `
Host alpha
    HostName alpha.example.test
    User root
    Port 2222
    RemoteCommand su -P -s /bin/bash -l webuser -c 'cd /var/www/site && exec bash'
`)
	entries, err := ParseSSHConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].Port != "2222" {
		t.Errorf("Port falsch geparst: %+v", entries[0])
	}
	want := "su -P -s /bin/bash -l webuser -c 'cd /var/www/site && exec bash'"
	if entries[0].RemoteCommand != want {
		t.Errorf("RemoteCommand falsch geparst: %q", entries[0].RemoteCommand)
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
		{"mit port", HostEntry{Alias: "alpha", Aliases: []string{"alpha"}, HostName: "alpha.example.test", User: "root", Port: "2222"}, "root@alpha.example.test:2222"},
		{
			"remotecommand su-stil",
			HostEntry{Alias: "alpha", Aliases: []string{"alpha"}, HostName: "alpha.example.test", User: "root", Port: "2222", RemoteCommand: "su -P -s /bin/bash -l webuser -c 'cd /var/www/site && exec bash'"},
			"root@alpha.example.test:2222 → webuser · /var/www/site",
		},
		{
			"remotecommand unbekannt: roh anzeigen",
			HostEntry{Alias: "alpha", Aliases: []string{"alpha"}, HostName: "alpha.example.test", RemoteCommand: "tmux attach"},
			"alpha.example.test → tmux attach",
		},
	}
	for _, c := range cases {
		if got := hostDetail(c.in); got != c.want {
			t.Errorf("%s: %q erwartet, %q erhalten", c.name, c.want, got)
		}
	}
}

func TestExcludeState(t *testing.T) {
	entries := []HostEntry{
		{Alias: "alpha", Aliases: []string{"alpha"}},
		{Alias: "beta", Aliases: []string{"beta", "beta.alias"}},
		{Alias: "gamma", Aliases: []string{"gamma"}},
	}
	// "beta.alias" trifft den beta-Block über den Zweit-Alias,
	// "verwaist.example.test" gehört zu keinem Block und muss erhalten bleiben.
	state := NewExcludeState(entries, []string{"beta.alias", "verwaist.example.test"})
	if !state.Excluded["beta"] {
		t.Error("beta muss über den Zweit-Alias ausgeschlossen sein")
	}
	if len(state.Unknown) != 1 || state.Unknown[0] != "verwaist.example.test" {
		t.Errorf("unbekannter Eintrag muss erhalten bleiben: %+v", state.Unknown)
	}

	visible := state.Visible(entries)
	if len(visible) != 2 || visible[0].Alias != "alpha" || visible[1].Alias != "gamma" {
		t.Errorf("Visible falsch: %+v", visible)
	}

	state.Excluded["alpha"] = true
	list := state.List(entries)
	want := []string{"alpha", "beta", "verwaist.example.test"}
	if len(list) != len(want) {
		t.Fatalf("List falsch: %+v", list)
	}
	for i := range want {
		if list[i] != want[i] {
			t.Errorf("List[%d]: %q erwartet, %q erhalten", i, want[i], list[i])
		}
	}
}

func TestSaveExcludesRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "exclude")
	if err := SaveExcludes(path, []string{"alpha", "beta"}); err != nil {
		t.Fatal(err)
	}
	got, err := LoadExcludes(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Errorf("Roundtrip falsch: %+v", got)
	}
}

func TestThemesHaveUniqueNamesAndDisplayNames(t *testing.T) {
	seen := map[string]bool{}
	for _, th := range themes {
		if th.Name == "" || th.DisplayName == "" {
			t.Errorf("Theme ohne Name/DisplayName: %+v", th)
		}
		if seen[th.Name] {
			t.Errorf("doppelter Theme-Name %q", th.Name)
		}
		seen[th.Name] = true
	}
}

func TestGroupHosts(t *testing.T) {
	hosts := []HostEntry{
		{Alias: "a1", Aliases: []string{"a1"}, HostName: "srv-a.example.test"},
		{Alias: "b1", Aliases: []string{"b1"}, HostName: "srv-b.example.test"},
		{Alias: "a2", Aliases: []string{"a2"}, HostName: "srv-a.example.test"},
		{Alias: "solo", Aliases: []string{"solo"}},
	}
	groups := groupHosts(hosts)
	if len(groups) != 3 {
		t.Fatalf("3 Gruppen erwartet, %d erhalten: %+v", len(groups), groups)
	}
	if groups[0].Server != "srv-a.example.test" || len(groups[0].Hosts) != 2 {
		t.Errorf("Gruppe 0 falsch: %+v", groups[0])
	}
	if groups[1].Server != "srv-b.example.test" || len(groups[1].Hosts) != 1 {
		t.Errorf("Gruppe 1 falsch: %+v", groups[1])
	}
	if groups[2].Server != "solo" {
		t.Errorf("Fallback auf Alias fehlt: %+v", groups[2])
	}
}

func TestThemeRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "theme")
	if err := SaveThemeName(path, "light"); err != nil {
		t.Fatal(err)
	}
	name, err := LoadThemeName(path)
	if err != nil {
		t.Fatal(err)
	}
	if name != "light" {
		t.Errorf("light erwartet, %q erhalten", name)
	}
	if _, err := themeIndex(name); err != nil {
		t.Errorf("gespeichertes Theme muss auflösbar sein: %v", err)
	}
}

func TestLoadThemeNameMissingFileIsDefault(t *testing.T) {
	name, err := LoadThemeName(filepath.Join(t.TempDir(), "gibt-es-nicht"))
	if err != nil {
		t.Fatal(err)
	}
	if name != themes[0].Name {
		t.Errorf("Default-Theme %q erwartet, %q erhalten", themes[0].Name, name)
	}
}

func TestLoadThemeNameEmptyFileFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme")
	if err := os.WriteFile(path, []byte("  \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadThemeName(path); err == nil {
		t.Fatal("leere Theme-Datei muss einen Fehler liefern")
	}
}

func TestThemeIndexUnknownFails(t *testing.T) {
	if _, err := themeIndex("gibt-es-nicht"); err == nil {
		t.Fatal("unbekanntes Theme muss einen Fehler liefern")
	}
}

func TestParseRemoteTarget(t *testing.T) {
	user, dir := parseRemoteTarget("su -P -s /bin/bash -l webuser -c 'cd /var/www/site && exec bash'")
	if user != "webuser" || dir != "/var/www/site" {
		t.Errorf("su-Stil falsch extrahiert: user=%q dir=%q", user, dir)
	}
	user, dir = parseRemoteTarget("tmux attach")
	if user != "" || dir != "" {
		t.Errorf("unbekanntes Kommando muss leer extrahieren: user=%q dir=%q", user, dir)
	}
}
