# Flux

Ein minimales Terminal-UI (TUI), das die SSH-Login-Ziele aus `~/.ssh/config`
in einem zentrierten Fenster anzeigt — nach Servern gruppiert, mit einer
Fußzeile, die die Details des markierten Hosts zeigt. Auswahl per Tastatur
**oder** Maus — danach ersetzt Flux den eigenen Prozess durch
`ssh -- <alias>`. Nach dem Ende der SSH-Sitzung bist du wieder in deiner
Shell.

## Bedienung

**Einfach lostippen:** Jede Buchstaben-/Zifferntaste filtert die Liste sofort
(Suche über Alias, HostName, User, su-User und Zielordner); der erste Treffer
ist vorgewählt, `Enter` verbindet also direkt. `Backspace` löscht, `Esc`
leert die Suche.

| Eingabe | Wirkung |
|---|---|
| Tippen | Sofort filtern |
| `↑`/`↓`, `Home`/`End` | Navigation (Überschriften werden übersprungen) |
| `Enter` oder **Linksklick** | Mit dem gewählten Host verbinden |
| Mausrad | Scrollen |
| `Ctrl+E` | Filter-Ansicht öffnen/schließen (Einträge ein-/ausblenden) |
| `Ctrl+T` | Theme wechseln (wird gespeichert) |
| `Ctrl+B` | Banner-Stil wechseln (wird gespeichert) |
| `Ctrl+A` | Banner-Ausrichtung wechseln (wird gespeichert) |
| `Ctrl+O` | Optionen/Hilfe öffnen/schließen |
| `Esc` | Suche leeren → Filter-Ansicht schließen → beenden |

In der Filter-Ansicht schalten `Enter`, **Linksklick** oder die Leertaste den
markierten Eintrag um: `[x]` = sichtbar, `[ ]` = ausgeblendet. Jede Änderung
wird sofort gespeichert; die Tipp-Suche funktioniert auch hier.

## Build

```sh
go build -o flux .
```

Benötigt Go ≥ 1.26 und ein `ssh`-Binary im `PATH`.

## Was angezeigt wird

- Die Hosts sind nach Server (`HostName`) gruppiert; Überschriftszeilen sind
  nicht wählbar. Der erste Nicht-Wildcard-Alias eines `Host`-Blocks ist das
  Verbindungsziel. Bei einem `RemoteCommand` im su-Stil
  (`su … -l <user> -c 'cd <ordner> && …'`) steht der effektive User als
  `→ user` direkt in der Zeile.
- Die Fußzeile zeigt die vollen Details des markierten Hosts
  (`User@HostName:Port`, effektiver User und Ordner, weitere Aliase) sowie
  das aktive Theme.
