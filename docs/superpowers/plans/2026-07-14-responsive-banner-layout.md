# Responsive Banner Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the ornamented FLUX wordmark with automatic large/compact selection, theme-primary monochrome color, a capped responsive window width, and a wrapping footer.

**Architecture:** Separate logical banner color mode from physical banner form in `banner.go`, and put all screen-dependent geometry in a new pure `layout.go` unit. `runTUI` keeps the existing state/input behavior but applies a complete `tuiLayout` from its before-draw hook, so resize changes width, footer/body allocation, and banner form atomically.

**Tech Stack:** Go 1.25, `tview` v0.42.0 (`WordWrap`, `TextView`, `Flex.ResizeItem`), `tcell` v2.8.1, `go-runewidth` v0.0.16, standard `testing`.

## Global Constraints

- Use the exact seven-line large and five-line compact `░▒▓` rows from `docs/superpowers/specs/2026-07-14-responsive-banner-layout-design.md`.
- Physical form selection is automatic: large when fully fitting, then compact when fully fitting, otherwise hidden; never clip or scale.
- `Ctrl+B` cycles exactly `ansi` and `monochrome`; monochrome renders with `Theme.Text` and updates immediately on theme change.
- Normalize legacy banner identifiers: `wordmark-ansi`/`terminal-ansi` → `ansi`; `wordmark-mono`/`terminal-mono` → `monochrome`.
- Leave at least two screen columns on each side and cap TUI width at 100 columns.
- Titles and footer text must not increase TUI width; footer wraps without reducing the body below three rows when the screen can accommodate the minimum window.
- Keep `Ctrl+A`, `Ctrl+O`, help, filtering, Backspace, mouse, theme, strict save errors, and SSH behavior unchanged.
- Do not add dependencies.

---

## File Map

- Modify `banner.go`: two logical color modes, exact large/compact forms, legacy normalization, `Theme.Text` rendering.
- Modify `banner_test.go`: definitions, migration, color, and two-mode cycle tests.
- Create `layout.go`: pure width, footer-wrap, vertical-allocation, and banner-form selection.
- Create `layout_test.go`: exact boundary and regression tests for responsive geometry.
- Modify `tui.go`: natural table width only, dynamic footer/body/banner sizing in before-draw, updated status/help names.
- Modify `tui_test.go`: status/help expectations and proof that titles/footer do not enter width calculation.
- Modify `README.md`: automatic form behavior, two modes, text color, responsive width/footer behavior.

### Task 1: Logical Banner Modes, Exact Forms, and Legacy Migration

**Files:**
- Modify: `banner.go`
- Modify: `banner_test.go`

**Interfaces:**
- Produces: `BannerMode{Name, DisplayName string; ColorMode BannerColorMode}`, `banners []BannerMode`, `BannerForm{Name string; Rows []string}`, `largeBanner`, `compactBanner`, `normalizeBannerName(string) (string, error)`.
- Changes: `renderBanner(form BannerForm, mode BannerMode, theme Theme) string`, `alignedBannerText(form BannerForm, mode BannerMode, width int, alignment BannerAlignment, theme Theme) string`.

- [ ] **Step 1: Write failing definition, migration, and color tests**

```go
func TestBannerModesAndForms(t *testing.T) {
    if got := []string{banners[0].Name, banners[1].Name}; !reflect.DeepEqual(got, []string{"ansi", "monochrome"}) {
        t.Fatalf("Banner-Modi: %v", got)
    }
    if len(largeBanner.Rows) != 7 || len(compactBanner.Rows) != 5 { t.Fatalf("Banner-Höhen falsch") }
    for i, row := range compactBanner.Rows {
        if strings.HasPrefix(row, " ") || strings.HasSuffix(row, " ") { t.Errorf("Kompaktzeile %d gepolstert: %q", i, row) }
    }
    if largeBanner.Rows[0] != "░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░" { t.Fatal("große Wortmarke verändert") }
    if compactBanner.Rows[2] != "░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░" { t.Fatal("kompakte Wortmarke verändert") }
}

func TestLegacyBannerNamesNormalize(t *testing.T) {
    cases := map[string]string{"wordmark-ansi":"ansi", "terminal-ansi":"ansi", "wordmark-mono":"monochrome", "terminal-mono":"monochrome", "ansi":"ansi", "monochrome":"monochrome"}
    for in, want := range cases { if got, err := normalizeBannerName(in); err != nil || got != want { t.Errorf("%q: %q, %v", in, got, err) } }
    if _, err := normalizeBannerName("unknown"); err == nil { t.Fatal("unbekannter Modus muss fehlschlagen") }
}

func TestMonochromeBannerUsesThemeText(t *testing.T) {
    th := Theme{Text:tcell.NewHexColor(0x123456), Header:tcell.NewHexColor(0xabcdef)}
    got := renderBanner(compactBanner, banners[1], th)
    if !strings.Contains(got, "[#123456]") || strings.Contains(got, "[#abcdef]") { t.Fatalf("falsche Monochromfarbe: %q", got) }
}
```

