# Banner-First Terminal Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Flux use the full available terminal width and reserve a selected banner before reducing the hosts viewport to its three-row scrolling minimum.

**Architecture:** Keep all geometry in the pure `calculateTUILayout` function. Select width directly from the screen cap, choose an eligible banner against a minimum usable window, then allocate the remaining window height in priority order: fixed rows, three body rows, footer, and additional body rows. `runTUI` already consumes `tuiLayout` atomically and needs no structural change.

**Tech Stack:** Go, tview, go-runewidth, standard `testing`

## Global Constraints

- Implement `docs/superpowers/specs/2026-07-15-banner-first-terminal-layout-design.md` exactly.
- Preserve the 100-column maximum and two-column horizontal margins.
- Use all width available inside those limits; natural table and banner widths must not shrink the TUI.
- Keep a complete selected banner visible whenever it fits with its gap, borders, search row, and three body rows.
- BlurVision falls back from seven rows to the corrected five-row form before hiding.
- Footer rows may reduce to zero before an eligible banner is hidden.
- No component may receive a negative size or exceed the screen height.
- Do not change banner glyphs, colors, catalog order, persistence, themes, alignment, help behavior, table navigation, or scrolling controls.
- Use TDD and commit each task independently.

---

### Task 1: Full-Width, Banner-First Geometry

**Files:**
- Modify: `layout.go`
- Modify: `layout_test.go`

**Interfaces:**
- Consumes: existing constants, `BannerMode.Family.Forms`, `bannerFormWidth`, `bannerHeight`, and `wrappedLineCount`.
- Preserves: `calculateTUILayout(screenWidth, screenHeight, naturalWidth, preferredBodyHeight int, footerText string, mode BannerMode) tuiLayout`.
- Produces: full-width `tuiLayout.Width`; banner-first `Banner`, `WindowHeight`, `BodyHeight`, and `FooterHeight` values.

- [ ] **Step 1: Write failing full-width tests**

Replace shrink-wrap assumptions with exact screen-derived widths:

```go
func TestCalculateTUILayoutUsesAvailableWidthUpToCap(t *testing.T) {
    tests := []struct{ name string; screenWidth, naturalWidth, want int }{
        {"capped", 200, 1, 100},
        {"inside margins", 80, 1, 76},
        {"natural width ignored", 80, 140, 76},
        {"tiny", 3, 140, 1},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := calculateTUILayout(tt.screenWidth, 40, tt.naturalWidth, 3, "kurz", banners[12]).Width
            if got != tt.want { t.Fatalf("Width = %d, %d erwartet", got, tt.want) }
        })
    }
}
```

Update `TestNoneBannerNeverSelectsOrWidens` to expect width `100` on a 200-column screen regardless of natural width. Add a footer test showing identical screen widths produce identical wrap counts for different natural widths.

Run: `env GOCACHE=/tmp/flux-go-cache go test -count=1 ./... -run 'TestCalculateTUILayoutUsesAvailableWidth|TestNoneBanner|TestFooterWrap'`
Expected: FAIL because current width remains shrink-wrapped to natural/banner width.

- [ ] **Step 2: Write failing banner-priority boundary tests**

Use the minimum usable window height:

```go
minimumWindowHeight := borderHeight + searchHeight + minBodyHeight
largeExact := 2*verticalMargin + bannerHeight(largeBanner) + bannerGapHeight + minimumWindowHeight
compactExact := 2*verticalMargin + bannerHeight(compactBanner) + bannerGapHeight + minimumWindowHeight
```

Assert:

- BlurVision selects large at `largeExact` with body height `3` and footer `0`;
- one row below `largeExact` selects compact rather than hiding;
- compact remains visible at `compactExact` with body height `3` and footer `0`;
- one row below `compactExact` hides the banner;
- a preferred body height of `50` still reduces to `3` to preserve large BlurVision;
- a fixed ANSI Compact banner remains visible at its exact banner + gap + minimum-window height;
- a long footer becomes `0` at the exact boundary before the banner hides;
- one additional row goes to the footer before growing the body above `3`.

Run: `env GOCACHE=/tmp/flux-go-cache go test -count=1 ./... -run 'TestBannerFirst|TestBannerFormSelection|TestSelectedBannerFamilyBoundaries|TestFooterShrinksBeforeBanner'`
Expected: FAIL because current code allocates the preferred window before considering the banner.

- [ ] **Step 3: Add failing reconciliation and no-banner regression tests**

For screen heights `0..40`, every catalog mode, a preferred body height of `50`, and a long footer, assert:

```go
if layout.BodyHeight < 0 || layout.FooterHeight < 0 || layout.WindowHeight < 0 {
    t.Fatalf("negative Größe: %+v", layout)
}
bannerAndGap := 0
if layout.Banner != nil { bannerAndGap = bannerHeight(*layout.Banner) + bannerGapHeight }
if got := 2*verticalMargin + bannerAndGap + layout.WindowHeight; got > screenHeight && screenHeight >= 2*verticalMargin {
    t.Fatalf("Layout %d Zeilen größer als Screen %d: %+v", got, screenHeight, layout)
}
```

Retain the existing tiny-terminal fixed/body reconciliation. For `none`, assert no banner/gap is reserved and the window stops at fixed rows + preferred body + wrapped footer rather than adding empty rows.

