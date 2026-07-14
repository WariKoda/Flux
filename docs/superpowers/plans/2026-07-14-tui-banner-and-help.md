# Configurable TUI Banner and Help Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add four persisted, theme-aware banner styles with configurable alignment and a state-preserving in-TUI help view.

**Architecture:** Keep banner definitions, persistence, coloring, and sizing in a focused `banner.go` unit with pure helpers that can be tested without a terminal. Extend the existing `tview` composition with a banner `TextView`, a `Pages` body for host/help switching, and small state-transition helpers; the current table population and SSH selection flow remain intact.

**Tech Stack:** Go 1.25, `github.com/rivo/tview` v0.42.0, `github.com/gdamore/tcell/v2` v2.8.1, standard `testing` package.

## Global Constraints

- Banner styles cycle in this exact order: wordmark ANSI, wordmark monochrome, four-line ANSI, four-line monochrome.
- Alignments cycle in this exact order: left, center, right.
- `Ctrl+B`, `Ctrl+A`, and `Ctrl+T` persist banner, alignment, and theme immediately; write failures stop the TUI and are returned.
- ANSI colors are fixed; monochrome banner color is the active theme's existing `Header` color.
- Banner alignment is relative to the calculated TUI width, with one blank row between banner and window.
- Hide the complete banner when the screen is too short; never clip it or mutate the saved setting.
- `Ctrl+O` opens options/help; `Ctrl+O` or `Esc` closes it; all other input is consumed while help is open.
- Missing setting files select defaults; empty, unreadable, or unknown settings fail explicitly.
- Preserve all existing filtering, selection, mouse, SSH, exclude-list, and theme behavior.
- Do not add dependencies.

---

## File Map

- Create `banner.go`: banner/alignment definitions, strict setting persistence, ANSI text generation, alignment and visibility calculations.
- Create `banner_test.go`: unit tests for definitions, cycling, persistence, rendering, alignment, and responsive visibility.
- Create `tui_test.go`: focused tests for help content/state transitions and status/title text.
- Modify `theme.go`: expose no new theme data; reuse `Theme.Header` as the monochrome banner color.
- Modify `tui.go`: compose banner and help primitives, wire shortcuts, responsive layout, titles, footer, and save errors.
- Modify `main.go`: load banner/alignment paths and pass settings into `runTUI`.
- Modify `README.md`: document banner variants, alignment, help, shortcuts, persistence, and small-terminal behavior.

### Task 1: Banner and Alignment Definitions

**Files:**
- Create: `banner.go`
- Create: `banner_test.go`

**Interfaces:**
- Produces: `Banner`, `BannerColorMode`, `BannerAlignment`, `banners`, `bannerAlignments`, `bannerIndex(string) (int, error)`, `bannerAlignmentIndex(string) (int, error)`, `nextIndex(int, int) int`.
- Consumes: `Theme.Header` from `theme.go` in later tasks only.

- [ ] **Step 1: Write failing definition and cycle tests**

```go
func TestBannerDefinitionsAndCycleOrder(t *testing.T) {
    want := []string{"wordmark-ansi", "wordmark-mono", "terminal-ansi", "terminal-mono"}
    if len(banners) != len(want) { t.Fatalf("%d Banner erwartet, %d erhalten", len(want), len(banners)) }
    for i, name := range want {
        if banners[i].Name != name { t.Errorf("Banner %d: %q erwartet, %q erhalten", i, name, banners[i].Name) }
    }
    if got := nextIndex(3, len(banners)); got != 0 { t.Errorf("Wraparound: 0 erwartet, %d erhalten", got) }
}

func TestBannerAlignmentDefinitionsAndCycleOrder(t *testing.T) {
    want := []string{"left", "center", "right"}
    for i, name := range want {
        if bannerAlignments[i].Name != name { t.Errorf("Ausrichtung %d: %q erwartet, %q erhalten", i, name, bannerAlignments[i].Name) }
    }
    if got := nextIndex(2, len(bannerAlignments)); got != 0 { t.Errorf("Wraparound: 0 erwartet, %d erhalten", got) }
}

func TestUnknownBannerAndAlignmentFail(t *testing.T) {
    if _, err := bannerIndex("unknown"); err == nil { t.Fatal("unbekannter Banner muss fehlschlagen") }
    if _, err := bannerAlignmentIndex("unknown"); err == nil { t.Fatal("unbekannte Ausrichtung muss fehlschlagen") }
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestBannerDefinitions|TestBannerAlignmentDefinitions|TestUnknownBanner'`

