# TAAG Banner Catalog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the approved thirteen-state banner catalog, exact TAAG `rainbow3` rendering, corrected responsive BlurVision forms, fixed-form layout behavior, persistence migration, and matching TUI documentation.

**Architecture:** Replace the current split between two color modes and global BlurVision forms with catalog entries that own a form family and color strategy. Keep rendering pure in `banner.go`; pass the selected catalog entry into the pure `calculateTUILayout` function so only BlurVision can choose a compact fallback. `tui.go` remains responsible for cycling, persistence, and applying the calculated geometry.

**Tech Stack:** Go, tcell, tview, go-runewidth, standard `testing`

## Global Constraints

- Implement `docs/superpowers/specs/2026-07-14-taag-banner-catalog-design.md` exactly.
- Do not add a runtime FIGlet dependency or generate banners at runtime.
- Preserve the 100-column cap, two-column margins, wrapped footer, minimum body height, theme shortcuts, alignment shortcut, and SSH behavior.
- `rainbow3` uses the twelve exact RGB values and `floor((column + row + 1) / 2) mod 12`; the single-row banner uses each color once.
- Alignment padding must not change the banner color phase.
- Empty and unknown settings continue to fail explicitly.
- Use TDD for every behavioral change and commit each task independently.

---

### Task 1: Catalog, Exact Forms, Color Rendering, and Migration

**Files:**
- Modify: `banner.go`
- Modify: `banner_test.go`

**Interfaces:**
- Produces: `BannerMode{Name, DisplayName string; Family BannerFamily; ColorMode BannerColorMode}`.
- Produces: `BannerFamily{Name string; Forms []BannerForm}` where forms are ordered from preferred to fallback.
- Produces: `bannerNone BannerColorMode`, the thirteen-entry `banners`, six exact families, and `bannerRainbow3Colors []string`.
- Produces: `renderBanner(form BannerForm, mode BannerMode, theme Theme) string`, `alignedBannerText(form BannerForm, mode BannerMode, width int, alignment BannerAlignment, theme Theme) string`, and unchanged persistence function signatures.
- Consumes: exact identifiers, rows, colors, and migration map from the approved spec.

- [ ] **Step 1: Replace definition tests with failing catalog and exact-form tests**

Add table-driven assertions for this exact order:

```go
want := []string{
    "blurvision-rainbow3", "blurvision-monochrome",
    "single-rainbow3", "single-monochrome",
    "ansi-regular-rainbow3", "ansi-regular-monochrome",
    "banner3-rainbow3", "banner3-monochrome",
    "ansi-compact-rainbow3", "ansi-compact-monochrome",
    "terrace-rainbow3", "terrace-monochrome", "none",
}
```

Assert family form counts and heights: BlurVision `7,5`; single `1`; ANSI Regular `5`; Banner3 `7`; ANSI Compact `3`; Terrace `7`; none `0`. Assert every row against the complete code blocks in the approved spec, including:

```go
if got := blurVisionFamily.Forms[1].Rows[4]; got != "░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░" {
    t.Fatalf("kompakte Abschlusszeile: %q", got)
}
```

Run: `go test ./... -run 'TestBannerCatalog|TestExactBannerForms'`
Expected: FAIL because the catalog and families do not exist.

- [ ] **Step 2: Add failing rendering tests for exact TAAG phase and single-row exception**

Use a small form to make every cell observable:

```go
func TestRainbow3UsesTAAGDiagonalPhase(t *testing.T) {
    form := BannerForm{Rows: []string{"abcd", "efgh", "ijkl"}}
    got := renderBanner(form, BannerMode{ColorMode: bannerRainbow3}, Theme{})
    want := "[#ff2828]a[#ff7800]bc[#ffb400]d\n" +
        "[#ff7800]ef[#ffb400]gh\n" +
        "[#ff7800]i[#ffb400]jk[#ffdc00]l"
    if got != want { t.Fatalf("rainbow3:\n%s\nwant:\n%s", got, want) }
}

func TestSingleRowUsesEveryRainbow3ColorOnce(t *testing.T) {
    got := renderBanner(singleFamily.Forms[0], banners[2], Theme{})
    for _, color := range bannerRainbow3Colors {
        if strings.Count(got, "["+color+"]") != 1 { t.Errorf("Farbe %s nicht genau einmal: %q", color, got) }
    }
}
```

Retain and update combining-mark, wide-rune, theme-independence, monochrome `Theme.Text`, and alignment tests. Add an assertion that center/right padding precedes the first color tag but does not change the first glyph color.

Run: `go test ./... -run 'TestRainbow3|TestSingleRow|TestMonochrome|TestAlignedBanner'`
Expected: FAIL with old palette/output.

- [ ] **Step 3: Add failing persistence and migration tests**

Test the exact legacy map:

```go
cases := map[string]string{
    "ansi": "blurvision-rainbow3",
    "monochrome": "blurvision-monochrome",
    "wordmark-ansi": "single-rainbow3",
    "wordmark-mono": "single-monochrome",
    "terminal-ansi": "ansi-regular-rainbow3",
    "terminal-mono": "ansi-regular-monochrome",
}
```

Update round-trip tests to persist `terrace-monochrome`, verify missing files default to `blurvision-rainbow3`, verify `none` round-trips, and verify old identifiers are rejected by `SaveBannerName`.

Run: `go test ./... -run 'TestLegacyBanner|TestBannerSettings|TestMissingBanner|TestSaveBanner'`
Expected: FAIL with current normalization and two-entry validation.

- [ ] **Step 4: Implement the catalog and exact static forms**

Introduce:

```go
type BannerColorMode int
const (
    bannerRainbow3 BannerColorMode = iota
    bannerMonochrome
    bannerNone
)

type BannerFamily struct {
    Name  string
    Forms []BannerForm
}

type BannerMode struct {
    Name, DisplayName string
    Family            BannerFamily
    ColorMode         BannerColorMode
}
```

Define the six families with the exact rows in the approved spec. Define the catalog in the thirteen-entry order from Step 1; `none` owns an empty family and uses `bannerNone`. Use concise German names such as `BlurVision · Regenbogen`, `ANSI Compact · Monochrom`, and `Kein Banner`.

Run: `gofmt -w banner.go banner_test.go && go test ./... -run 'TestBannerCatalog|TestExactBannerForms'`
Expected: PASS for definitions while rendering/migration tests still fail.

- [ ] **Step 5: Implement exact rendering without coloring alignment padding**

Set:

```go
var bannerRainbow3Colors = []string{
    "#ff2828", "#ff7800", "#ffb400", "#ffdc00",
    "#dcff00", "#78ff00", "#00ff50", "#00ffa0",
    "#00c8ff", "#0078ff", "#7850ff", "#ff00c8",
}
```

Render each form row before alignment. For normal `rainbow3`, choose the palette index from `(displayColumn + rowIndex + 1) / 2 % len(palette)`. For `single-rainbow3`, choose `displayColumn % len(palette)`. Zero-width combining runes keep the prior tag; a wide rune advances by its full display width but receives one tag. `bannerNone` returns an empty string. `alignedBannerText` calculates and prepends raw spaces after `renderBanner`, line by line, so padding never advances the color phase.

Run: `gofmt -w banner.go banner_test.go && go test ./... -run 'TestRainbow3|TestSingleRow|TestMonochrome|TestAlignedBanner'`
Expected: PASS.

- [ ] **Step 6: Implement stable normalization and persistence**

Make `normalizeBannerName` return catalog identifiers unchanged and apply the Step 3 legacy map. `LoadBannerName` validates through normalization and returns the normalized name. `SaveBannerName` accepts only stable catalog identifiers through `bannerIndex`.

Run: `gofmt -w banner.go banner_test.go && go test ./... -run 'Banner|Legacy|Alignment'`
Expected: PASS.

- [ ] **Step 7: Commit Task 1**

```bash
git add banner.go banner_test.go
git commit -m "feat: add exact TAAG banner catalog"
```

---

### Task 2: Selected-Family Responsive Layout

**Files:**
- Modify: `layout.go`
- Modify: `layout_test.go`

**Interfaces:**
- Consumes: `BannerMode.Family.Forms`, ordered preferred-to-fallback.
- Produces: `calculateTUILayout(screenWidth, screenHeight, naturalWidth, preferredBodyHeight int, footerText string, mode BannerMode) tuiLayout`.
- Produces: `tuiLayout.Banner *BannerForm`, nil for explicit none or temporary hiding.

- [ ] **Step 1: Write failing family-specific width and height tests**

Update every current call with an explicit mode. Test:

```go
func TestLayoutUsesOnlySelectedBannerFamily(t *testing.T) {
    mode := banners[8] // ANSI Compact rainbow3
    layout := calculateTUILayout(80, 30, 20, 3, "kurz", mode)
    if layout.Banner == nil || layout.Banner.Name != "ansi-compact" { t.Fatalf("Banner: %+v", layout.Banner) }
    if layout.Width != max(20, bannerFormWidth(mode.Family.Forms[0])) { t.Fatalf("Width: %d", layout.Width) }
}
```

Add boundaries for: BlurVision large -> corrected compact -> nil; each fixed family exact fit -> nil one row/column below; none never selects a form or widens the TUI; footer wrapping and tiny-terminal nonnegative allocations remain unchanged.

Run: `go test ./... -run 'TestLayout|TestBannerFormSelection|TestCalculateTUILayout|TestTinyTerminal'`
Expected: FAIL because layout has no selected-mode input and always considers BlurVision.

- [ ] **Step 2: Generalize layout selection**

Change the signature to accept `mode BannerMode`. Raise `desiredWidth` only to the first selected-family form width that fits `widthCap`. After allocating the unbannered window, iterate `mode.Family.Forms` in order and select the first complete form whose width and height plus `bannerGapHeight` fit. An empty family leaves `Banner` nil and does not change width.