- Themes: `dark` (Dunkel), `light` (Hell), `matrix` (Matrix) sowie die
  farbenblind-freundlichen `cb-dark` (Okabe-Ito Dunkel) und `cb-light`
  (Okabe-Ito Hell) auf Basis der [Okabe-Ito-Palette](https://davidmathlogic.com/colorblind/)
  (Color Universal Design; für Protanopie, Deuteranopie und Tritanopie
  unterscheidbar). Mit `Ctrl+T` durchschalten — der Anzeigename steht in der
  Fußzeile; die Wahl landet in `~/.config/flux/theme` und gilt beim nächsten
  Start wieder. Eine kaputte (leere/unbekannte) Theme-Datei ist ein harter
  Fehler.
- Wildcard-Patterns (`*`, `?`) werden nie angezeigt (`Host *` ist ein
  Default-Block, kein Ziel).

## Banner und Hilfe

Mit `Ctrl+B` wechselst du der Reihe nach durch sechs Banner-Designs, jeweils
als **Regenbogen** und **Monochrom**, sowie die Auswahl **Kein Banner**:

- **BlurVision** — die klassische FLUX-Wortmarke,
- **Single** — eine kompakte einzeilige Wortmarke,
- **ANSI Regular** — eine fünfzeilige Blockschrift,
- **Banner3** — eine siebenzeilige Rautenschrift,
- **ANSI Compact** — eine dreizeilige kompakte Blockschrift,
- **Terrace** — eine siebenzeilige Terrassenschrift.

Die Regenbogen-Varianten verwenden `rainbow3`: zwölf Farben laufen über die
sichtbaren Terminalspalten, wechseln bei mehrzeiligen Bannern alle zwei
Spalten und beginnen je Zeile leicht versetzt. Nur der einzeilige
**Single**-Banner durchläuft das volle Farbspektrum spaltenweise. Die
monochromen Varianten verwenden die Textfarbe des aktiven Themes und werden
bei einem Theme-Wechsel sofort neu gezeichnet. **Kein Banner** entfernt
Banner und Abstand vollständig.

Die Wahl wird als stabiler Bezeichner in `~/.config/flux/banner` gespeichert
und beim nächsten Start wiederhergestellt. Ältere gespeicherte Werte
`ansi`, `monochrome`, `wordmark-ansi`, `terminal-ansi`, `wordmark-mono` und
`terminal-mono` bleiben kompatibel und werden dem jeweils passenden
BlurVision-, Single- oder ANSI-Regular-Banner zugeordnet.

Mit `Ctrl+A` wechselst du zwischen den Ausrichtungen Links, Mitte und Rechts.
Flux speichert die Ausrichtung unverändert in
`~/.config/flux/banner-alignment`. Größe und Sichtbarkeit werden bei jedem
Zeichnen neu bestimmt. Nur **BlurVision** besitzt zwei Formen: Passt die
vollständige siebenzeilige Form samt Trennzeile nicht, versucht Flux die
vollständige fünfzeilige Form; passt auch sie nicht in Breite oder Höhe,
bleibt der Banner verborgen. Single, ANSI Regular, Banner3, ANSI Compact und
Terrace haben jeweils eine feste Form und werden vollständig verborgen,
sobald diese nicht passt. Banner werden nie abgeschnitten oder skaliert und
nach einer Terminal-Größenänderung sofort neu gewählt. Farbvariante und die
mit `Ctrl+A` gewählte Ausrichtung bleiben dabei erhalten.

Das zentrierte TUI bleibt höchstens 100 Spalten breit und lässt links wie
rechts mindestens zwei Terminalspalten frei. Lange Fußzeilen werden innerhalb
dieser Breite auf mehrere Zeilen umgebrochen, statt das Fenster zu verbreitern.

`Ctrl+O` öffnet die Optionen/Hilfe innerhalb desselben Fensters. `Ctrl+O`
oder `Esc` schließt sie wieder. Suchtext, Auswahl, Ansicht, Theme, Banner und
Ausrichtung bleiben beim Öffnen und Schließen unverändert. Solange die Hilfe
geöffnet ist, werden alle anderen Eingaben von den darunterliegenden
Ansichten ferngehalten; Navigationstasten können den Hilfetext scrollen.

## Ausschlussliste

Am einfachsten über die Filter-Ansicht (`Ctrl+E`) pflegen. Die Liste liegt in
`~/.config/flux/exclude` (ein Alias pro Zeile, `#` leitet Kommentare ein) und
kann auch von Hand bearbeitet werden — **Achtung**: Sobald du in der
Filter-Ansicht etwas umschaltest, schreibt Flux die Datei komplett neu;
eigene Kommentare gehen dabei verloren. Einträge, die zu keinem aktuellen
Host passen (z. B. entfernte Hosts), bleiben erhalten. Fehlt die Datei, wird
nichts ausgeschlossen. Die Datei bleibt lokal auf deinem System; sie gehört
nicht ins Repository.

Sind beim Start alle Hosts ausgeblendet, öffnet Flux direkt die
Filter-Ansicht statt mit einem Fehler abzubrechen.

## Fehlerverhalten (bewusst strikt)

Flux bricht mit einer klaren Fehlermeldung ab statt still zu degradieren:

- `~/.ssh/config` fehlt oder ist unlesbar,
- die Config enthält `Include`- oder `Match`-Direktiven (nicht unterstützt),
- ein Host-Alias beginnt mit `-` (als ssh-Ziel nicht nutzbar),
- die Config enthält ausschließlich Wildcard-Blöcke,
- `ssh` ist nicht im `PATH`,
- Theme-, Banner-, Banner-Ausrichtungs- oder Ausschlussdatei lassen sich nicht
  lesen/schreiben.

## Sicherheit

- Flux liest keine Schlüssel oder `known_hosts` und führt keine eigenen
  Netzwerkzugriffe aus. Schreibzugriffe beschränken sich auf seine Dateien
  unter `~/.config/flux/`.
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