Expected: build failure because `banners`, `bannerAlignments`, and index helpers do not exist.

- [ ] **Step 3: Implement the minimal definitions**

```go
type BannerColorMode int
const (
    bannerANSI BannerColorMode = iota
    bannerMonochrome
)

type Banner struct {
    Name, DisplayName string
    Rows []string
    ColorMode BannerColorMode
}

type BannerAlignment struct {
    Name, DisplayName string
    TViewAlign int
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
```

Implement both lookup helpers with errors listing valid identifiers, following `themeIndex`, and implement `nextIndex` as `(current+1)%length` with a panic for `length <= 0` because an empty cycle is a programmer error.

- [ ] **Step 4: Run tests and verify GREEN**

Run: `go test ./... -run 'TestBannerDefinitions|TestBannerAlignmentDefinitions|TestUnknownBanner'`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add banner.go banner_test.go
git commit -m "feat: define configurable TUI banners"
```

### Task 2: Strict Banner Setting Persistence

**Files:**
- Modify: `banner.go`
- Modify: `banner_test.go`

**Interfaces:**
- Produces: `LoadBannerName(path string) (string, error)`, `SaveBannerName(path, name string) error`, `LoadBannerAlignmentName(path string) (string, error)`, `SaveBannerAlignmentName(path, name string) error`.
- Consumes: `banners[0].Name`, `bannerAlignments[0].Name`, and lookup helpers from Task 1.

- [ ] **Step 1: Write failing persistence tests**

```go
func TestBannerSettingsRoundTrip(t *testing.T) {
    dir := t.TempDir()
    bannerPath := filepath.Join(dir, "nested", "banner")
    alignPath := filepath.Join(dir, "nested", "banner-alignment")
    if err := SaveBannerName(bannerPath, "terminal-mono"); err != nil { t.Fatal(err) }
    if err := SaveBannerAlignmentName(alignPath, "right"); err != nil { t.Fatal(err) }
    if got, err := LoadBannerName(bannerPath); err != nil || got != "terminal-mono" { t.Fatalf("Banner: %q, %v", got, err) }
    if got, err := LoadBannerAlignmentName(alignPath); err != nil || got != "right" { t.Fatalf("Ausrichtung: %q, %v", got, err) }
}

func TestMissingBannerSettingsUseDefaults(t *testing.T) {
    dir := t.TempDir()
    if got, err := LoadBannerName(filepath.Join(dir, "banner")); err != nil || got != banners[0].Name { t.Fatalf("Banner: %q, %v", got, err) }
    if got, err := LoadBannerAlignmentName(filepath.Join(dir, "align")); err != nil || got != bannerAlignments[0].Name { t.Fatalf("Ausrichtung: %q, %v", got, err) }
}

