# Flux TAAG Banner Catalog Design

## Goal

Restore the two banner designs removed by the responsive-banner change, add
three user-selected TAAG designs, reproduce the TAAG `rainbow3` colors exactly,
and add an explicit no-banner choice. The existing responsive TUI width,
wrapping footer, help, theme, and alignment behavior remain intact.

This design extends
`2026-07-14-responsive-banner-layout-design.md`. Where the earlier design
limits `Ctrl+B` to two color modes, this document supersedes that behavior with
the catalog and cycle defined below.

## Banner Catalog and Cycle

`Ctrl+B` advances through these thirteen persisted choices in this exact order:

1. `blurvision-rainbow3` — BlurVision ASCII with `rainbow3`
2. `blurvision-monochrome` — BlurVision ASCII in the theme text color
3. `single-rainbow3` — `▓▒░ FLUX ░▒▓` using its full-spectrum mapping
4. `single-monochrome` — `▓▒░ FLUX ░▒▓` in the theme text color
5. `ansi-regular-rainbow3` — ANSI Regular with `rainbow3`
6. `ansi-regular-monochrome` — ANSI Regular in the theme text color
7. `banner3-rainbow3` — Banner3 with `rainbow3`
8. `banner3-monochrome` — Banner3 in the theme text color
9. `ansi-compact-rainbow3` — ANSI Compact with `rainbow3`
10. `ansi-compact-monochrome` — ANSI Compact in the theme text color
11. `terrace-rainbow3` — Terrace with `rainbow3`
12. `terrace-monochrome` — Terrace in the theme text color
13. `none` — no banner

The status/footer and help dialog use concise German display names for these
choices. `none` renders no banner and consumes no banner gap.

## Exact Banner Forms

All forms are static strings derived from the user-selected TAAG output. Flux
does not generate FIGlet text at runtime and does not add a FIGlet dependency.
Trailing whitespace and fully empty outer FIGfont rows are omitted.

### BlurVision ASCII

The large form remains the exact seven visible rows:

```text
░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░
```

The corrected compact form keeps the original final row so L and U retain
their lower strokes:

```text
░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░
```

Only BlurVision has multiple physical forms. The layout selects large,
compact, or temporarily hidden according to the existing width and height
rules.

### Single Row

```text
▓▒░ FLUX ░▒▓
```

### ANSI Regular

```text
███████ ██      ██    ██ ██   ██
██      ██      ██    ██  ██ ██
█████   ██      ██    ██   ███
██      ██      ██    ██  ██ ██
██      ███████  ██████  ██   ██
```

### Banner3

```text
######## ##       ##     ## ##     ##
##       ##       ##     ##  ##   ##
##       ##       ##     ##   ## ##
######   ##       ##     ##    ###
##       ##       ##     ##   ## ##
##       ##       ##     ##  ##   ##
##       ########  #######  ##     ##
```

### ANSI Compact

The two leading and one trailing empty FIGfont rows are not stored:

```text
██████ ▄▄    ▄▄ ▄▄ ▄▄ ▄▄
██▄▄   ██    ██ ██ ▀█▄█▀
██     ██▄▄▄ ▀███▀ ██ ██
```

### Terrace

```text
░██████████░██
░██        ░██
░██        ░██ ░██    ░██ ░██    ░██
░█████████ ░██ ░██    ░██  ░██  ░██
░██        ░██ ░██    ░██   ░█████
░██        ░██ ░██   ░███  ░██  ░██
░██        ░██  ░█████░██ ░██    ░██
```

## Color Rendering

### TAAG `rainbow3`

The twelve-color palette is fixed and theme-independent:

1. `(255, 40, 40)`
2. `(255, 120, 0)`
3. `(255, 180, 0)`
4. `(255, 220, 0)`
5. `(220, 255, 0)`
6. `(120, 255, 0)`
7. `(0, 255, 80)`
8. `(0, 255, 160)`
9. `(0, 200, 255)`
10. `(0, 120, 255)`
11. `(120, 80, 255)`
12. `(255, 0, 200)`

For every display cell in the multi-row TAAG designs, the zero-based palette
index is:

```text
floor((column + row + 1) / 2) mod 12
```

This reproduces the two-cell bands and diagonal row phase from TAAG. Spaces
within a form consume display cells and therefore advance the colors. Alignment
padding added outside the form does not affect the color calculation. Combining
marks remain attached to the preceding display cell and wide runes retain one
color across their full cell width.

The single-row design is a deliberate exception. Its twelve characters use the
twelve palette entries exactly once from left to right, so the short banner
shows the complete spectrum rather than ending in green.

### Monochrome

All monochrome variants use `Theme.Text`. A theme change recolors a visible
monochrome banner immediately. The glyphs and layout are otherwise identical to
the corresponding colored form.

## Responsive Layout

The existing 100-column TUI cap, two-column horizontal margins, natural table
width, wrapped footer, and minimum body height remain unchanged.

The selected form may raise the desired TUI width only up to the existing cap
and only when it fits inside the current screen margins. Fixed-height forms are
shown when their complete width and height plus the banner gap fit. Otherwise
they are temporarily hidden without changing or persisting a different choice.
BlurVision alone falls back from seven rows to its corrected five-row form
before being hidden.

Left, center, and right alignment continue to be relative to the selected TUI
width and are controlled by the existing alignment setting and shortcut.

## Persistence and Migration

The existing banner settings file remains the single persisted choice. New
writes use only the thirteen stable identifiers in the catalog.

Legacy values normalize as follows:

- `ansi` -> `blurvision-rainbow3`
- `monochrome` -> `blurvision-monochrome`
- `wordmark-ansi` -> `single-rainbow3`
- `wordmark-mono` -> `single-monochrome`
- `terminal-ansi` -> `ansi-regular-rainbow3`
- `terminal-mono` -> `ansi-regular-monochrome`

Missing settings default to `blurvision-rainbow3`. Empty files and unknown
identifiers remain explicit errors rather than silently selecting a banner.

## Testing

Automated tests cover:

- exact catalog identifiers and `Ctrl+B` cycle order;
- exact rows and dimensions for every form;
- the corrected final row of compact BlurVision;
- the twelve RGB values, two-cell grouping, diagonal row phase, space handling,
  combining marks, and wide runes;
- the single-row full-spectrum exception;
- theme-dependent monochrome rendering;
- no-banner height and gap behavior;
- large-to-compact-to-hidden BlurVision selection;
- complete-or-hidden selection for fixed forms at width and height boundaries;
- alignment padding remaining outside the color phase;
- persistence, defaults, legacy migration, and invalid-setting failures;
- help, footer status, and option text for all thirteen choices.

Full Go formatting, tests, vetting, build, and whitespace checks are required
before the implementation is considered complete.

## Out of Scope

- Runtime FIGlet generation or a FIGlet dependency
- User-authored banner text
- Editable color palettes
- Horizontal clipping or scaling of glyphs
- Changes to themes, table contents, SSH behavior, or non-banner shortcuts