- [ ] **Step 2: Run focused tests and verify RED**

Run: `go test ./... -run 'TestBannerModesAndForms|TestLegacyBannerNamesNormalize|TestMonochromeBannerUsesThemeText'`

Expected: build failure because forms/modes/new signatures do not exist.

- [ ] **Step 3: Implement exact forms and two modes**

Copy both code blocks verbatim from the approved design into `largeBanner.Rows` and `compactBanner.Rows`. Replace the four selectable `Banner` definitions with two `BannerMode` values. Make monochrome use `theme.Text.Hex()`; keep the existing ANSI palette and zero-width-rune behavior.

Update `bannerIndex`, display status, and help generation to consume `BannerMode`. `nextIndex` remains unchanged.

- [ ] **Step 4: Implement strict legacy normalization in load/save**

`LoadBannerName` trims through `loadChoice`, then returns the normalized identifier. `SaveBannerName` accepts and writes only `ansi` or `monochrome`; it must not persist legacy identifiers. Missing file defaults to `ansi`; empty/unknown files remain hard errors with the existing context.

- [ ] **Step 5: Update existing renderer/persistence tests and verify GREEN**

Update existing calls to pass a form plus mode. Add file-backed tests proving all four legacy values load as normalized values and a subsequent save writes only the new identifier.

Run: `gofmt -w banner.go banner_test.go && go test ./... -run 'Banner|Legacy'`

Expected: PASS.

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add banner.go banner_test.go tui.go tui_test.go
git commit -m "fix: restore responsive FLUX wordmark modes"
```

### Task 2: Pure Responsive Layout Calculation

**Files:**
- Create: `layout.go`
- Create: `layout_test.go`

**Interfaces:**
- Produces: `tuiLayout{Width, WindowHeight, BodyHeight, FooterHeight int; Banner *BannerForm}`, `calculateTUILayout(screenWidth, screenHeight, naturalWidth, preferredBodyHeight int, footerText string) tuiLayout`, `wrappedLineCount(text string, width int) int`.
- Consumes: `largeBanner`, `compactBanner`, `runewidth.StringWidth`, and verified `tview.WordWrap(text, width)`.

- [ ] **Step 1: Write failing width and wrap tests**

```go
func TestCalculateTUILayoutCapsWidthAndLeavesMargins(t *testing.T) {
    if got := calculateTUILayout(200, 40, 140, 10, "kurz").Width; got != 100 { t.Fatalf("100 erwartet, %d", got) }
    if got := calculateTUILayout(80, 40, 90, 10, "kurz").Width; got != 76 { t.Fatalf("76 erwartet, %d", got) }
}

func TestFooterWrapDoesNotChangeWidth(t *testing.T) {
    short := calculateTUILayout(90, 30, 60, 8, "kurz")
    long := calculateTUILayout(90, 30, 60, 8, strings.Repeat("lang ", 40))
    if short.Width != long.Width { t.Fatalf("Footer änderte Breite: %d != %d", short.Width, long.Width) }
    if long.FooterHeight <= short.FooterHeight { t.Fatalf("Footer brach nicht um") }
}
```

- [ ] **Step 2: Write failing vertical/banner boundary tests**

Calculate the exact display widths with a private `bannerFormWidth`. Test large at exact width/height, compact one row below large, hidden one row below compact, and hidden one column below form width. Assert `BodyHeight >= 3` for screens at or above the minimum supported window height.

- [ ] **Step 3: Run layout tests and verify RED**

Run: `go test ./... -run 'TestCalculateTUILayout|TestFooterWrap|TestBannerFormSelection'`

Expected: build failure because `tuiLayout` and calculator do not exist.

- [ ] **Step 4: Implement `layout.go`**

Use constants:

```go
const (
    maxTUIWidth = 100
    horizontalMargin = 2
    verticalMargin = 1
    minBodyHeight = 3
    borderHeight = 2
    searchHeight = 1
    bannerGapHeight = 1
)
```

Width is `min(maxTUIWidth, screenWidth-2*horizontalMargin)` with a lower bound of 1, then no greater than the natural desired width after raising that desire to the banner form width when the form can fit the screen cap. Inner footer width is `max(1, width-2)` for borders. `wrappedLineCount` returns at least one line and uses `tview.WordWrap`.

Allocate the unbannered window first: borders + search + preferred body + wrapped footer, capped at `screenHeight-2*verticalMargin`. Reduce body toward `minBodyHeight` only after using its preferred height; cap footer to the remaining rows without taking the body below three. Select large, compact, or nil from the remaining height and selected width.

- [ ] **Step 5: Run tests and verify GREEN**

Run: `gofmt -w layout.go layout_test.go && go test ./... -run 'TestCalculateTUILayout|TestFooterWrap|TestBannerFormSelection'`

Expected: PASS.

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add layout.go layout_test.go
git commit -m "feat: calculate responsive TUI geometry"
```

