# ANSI Regular Optical Balance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend only the ANSI Regular F middle crossbar by a left half block while preserving the positions of the remaining letters and all Banner3 artwork.

**Architecture:** Keep the existing static `BannerFamily` definitions and exact-artwork regression table. Change one expected test row first, then make the corresponding one-row production artwork change without altering rendering logic.

**Tech Stack:** Go, Go standard testing, Unicode terminal artwork

## Global Constraints

- ANSI Regular row 3 is exactly `█████▌  ██      ██    ██   ███`.
- The two spaces after `▌` preserve the current `L`, `U`, and `X` columns.
- Banner3 remains byte-for-byte unchanged, including row 4 `######` without a half block.
- Rainbow coloring, monochrome theme coloring, alignment, responsive layout, catalog order, persistence, shortcuts, and every other banner remain unchanged.

---

### Task 1: Balance the ANSI Regular middle row

**Files:**
- Modify: `banner_test.go`
- Modify: `banner.go`

**Interfaces:**
- Consumes: `ansiRegularFamily` and the exact family artwork table in `TestBannerDefinitionsMatchApprovedArt`.
- Produces: the approved ANSI Regular row while preserving the existing `BannerFamily` interface and Banner3 regression fixture.

- [ ] **Step 1: Change the exact artwork expectation**

In the ANSI Regular case in `TestBannerDefinitionsMatchApprovedArt`, replace only the third row with:

```go
"█████▌  ██      ██    ██   ███"
```

Leave the complete Banner3 expectation unchanged so it continues to prove that no leading spaces or half blocks were introduced.

- [ ] **Step 2: Run the focused test and verify RED**

Run: `env GOCACHE=/tmp/flux-go-cache go test -run '^TestBannerDefinitionsMatchApprovedArt$/ANSI_Regular$' -count=1`

Expected: FAIL showing the old `█████   ...` production row differs from the new `█████▌  ...` expectation.

- [ ] **Step 3: Make the minimal artwork change**

In `ansiRegularFamily` in `banner.go`, replace only its third row with:

```go
"█████▌  ██      ██    ██   ███"
```

Do not change any other row or family.

- [ ] **Step 4: Verify focused and complete behavior**

Run:

```bash
gofmt -d *.go
env GOCACHE=/tmp/flux-go-cache go test -count=1 ./...
env GOCACHE=/tmp/flux-go-cache go vet ./...
env GOCACHE=/tmp/flux-go-cache go build -o /home/bdgraue/BERUF/DEV/Projects/flux/flux .
git diff --check
```

Expected: formatting and whitespace checks are empty; focused/full tests and vet pass; the ignored local binary builds successfully.

- [ ] **Step 5: Commit the artwork correction**

```bash
git add banner.go banner_test.go
git commit -m "fix: balance ANSI Regular banner"
```
