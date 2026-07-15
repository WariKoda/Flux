# Banner Block Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align uneven banner rows as one block and reduce only the main TUI title to `Flux · ^O: Optionen`.

**Architecture:** Compute one external padding value from the selected form's maximum display width and apply it uniformly after row color rendering. Keep the title change isolated to the existing `titleMain` constant and retain the exact Filter/Options title fixtures.

**Tech Stack:** Go, tview, go-runewidth, Go standard testing

## Global Constraints

- Stored ANSI Regular, Banner3, and all other banner rows remain byte-for-byte unchanged and receive no leading spaces.
- The widest display-width row defines the complete banner block width.
- Left/center/right external padding is computed once per block and applied equally to every row.
- Unicode display-width handling and rainbow phase remain unchanged.
- Only `titleMain` changes to ` Flux · ^O: Optionen `; `titleEdit` and `titleHelp` remain byte-for-byte unchanged.
- Footer, help, shortcuts, themes, responsive sizing, catalog, and persistence remain unchanged.

---

### Task 1: Align banners as complete blocks

**Files:**
- Modify: `banner.go`
- Modify: `banner_test.go`

**Interfaces:**
- Consumes: `alignedBannerText(form BannerForm, mode BannerMode, width int, alignment BannerAlignment, theme Theme) string`, `bannerFormWidth(form BannerForm) int`, and rendered rows from `renderBanner`.
- Produces: uniform external padding for all rows based on the maximum display width.

- [ ] **Step 1: Add failing uneven-row block-alignment tests**

Use `BannerForm{Rows: []string{"123456", "12", "1234"}}` with width `10` and monochrome rendering. Assert stripped results exactly equal:

```go
left := "123456\n12\n1234"
center := "  123456\n  12\n  1234"
right := "    123456\n    12\n    1234"
```

Retain the exact ANSI Regular and Banner3 definition assertions to prove no artwork padding was introduced.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `env GOCACHE=/tmp/flux-go-cache go test -run 'TestAlignedBanner(Text|Block)' -count=1`

Expected: FAIL because the current center/right implementation assigns different padding to shorter rows.

- [ ] **Step 3: Implement one block padding calculation**

In `alignedBannerText`, calculate `blockWidth := bannerFormWidth(form)`. If `width > blockWidth`, derive one padding count from `width-blockWidth` using the selected alignment. Prepend `strings.Repeat(" ", padding)` to every rendered row. Do not calculate padding from individual row widths and do not modify `form.Rows`.

- [ ] **Step 4: Verify focused and full banner tests**

Run:

```bash
gofmt -w banner.go banner_test.go
env GOCACHE=/tmp/flux-go-cache go test -run 'Test(BannerDefinitionsMatchApprovedArt|AlignedBanner)' -count=1
env GOCACHE=/tmp/flux-go-cache go test -count=1 ./...
```

Expected: PASS; uneven rows share one leading padding, Unicode and rainbow tests remain green, and exact artwork is unchanged.

- [ ] **Step 5: Commit block alignment**

```bash
git add banner.go banner_test.go
git commit -m "fix: align banners as complete blocks"
```

### Task 2: Minimize only the main TUI title

**Files:**
- Modify: `tui.go`
- Modify: `tui_test.go`

**Interfaces:**
- Consumes: `titleMain`, `titleEdit`, and `titleHelp` constants.
- Produces: compact main title with unchanged contextual subview titles.

- [ ] **Step 1: Write exact title regression tests**

Replace the broad options-shortcut title assertion with exact values:

```go
if titleMain != " Flux · ^O: Optionen " {
	t.Fatalf("Haupttitel = %q", titleMain)
}
if titleEdit != " Flux · Filter — Enter/Klick/Leertaste: umschalten · ^E/Esc: fertig · ^O: Optionen " {
	t.Fatalf("Filtertitel verändert: %q", titleEdit)
}
if titleHelp != " Flux · Optionen — ^O/Esc: zurück " {
	t.Fatalf("Optionstitel verändert: %q", titleHelp)
}
```

- [ ] **Step 2: Run the title test and verify RED**

Run: `env GOCACHE=/tmp/flux-go-cache go test -run '^Test.*Title' -count=1`

Expected: FAIL because `titleMain` still contains search, connect, filter, theme, and exit hints.

- [ ] **Step 3: Change only the main title constant**

Set:

```go
titleMain = " Flux · ^O: Optionen "
```

Leave `titleEdit` and `titleHelp` unchanged.

- [ ] **Step 4: Run complete verification and rebuild local binary**

Run:

```bash
gofmt -d *.go
env GOCACHE=/tmp/flux-go-cache go test -count=1 ./...
env GOCACHE=/tmp/flux-go-cache go vet ./...
env GOCACHE=/tmp/flux-go-cache go build -o /home/bdgraue/BERUF/DEV/Projects/flux/flux .
git diff --check
```

Expected: formatting and whitespace output are empty; tests and vet pass; the ignored main-project binary builds successfully.

- [ ] **Step 5: Commit the compact title**

```bash
git add tui.go tui_test.go
git commit -m "ui: minimize main TUI title"
```