### Task 3: Apply Responsive Geometry in the TUI

**Files:**
- Modify: `tui.go`
- Modify: `tui_test.go`

**Interfaces:**
- Consumes: `calculateTUILayout`, `renderBanner`, `alignedBannerText`, modes/forms from Task 1.
- Produces: resize-driven application of `tuiLayout` using existing `Flex.ResizeItem` and `Application.SetBeforeDrawFunc` APIs.

- [ ] **Step 1: Write failing natural-width and status tests**

Extract `naturalTableWidth(entries []HostEntry) int` and test that it depends on group/host rows only—not `titleMain`, footer text, theme names, banner names, or alignment names. Update status/help assertions to require only `ANSI` and `Monochrom`, with no old Wortmarke/Terminal choices.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `go test ./... -run 'TestNaturalTableWidth|TestFooterSettingsText|TestHelpTextListsCommandsAndOptions'`

Expected: FAIL because current width still uses titles/worst-case footer and old display names.

- [ ] **Step 3: Replace fixed geometry with layout application**

Build `content`, `bannerStack`, `inner`, and `outer` with initial safe sizes. In `SetBeforeDrawFunc`:

1. read `screen.Size()`;
2. calculate the current footer string and `tuiLayout`;
3. choose and render `layout.Banner` with current mode/theme/alignment, or clear it when nil;
4. resize `bannerView`, gap, `bodyPages`, footer, content/window stack, horizontal center item, and vertical center item from the single layout result;
5. return false.

Set `footer.SetWrap(true).SetWordWrap(true)`. Remove `longestSettings`, `footerMax`, and title-driven `maxLine`. Do not mutate query, selection, mode, or persisted settings in before-draw.

- [ ] **Step 4: Update cycling and theme refresh**

`Ctrl+B` cycles two modes and persists their identifiers. `Ctrl+T` redraws monochrome with `Theme.Text`. `Ctrl+A` remains unchanged. `settingsStatus` reports mode plus alignment; help lists the two modes.

- [ ] **Step 5: Verify focused behavior and full suite**

Run: `gofmt -w tui.go tui_test.go && go test ./... -run 'TestNaturalTableWidth|TestFooterSettingsText|TestHelp|TestBanner'`

Expected: PASS.

Run: `go test ./...`

Expected: PASS.

Run: `go vet ./...`

Expected: exit 0.

- [ ] **Step 6: Commit**

```bash
git add tui.go tui_test.go
git commit -m "fix: keep Flux TUI compact and wrap footer"
```

### Task 4: Documentation, Regression Verification, and Local Build

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: final names and behavior from Tasks 1–3.

- [ ] **Step 1: Update README**

Replace the four selectable banner descriptions with two color modes and automatic large/compact/hidden behavior. Document `Theme.Text` monochrome color, 100-column cap, two-column margins, wrapped footer, legacy setting compatibility, and unchanged `Ctrl+A` alignment.

- [ ] **Step 2: Scan for stale behavior**

Run:

```bash
rg -n 'wordmark-ansi|wordmark-mono|terminal-ansi|terminal-mono|Wortmarke ·|Terminal ·|Theme.Header|vier Banner|four banner' README.md banner.go tui.go docs/superpowers/specs/2026-07-14-responsive-banner-layout-design.md
```

Expected: legacy identifiers occur only in normalization/tests or migration documentation; no current UI documentation advertises four selectable styles or `Theme.Header` monochrome color.

- [ ] **Step 3: Run fresh final verification**

Run: `gofmt -d *.go`

Expected: no output.

Run: `go test -count=1 ./...`

Expected: PASS.

Run: `go vet ./...`

Expected: exit 0.

Run: `go build -o flux .`

Expected: exit 0 and an executable ignored `./flux`.

Run: `git diff --check`

Expected: exit 0.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: explain responsive FLUX banner"
```

- [ ] **Step 5: Confirm scope**

Run: `git status --short`

Expected: no tracked changes; the ignored local binary and existing `.superpowers/` scratch directory are not included.
