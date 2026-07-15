# Trimmed Banner Catalog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove Terrace and the tall BlurVision form while preserving a compatible eleven-state banner cycle with the approved five-row BlurVision design.

**Architecture:** Keep the existing data-driven banner catalog and rendering pipeline. Trim the static definitions, normalize legacy Terrace selections during settings load, and update layout expectations so BlurVision behaves like every other fixed complete-or-hidden family.

**Tech Stack:** Go, Bubble Tea/Lip Gloss TUI, Go standard testing, Markdown

## Global Constraints

- `Ctrl+B` cycles the exact eleven choices specified in `docs/superpowers/specs/2026-07-15-trim-banner-catalog-design.md`.
- BlurVision has exactly one form: the approved corrected five-row artwork.
- `terrace-rainbow3` loads as `blurvision-rainbow3`; `terrace-monochrome` loads as `blurvision-monochrome`.
- New saves reject Terrace identifiers.
- Width caps, margins, banner-first height priorities, footer behavior, themes, shortcuts, and scrolling controls remain unchanged.
- Complete fresh checks are required: formatting, tests, vet, build, and whitespace validation.

---

### Task 1: Trim the catalog and preserve settings compatibility

**Files:**
- Modify: `banner.go`
- Modify: `banner_test.go`

**Interfaces:**
- Consumes: existing `Banner`, `BannerFamily`, `normalizeBannerName(string) string`, settings load/save validation, and catalog cycling helpers.
- Produces: an eleven-entry `banners` catalog, a one-form `blurVisionFamily`, and Terrace legacy normalization without accepting Terrace as a current identifier.

- [ ] **Step 1: Write failing catalog, artwork, and persistence tests**

Update `banner_test.go` so the expected stable identifiers are exactly:

```go
wantNames := []string{
	"blurvision-rainbow3", "blurvision-monochrome",
	"single-rainbow3", "single-monochrome",
	"ansi-regular-rainbow3", "ansi-regular-monochrome",
	"banner3-rainbow3", "banner3-monochrome",
	"ansi-compact-rainbow3", "ansi-compact-monochrome",
	"none",
}
```

Assert `len(blurVisionFamily.Forms) == 1` and that its rows equal the five lines in the design spec. Add both Terrace cases to the legacy-load table, assert saving either Terrace identifier returns an error, and use retained identifiers plus `none` for round-trip coverage. Remove assertions that expect Terrace or a seven-row BlurVision form.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `go test ./... -run 'Test(Banner|Blur|Settings|Config|Legacy|RoundTrip|Save|Load)' -count=1`

Expected: FAIL because the catalog still contains Terrace, BlurVision still has two forms, and Terrace is still accepted as current state.

- [ ] **Step 3: Implement the minimal catalog and migration change**

In `banner.go`, remove `largeBanner` and `terraceFamily`, define `blurVisionFamily` from only the approved five-row `compactBanner`, remove both Terrace entries from `banners`, and leave `none` last. Extend legacy normalization with:

```go
case "terrace-rainbow3":
	return "blurvision-rainbow3"
case "terrace-monochrome":
	return "blurvision-monochrome"
```

Keep these cases outside the current-identifier acceptance path so loads migrate old values while save validation rejects them.

- [ ] **Step 4: Run focused and full tests**

Run: `gofmt -w banner.go banner_test.go && go test -count=1 ./...`

Expected: PASS with eleven catalog states, exact five-row artwork, both migrations, and no Terrace definitions.

- [ ] **Step 5: Commit the catalog change**

```bash
git add banner.go banner_test.go
git commit -m "feat: trim banner catalog"
```

### Task 2: Reconcile responsive layout, help, and documentation

**Files:**
- Modify: `layout_test.go`
- Modify: `tui_test.go`
- Modify: `README.md`

**Interfaces:**
- Consumes: Task 1's eleven-entry `banners` catalog and single-form `blurVisionFamily`.
- Produces: fixed-form exact-fit layout coverage, Terrace-free help coverage, and user documentation matching the final catalog.

- [ ] **Step 1: Update tests to express fixed BlurVision behavior and Terrace-free UI**

In `layout_test.go`, replace hard-coded former `none` indices with `banners[len(banners)-1]`. Test the sole BlurVision form at exact width and at one column below; compute the exact visible height from margins, its five rows, banner gap, fixed/search rows, and the three-row minimum body, then assert one row below hides it. Remove Terrace from family loops and remove large-to-compact fallback expectations.

In `tui_test.go`, remove `Terrace` from expected help content and add:

```go
if strings.Contains(help, "Terrace") {
	t.Fatal("help unexpectedly advertises removed Terrace banner")
}
```

- [ ] **Step 2: Run focused tests and verify they identify stale assumptions**

Run: `go test ./... -run 'Test(BannerFormSelection|BannerFirst|Footer|Help|NoBanner|Layout)' -count=1`

Expected: PASS only after all stale indices, fallback assumptions, and Terrace help expectations are removed; any remaining stale assumption fails with its focused assertion.

- [ ] **Step 3: Update README catalog and compatibility text**

Document five visible families and eleven cycle states. List only BlurVision, Single, ANSI Regular, Banner3, and ANSI Compact; describe BlurVision as a fixed five-row complete-or-hidden form; remove Terrace from current choices; and document both Terrace-to-BlurVision legacy mappings. Preserve the existing banner-first width, footer, scrolling, theme, alignment, and shortcut descriptions.

- [ ] **Step 4: Add a stale-documentation regression check**

Add or extend a Go test that reads `README.md` and fails if it contains a current Terrace design entry, claims six designs or thirteen states, or describes a seven-to-five-row BlurVision fallback. Assert the README contains both legacy mappings.

- [ ] **Step 5: Run all verification commands**

Run:

```bash
gofmt -w layout_test.go tui_test.go banner_test.go
gofmt -d *.go
go test -count=1 ./...
go vet ./...
go build -o flux .
git diff --check
```

Expected: formatting produces no diff; tests and vet pass; the local ignored binary builds; whitespace check is clean.

- [ ] **Step 6: Commit layout tests and documentation**

```bash
git add layout_test.go tui_test.go banner_test.go README.md
git commit -m "docs: explain trimmed banner catalog"
```
