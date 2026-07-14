# Flux

Ein minimales Terminal-UI (TUI), das die SSH-Login-Ziele aus `~/.ssh/config`
auflistet. Auswahl per Tastatur **oder** Maus — danach ersetzt Flux den eigenen
Prozess durch `ssh -- <alias>`. Nach dem Ende der SSH-Sitzung bist du wieder in
deiner Shell.

## Bedienung

| Eingabe | Wirkung |
|---|---|
| `↑`/`↓`, `j`/`k`, `Tab`, `PgUp`/`PgDn`, `Home`/`End` | Navigation |
| `Enter` oder **Linksklick** | Mit dem gewählten Host verbinden |
| Mausrad | Scrollen |
| `q` oder `Esc` | Beenden ohne Verbindung |

## Build

```sh
go build -o flux .
```

Benötigt Go ≥ 1.26 und ein `ssh`-Binary im `PATH`.

## Was angezeigt wird

- Jeder `Host`-Block aus `~/.ssh/config` erscheint als Eintrag; der erste
  Nicht-Wildcard-Alias ist das Verbindungsziel, `HostName`/`User` und weitere
  Aliase werden als Zweitzeile angezeigt.
- Wildcard-Patterns (`*`, `?`) werden nie angezeigt (`Host *` ist ein
  Default-Block, kein Ziel).

## Ausschlussliste (optional)

Aliase, die nicht in der Liste erscheinen sollen (z. B. reine Git-Hosts),
kommen in `~/.config/flux/exclude` — ein Alias pro Zeile, `#` leitet
Kommentare ein:

```
# Git-Hosts ausblenden
git.example.test
```

Fehlt die Datei, wird nichts ausgeschlossen. Die Datei bleibt lokal auf deinem
System; sie gehört nicht ins Repository.

## Fehlerverhalten (bewusst strikt)

Flux bricht mit einer klaren Fehlermeldung ab statt still zu degradieren:

- `~/.ssh/config` fehlt oder ist unlesbar,
- die Config enthält `Include`- oder `Match`-Direktiven (nicht unterstützt),
- ein Host-Alias beginnt mit `-` (als ssh-Ziel nicht nutzbar),
- nach der Filterung bleibt kein anzeigbarer Host übrig,
- `ssh` ist nicht im `PATH`.

## Sicherheit

- Flux liest ausschließlich `~/.ssh/config` — keine Schlüssel, keine
  `known_hosts`, keine Netzwerkzugriffe, keine Schreibzugriffe.
- `ssh` wird ohne Shell-Zwischenschicht mit festen Argumenten gestartet
  (`ssh -- <alias>`); das `--` beendet die Optionsverarbeitung, ein Alias kann
  nie als ssh-Option interpretiert werden.

## Bekannte Einschränkungen

- `Include`/`Match` werden nicht unterstützt (harter Fehler statt
  unvollständiger Liste).
- Aliase mit Leerzeichen in Anführungszeichen (`Host "mein host"`) werden an
  den Leerzeichen getrennt.

## Lizenz

[MIT](LICENSE)