func TestInvalidBannerSettingsFail(t *testing.T) {
    for _, tc := range []struct{name, value string; load func(string)(string,error)}{
        {"leerer Banner", " \n", LoadBannerName}, {"unbekannter Banner", "wat\n", LoadBannerName},
        {"leere Ausrichtung", "\n", LoadBannerAlignmentName}, {"unbekannte Ausrichtung", "diagonal\n", LoadBannerAlignmentName},
    } {
        t.Run(tc.name, func(t *testing.T) {
            path := filepath.Join(t.TempDir(), "setting")
            if err := os.WriteFile(path, []byte(tc.value), 0o600); err != nil { t.Fatal(err) }
            if _, err := tc.load(path); err == nil { t.Fatal("Fehler erwartet") }
        })
    }
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestBannerSettings|TestMissingBannerSettings|TestInvalidBannerSettings'`

Expected: build failure because persistence functions do not exist.

- [ ] **Step 3: Implement strict load/save helpers**

Add private helpers to avoid duplicating file I/O while retaining setting-specific German error messages:

```go
func loadChoice(path, label, defaultName string, validate func(string) error) (string, error)
func saveChoice(path, label, name string, validate func(string) error) error
```

`loadChoice` returns the default only for `os.IsNotExist`, trims whitespace, rejects empty content, and calls the relevant index helper through the validator. `saveChoice` validates before writing, creates the parent with `0o700`, and writes `name+"\n"` with `0o600`. Wrap errors as `Banner-Datei nicht lesbar`, `Banner nicht speicherbar`, `Banner-Ausrichtungsdatei nicht lesbar`, or `Banner-Ausrichtung nicht speicherbar`.

- [ ] **Step 4: Run persistence and full tests**

Run: `go test ./... -run 'TestBannerSettings|TestMissingBannerSettings|TestInvalidBannerSettings'`

Expected: PASS.

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add banner.go banner_test.go
git commit -m "feat: persist banner preferences"
```

### Task 3: Banner Rendering and Responsive Layout Helpers

**Files:**
- Modify: `banner.go`
- Modify: `banner_test.go`

**Interfaces:**
- Produces: `renderBanner(Banner, Theme) string`, `bannerHeight(Banner) int`, `bannerVisible(screenHeight, tuiHeight int, Banner) bool`, `alignedBannerText(Banner, width int, BannerAlignment, Theme) string`.
- Consumes: `Theme.Header`, `tcell.Color.Hex()`, and Task 1 definitions.

- [ ] **Step 1: Write failing renderer tests**

```go
func TestMonochromeBannerUsesThemeHeader(t *testing.T) {
    th := Theme{Header: tcell.NewHexColor(0x123456)}
    got := renderBanner(banners[1], th)
    if !strings.Contains(got, "[#123456]") { t.Fatalf("Theme-Farbe fehlt: %q", got) }
    if strings.Contains(got, "[#ff5555]") { t.Fatalf("ANSI-Farbe in Monochrom-Banner: %q", got) }
}

func TestANSIBannerIsThemeIndependent(t *testing.T) {
    a := renderBanner(banners[0], Theme{Header:tcell.ColorRed})
    b := renderBanner(banners[0], Theme{Header:tcell.ColorBlue})
    if a != b { t.Fatalf("ANSI-Banner darf sich mit Theme nicht ändern") }
    if !strings.Contains(a, "[#ff5555]") || !strings.Contains(a, "[#8be9fd]") { t.Fatalf("ANSI-Palette fehlt: %q", a) }
}

func TestBannerVisibilityRequiresWholeBannerAndGap(t *testing.T) {
    b := banners[2]
    if !bannerVisible(15, 10, b) { t.Fatal("4 Zeilen + Abstand müssen exakt passen") }
    if bannerVisible(14, 10, b) { t.Fatal("eine Zeile zu wenig muss Banner ausblenden") }
}

func TestAlignedBannerText(t *testing.T) {
    b := Banner{Rows: []string{"FLUX"}, ColorMode:bannerMonochrome}
    th := Theme{Header:tcell.ColorGreen}
    if got := stripTags(alignedBannerText(b, 8, bannerAlignments[0], th)); got != "FLUX" { t.Errorf("links: %q", got) }
    if got := stripTags(alignedBannerText(b, 8, bannerAlignments[1], th)); got != "  FLUX" { t.Errorf("mitte: %q", got) }
    if got := stripTags(alignedBannerText(b, 8, bannerAlignments[2], th)); got != "    FLUX" { t.Errorf("rechts: %q", got) }
}
```

Use a test-only `stripTags` regexp to remove tview color tags before comparing padding.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestMonochromeBanner|TestANSIBanner|TestBannerVisibility|TestAlignedBannerText'`

Expected: build failure because rendering/layout helpers do not exist.

- [ ] **Step 3: Implement rendering and layout helpers**

Use this fixed ANSI palette in order and assign colors by printable rune position across each row:

```go
var bannerANSIColors = []string{"#ff5555", "#f1fa8c", "#50fa7b", "#8be9fd", "#6272a4", "#bd93f9", "#ff79c6", "#ffb86c"}
```

For monochrome output prepend one `[#rrggbb]` tag per row. For ANSI output choose `palette[position*len(palette)/max(1,rowWidth)]`, emitting a new tag only when the color changes. Preserve spaces and newlines exactly. `alignedBannerText` pads each rendered row before color-tag insertion, so tags never affect width. Return unpadded rows when `width <= rowWidth`. `bannerVisible` returns `screenHeight >= tuiHeight+bannerHeight(b)+1`.

- [ ] **Step 4: Run renderer and full tests**

Run: `go test ./... -run 'TestMonochromeBanner|TestANSIBanner|TestBannerVisibility|TestAlignedBannerText'`

Expected: PASS.

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add banner.go banner_test.go
git commit -m "feat: render responsive theme-aware banners"
```

### Task 4: Help Model and State-Preserving Input Rules

**Files:**
- Create: `tui_test.go`
- Modify: `tui.go`

**Interfaces:**
- Produces: `helpText() string`, `tuiViewState{HelpVisible bool}`, `handleHelpKey(*tcell.EventKey, *tuiViewState) (consumed bool)`.
- Consumes: `themes`, `banners`, and `bannerAlignments` to generate complete option lists.

- [ ] **Step 1: Write failing help tests**

```go
func TestHelpTextListsCommandsAndOptions(t *testing.T) {
    got := helpText()
    for _, want := range []string{"Ctrl+E", "Ctrl+T", "Ctrl+B", "Ctrl+A", "Ctrl+O", "Esc", "Wortmarke · ANSI", "Terminal · Monochrom", "Links", "Mitte", "Rechts", "Okabe-Ito Dunkel"} {
        if !strings.Contains(got, want) { t.Errorf("Hilfe enthält %q nicht", want) }
    }
}

func TestHelpInputRules(t *testing.T) {
    state := tuiViewState{}
    if !handleHelpKey(tcell.NewEventKey(tcell.KeyCtrlO, 0, 0), &state) || !state.HelpVisible { t.Fatal("Ctrl+O muss Hilfe öffnen") }
    if !handleHelpKey(tcell.NewEventKey(tcell.KeyRune, 'x', 0), &state) || !state.HelpVisible { t.Fatal("Eingabe muss konsumiert werden") }
    if !handleHelpKey(tcell.NewEventKey(tcell.KeyEscape, 0, 0), &state) || state.HelpVisible { t.Fatal("Esc muss Hilfe schließen") }
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestHelpText|TestHelpInputRules'`

Expected: build failure because help helpers do not exist.

- [ ] **Step 3: Implement pure help helpers**

Generate German help text with sections `Navigation`, `Ansichten`, `Banner`, `Ausrichtung`, and `Themes`. Derive display-name lists from the actual definition slices rather than duplicating them. `handleHelpKey` toggles on `KeyCtrlO`, closes on `KeyEscape` only when open, consumes every key while open, and returns false for unrelated keys while closed.

- [ ] **Step 4: Run help and full tests**

Run: `go test ./... -run 'TestHelpText|TestHelpInputRules'`

Expected: PASS.

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add tui.go tui_test.go
git commit -m "feat: define state-preserving TUI help"
```

### Task 5: Integrate Banner, Help, Shortcuts, and Persistence into the TUI

**Files:**
- Modify: `tui.go`
- Modify: `tui_test.go`
- Modify: `main.go`

**Interfaces:**
- Changes: `runTUI(entries []HostEntry, excludes []string, excludePath, themeName, themePath, bannerName, bannerPath, alignmentName, alignmentPath string) (string, error)`.
- Consumes: all Tasks 1–4 interfaces; verified tview APIs `NewTextView`, `TextView.SetDynamicColors`, `TextView.SetTextAlign`, `NewPages`, `Pages.AddPage`, `Pages.SwitchToPage`, `Flex.ResizeItem`, and `Application.SetBeforeDrawFunc`.

- [ ] **Step 1: Write failing status/title tests**

```go
func TestMainTitlesAdvertiseHelp(t *testing.T) {
    if !strings.Contains(titleMain, "^O: Optionen") || !strings.Contains(titleEdit, "^O: Optionen") { t.Fatal("Optionen-Shortcut fehlt im Titel") }
}

func TestFooterSettingsText(t *testing.T) {
    got := settingsStatus(themes[0], banners[2], bannerAlignments[2])
    for _, want := range []string{"Theme: Dunkel", "Banner: Terminal · ANSI", "Ausrichtung: Rechts"} {
        if !strings.Contains(got, want) { t.Errorf("Status enthält %q nicht: %q", want, got) }
    }
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./... -run 'TestMainTitlesAdvertiseHelp|TestFooterSettingsText'`

Expected: FAIL because titles and `settingsStatus` are not updated.

- [ ] **Step 3: Load settings in `main.go` and update the `runTUI` call**

Immediately after loading the theme, load:

```go
bannerPath := filepath.Join(configDir, "flux", "banner")
bannerName, err := LoadBannerName(bannerPath)
if err != nil { return err }
alignmentPath := filepath.Join(configDir, "flux", "banner-alignment")
alignmentName, err := LoadBannerAlignmentName(alignmentPath)
if err != nil { return err }
```

Pass all four new values/paths to `runTUI`. At the start of `runTUI`, resolve banner and alignment indices and return validation errors before constructing the application.

- [ ] **Step 4: Build the banner and help primitives**

Create `bannerView := tview.NewTextView().SetDynamicColors(true)` and `helpView := tview.NewTextView().SetScrollable(true)`. Put `table` and `helpView` into `bodyPages` named `hosts` and `help`. Replace the table item in `content` with `bodyPages` at the same fixed height. Build a `bannerStack` flex containing `bannerView`, a one-row spacer, and `content`; retain the existing horizontal/vertical centering flexes around that stack.

Add `refreshBanner` to set the dynamic-color text, `TextView.SetTextAlign` from the selected alignment, background from the active theme, and footer settings. Do not give focus to the banner.

- [ ] **Step 5: Add shortcuts and strict save handling**

At the top of the existing input capture, call `handleHelpKey`. On transitions, switch `bodyPages` between `help` and `hosts`, set the help title, focus `helpView` or `table`, and call `helpView.ScrollToBeginning()` when opening. Return `nil` for the open/close keys. For other keys while help remains open, return the event directly to the focused `helpView`: this bypasses all hidden-state handlers while retaining the TextView's arrow/PageUp/PageDown scrolling.

Add `KeyCtrlB` and `KeyCtrlA` cases using `nextIndex`. Save through `SaveBannerName` and `SaveBannerAlignmentName`; on failure assign `saveErr`, stop the app, and return `nil`, matching the current theme path. On success refresh banner/footer/layout. Keep `KeyCtrlT`, but also refresh the monochrome banner after applying the new theme.

- [ ] **Step 6: Add responsive hide/show behavior**

Keep a direct reference to the banner stack and install:

```go
app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
    _, screenHeight := screen.Size()
    size := 0
    if bannerVisible(screenHeight, height, banners[bannerIdx]) {
        size = bannerHeight(banners[bannerIdx]) + 1
    }
    bannerStack.ResizeItem(bannerView, size, 0)
    bannerStack.ResizeItem(bannerGap, map[bool]int{true:1, false:0}[size > 0], 0)
    return false
})
```

Use an actual zero-height `tview.NewBox()` as `bannerGap`; avoid a typed-nil primitive. Ensure the centered stack's fixed height is recomputed as `height+size`, so it stays vertically centered across resize and banner changes.

- [ ] **Step 7: Run focused and full verification**

Run: `gofmt -w banner.go banner_test.go tui.go tui_test.go main.go`

Run: `go test ./... -run 'TestMainTitlesAdvertiseHelp|TestFooterSettingsText|TestHelp'`

Expected: PASS.

Run: `go test ./...`

Expected: PASS.

Run: `go vet ./...`

Expected: exit 0 with no findings.

- [ ] **Step 8: Commit**

```bash
git add main.go tui.go tui_test.go banner.go banner_test.go
git commit -m "feat: add banner controls and in-TUI help"
```

### Task 6: Documentation and End-to-End Verification

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: final shortcut names, display names, paths, and behavior from Tasks 1–5.
- Produces: user-facing documentation matching the implementation.

- [ ] **Step 1: Update README behavior and controls**

Add rows for `Ctrl+B`, `Ctrl+A`, and `Ctrl+O` to the controls table. Add a `Banner und Hilfe` section documenting the four styles, theme-aware monochrome behavior, three alignments, persistence paths, `Ctrl+O`/`Esc` close behavior, state preservation, input suppression, and automatic hiding in short terminals. Correct the existing prose that says themes cycle with plain `t`; the implementation uses `Ctrl+T`.

- [ ] **Step 2: Verify documentation terms against code**

Run:

```bash
rg -n 'Ctrl\+B|Ctrl\+A|Ctrl\+H|wordmark-ansi|terminal-mono|banner-alignment|Ctrl\+T' README.md banner.go tui.go
```

Expected: every shortcut and persistence concept appears in both the relevant code and README; no README statement claims plain `t` changes themes.

- [ ] **Step 3: Run final verification**

Run: `gofmt -w *.go`

Run: `go test ./...`

Expected: PASS with no failures.

Run: `go vet ./...`

Expected: exit 0 with no findings.

Run: `go build ./...`

Expected: exit 0 and no output.

Run: `git diff --check`

Expected: exit 0 and no output.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: explain banner and help controls"
```

- [ ] **Step 5: Confirm scope and worktree cleanliness**

Run: `git status --short`

Expected: only the user's pre-existing untracked PNG files and `.superpowers/` mockup directory remain; tracked implementation and documentation changes are committed.