Do not select forms from another family and do not mutate the persisted mode.

Run: `gofmt -w layout.go layout_test.go && go test ./... -run 'TestLayout|TestBannerFormSelection|TestCalculateTUILayout|TestTinyTerminal|TestFooterWrap'`
Expected: PASS.

- [ ] **Step 3: Commit Task 2**

```bash
git add layout.go layout_test.go
git commit -m "feat: size the selected banner family"
```

---

### Task 3: TUI Cycle, Help, Status, and Documentation

**Files:**
- Modify: `tui.go`
- Modify: `tui_test.go`
- Modify: `README.md`

**Interfaces:**
- Consumes: thirteen-entry `banners`, generalized `calculateTUILayout(..., mode)`, unchanged save/load and alignment functions.
- Produces: visible status/help copy for all catalog choices; `Ctrl+B` persists the next stable identifier.

- [ ] **Step 1: Write failing status/help and no-banner integration tests**

Update `settingsStatus` expectations to the first catalog display name. Assert `helpText()` contains `Ctrl+B`, all six design names, `Monochrom`, `Regenbogen`, and `Kein Banner`, while not containing legacy identifiers. Assert the footer/status for `none` says `Banner: Kein Banner`.

Update any before-draw/helper tests or extracted-call tests so layout receives the active `banners[bannerIdx]`. Keep `Ctrl+O`, `Ctrl+H`, `Ctrl+A`, and existing input behavior unchanged.

Run: `go test ./... -run 'TestSettingsStatus|TestHelp|TestBanner|TestNaturalTableWidth'`
Expected: FAIL with two-mode copy and old layout call.

- [ ] **Step 2: Wire the selected mode into TUI layout and rendering**

In `SetBeforeDrawFunc`, call:

```go
mode := banners[bannerIdx]
layout := calculateTUILayout(screenWidth, screenHeight, naturalTableWidth(entries), totalRows, currentFooter, mode)
```

If `layout.Banner == nil`, clear the view and allocate zero banner/gap height. Otherwise render that returned form with `mode`. Keep `Ctrl+B` as one cycle through all thirteen entries and persist `banners[bannerIdx].Name`. Theme changes repaint monochrome entries; alignment changes repaint all visible entries.

Run: `gofmt -w tui.go tui_test.go && go test ./... -run 'TestSettingsStatus|TestHelp|TestBanner|TestNaturalTableWidth'`
Expected: PASS.

- [ ] **Step 3: Update README with exact user behavior**

Replace the two-mode banner section with the six design families, rainbow/monochrome variants, no-banner choice, exact `rainbow3` description, single-row full-spectrum exception, corrected BlurVision 7/5 fallback, fixed-form hide behavior, legacy compatibility, and unchanged `Ctrl+A` alignment. Do not document internal type names.

Run: `rg -n 'ANSI|Monochrom|BlurVision|Banner3|ANSI Compact|Terrace|Kein Banner|Ctrl\+B|Ctrl\+A' README.md`
Expected: every design and shortcut is documented.

- [ ] **Step 4: Run focused and full verification**

```bash
gofmt -d *.go
go test -count=1 ./...
go vet ./...
go build -o flux .
git diff --check
```

Expected: no formatting diff, all tests pass, vet/build succeed, no whitespace errors. Confirm `git status --short` lists only intended source/doc changes and ignored `flux` is not staged.

- [ ] **Step 5: Commit Task 3**

```bash
git add tui.go tui_test.go README.md
git commit -m "feat: expose the complete banner cycle"
```

---

### Task 4: Final Catalog Verification and Review Fixes

**Files:**
- Modify only files required by concrete review findings.

**Interfaces:**
- Consumes: Tasks 1-3 and the approved spec.
- Produces: a review-clean, fully verified feature branch.

- [ ] **Step 1: Compare implementation against every spec section**

Verify catalog/order, every exact row, palette/formula, full-spectrum single row, monochrome theme behavior, alignment phase, selected-family layout, no-banner gap, migrations, help/status, README, and out-of-scope constraints. Record no speculative enhancements.

- [ ] **Step 2: Run an independent code review and fix only confirmed findings with TDD**

For each finding, first add or adjust a failing test, run it to observe failure, make the smallest fix, and rerun the focused suite. Do not combine unrelated cleanup.

- [ ] **Step 3: Run fresh final verification**

```bash
gofmt -d *.go
go test -count=1 ./...
go vet ./...
go build -o flux .
git diff --check
git status --short
```

Expected: every command succeeds; the worktree is clean except ignored local artifacts.

- [ ] **Step 4: Commit review fixes if any**

```bash
git add banner.go banner_test.go layout.go layout_test.go tui.go tui_test.go README.md
git commit -m "fix: complete TAAG banner catalog review"
```

If there are no findings, do not create an empty commit.
