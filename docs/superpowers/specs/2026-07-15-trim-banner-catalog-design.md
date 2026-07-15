# Flux Trimmed Banner Catalog Design

## Goal

Remove the visually unsatisfactory Terrace banner and retain only the compact
five-row BlurVision form. Preserve color variants, alignment, theme behavior,
the banner-first responsive layout, settings compatibility, and every other
banner design.

This design supersedes the Terrace and multi-form BlurVision portions of
`2026-07-14-taag-banner-catalog-design.md` and the BlurVision fallback wording
in `2026-07-15-banner-first-terminal-layout-design.md`. All unrelated
requirements remain unchanged.

## Catalog and Cycle

`Ctrl+B` advances through these eleven persisted choices in this exact order:

1. `blurvision-rainbow3`
2. `blurvision-monochrome`
3. `single-rainbow3`
4. `single-monochrome`
5. `ansi-regular-rainbow3`
6. `ansi-regular-monochrome`
7. `banner3-rainbow3`
8. `banner3-monochrome`
9. `ansi-compact-rainbow3`
10. `ansi-compact-monochrome`
11. `none`

Terrace is removed from the catalog, status/help cycle text, documentation,
and static banner definitions. ANSI Compact becomes the last visible family
before `none`.

## BlurVision Form

BlurVision retains only the approved corrected five-row form:

```text
░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░
```

The former seven-row form is removed. BlurVision now follows the same
complete-or-hidden layout rule as every other fixed family: show the complete
five-row form when its width and height fit with the banner gap and minimum
usable window; otherwise hide it temporarily. It reappears after sufficient
terminal space returns.

## Persistence Migration

New writes accept only the eleven stable identifiers above. Existing valid
identifiers for retained designs are unchanged.

Previously stored Terrace identifiers normalize by preserving their color
semantics:

- `terrace-rainbow3` -> `blurvision-rainbow3`
- `terrace-monochrome` -> `blurvision-monochrome`

This migration occurs on load so an existing Terrace selection does not prevent
Flux from starting. A later save writes the normalized BlurVision identifier.
All older pre-catalog migration rules remain unchanged.

## TUI and Documentation

Help and settings status derive from the eleven-entry catalog and therefore no
longer mention Terrace. README design lists, cycle counts, responsive behavior,
and migration notes describe five visible families and the single fixed
BlurVision form.

No shortcut changes are introduced. `Ctrl+B`, `Ctrl+A`, `Ctrl+T`, `Ctrl+O`,
and all navigation/scroll behavior remain unchanged.

## Testing

Automated tests cover:

- the exact eleven-entry catalog and wraparound order;
- absence of Terrace from definitions and help text;
- exact corrected five-row BlurVision content and one-form family size;
- fixed-form BlurVision exact-fit and one-row/one-column-below hiding;
- removal of seven-row fallback assumptions from layout tests;
- Terrace-to-BlurVision load migration for both color modes;
- rejection of Terrace identifiers by new saves;
- round trips for retained identifiers and `none`;
- README stale-text checks;
- all existing color, alignment, width, banner-first height, footer, scrolling,
  theme, persistence, and tiny-terminal tests.

Fresh Go formatting, tests, vetting, build, and whitespace checks are required
before completion.

## Out of Scope

- Manually redesigning or retaining Terrace
- Replacing Terrace with another banner family
- Changing remaining glyphs, colors, or catalog order
- Reintroducing a large BlurVision form
- Changes to width caps, margins, height priorities, footer behavior, themes,
  shortcuts, or scrolling controls