Run: `env GOCACHE=/tmp/flux-go-cache go test -count=1 ./... -run 'TestLayoutHeightReconciliation|TestTinyTerminal|TestNoneBanner'`
Expected: existing cases may pass, while the new cross-mode reconciliation guards the new allocation.

- [ ] **Step 4: Implement full-width calculation**

Replace natural/banner-derived width selection with:

```go
width := max(1, min(maxTUIWidth, screenWidth-2*horizontalMargin))
```

Keep `naturalWidth` in the function signature for compatibility with the existing pure-call interface, but explicitly ignore it with `_ = naturalWidth`. Footer wrapping uses `max(1, width-borderHeight)`.

Run: `gofmt -w layout.go layout_test.go && env GOCACHE=/tmp/flux-go-cache go test -count=1 ./... -run 'TestCalculateTUILayoutUsesAvailableWidth|TestNoneBanner|TestFooterWrap'`
Expected: PASS.

- [ ] **Step 5: Implement banner-first selection and window allocation**

Calculate:

```go
availableHeight := max(0, screenHeight-2*verticalMargin)
minimumWindowHeight := borderHeight + searchHeight + minBodyHeight
```

Select the first form whose full width fits and whose `height + gap + minimumWindowHeight` fits `availableHeight`. Reserve its height and gap before window allocation.

For `windowCapacity := availableHeight - bannerAndGap`:

1. If `windowCapacity < borderHeight+searchHeight+minBodyHeight`, retain nonnegative tiny-terminal reconciliation: allocate fixed rows where possible, give remaining rows to body, and set footer to zero.
2. Otherwise allocate `minBodyHeight`, then as many wrapped footer rows as fit, then grow body toward `max(minBodyHeight, preferredBodyHeight)` with the remaining capacity.
3. Set `WindowHeight = borderHeight + searchHeight + BodyHeight + FooterHeight`; do not grow beyond preferred body height.

Run: `gofmt -w layout.go layout_test.go && env GOCACHE=/tmp/flux-go-cache go test -count=1 ./... -run 'TestBannerFirst|TestBannerFormSelection|TestSelectedBannerFamilyBoundaries|TestFooterShrinksBeforeBanner|TestLayoutHeightReconciliation|TestTinyTerminal|TestNoneBanner'`
Expected: PASS.

- [ ] **Step 6: Run full Task 1 verification**

```bash
gofmt -d *.go
env GOCACHE=/tmp/flux-go-cache go test -count=1 ./...
env GOCACHE=/tmp/flux-go-cache go vet ./...
env GOCACHE=/tmp/flux-go-cache go build -o flux .
git diff --check
```

Expected: all commands exit successfully and `flux` remains ignored.

- [ ] **Step 7: Commit Task 1**

```bash
git add layout.go layout_test.go
git commit -m "fix: prioritize banners in terminal layout"
```

---

### Task 2: User Documentation and Final Verification

**Files:**
- Modify: `README.md`
- Modify only if a concrete review defect requires it: `layout.go`, `layout_test.go`, `tui.go`, `tui_test.go`

**Interfaces:**
- Consumes: Task 1 geometry and the existing TUI before-draw integration.
- Produces: accurate user-facing width, banner-priority, footer, and scrolling documentation; review-clean branch.

- [ ] **Step 1: Update README behavior text**

Replace the final responsive-layout paragraphs in “Banner und Hilfe” so they state:

- the TUI uses all available width within two-column margins, capped at 100;
- a selected banner is reserved before body expansion;
- BlurVision selects seven rows, then five rows, then hidden;
- every banner must fit completely with borders, search, and at least three body rows;
- the hosts list scrolls when more rows exist than the allocated viewport;
- footer rows shrink, potentially to zero, before an eligible banner is hidden;
- fixed forms remain complete-or-hidden and reappear after resize.

Do not change shortcut, glyph, color, catalog, persistence, theme, or alignment documentation.

Run: `rg -n '100|zwei Terminalspalten|drei|scroll|Fußzeile|BlurVision|verborgen' README.md`
Expected: all approved concepts are present and old shrink-wrap/banner-last wording is absent.

- [ ] **Step 2: Review TUI integration without speculative changes**

Verify `tui.go` still passes the active mode to `calculateTUILayout`, resizes `bodyPages` from `layout.BodyHeight`, assigns zero banner/gap sizes when `layout.Banner == nil`, and uses the tview table as the scrollable hosts viewport. Make no code change if those contracts already hold.

- [ ] **Step 3: Run independent review and fix confirmed findings with TDD**

Compare every section of the approved spec with the Task 1 diff and README. For any concrete behavior defect, first add a failing focused test, observe RED, apply the smallest fix, and rerun the covering suite. Do not add unrelated layout options.

- [ ] **Step 4: Run fresh final verification**

```bash
gofmt -d *.go
env GOCACHE=/tmp/flux-go-cache go test -count=1 ./...
env GOCACHE=/tmp/flux-go-cache go vet ./...
env GOCACHE=/tmp/flux-go-cache go build -o flux .
git diff --check
git status --short
```

Expected: formatting is clean; tests, vet, and build pass; no whitespace errors; only intended tracked files are committed and local `.superpowers/`/`flux` artifacts remain ignored.

- [ ] **Step 5: Commit Task 2**

```bash
git add README.md layout.go layout_test.go tui.go tui_test.go
git commit -m "docs: explain banner-first terminal layout"
```

If review causes no source changes, Git stages and commits only `README.md`.
