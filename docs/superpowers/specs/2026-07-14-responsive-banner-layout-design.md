# Flux Responsive Banner and TUI Width Design

## Goal

Correct the banner and layout regressions introduced by the configurable banner
feature. Flux must preserve the original ornamented `░▒▓` wordmark, choose an
appropriate height automatically, use the theme's primary text color for
monochrome output, and keep the host picker comfortably narrower than the
terminal while wrapping long footer content.

This design supersedes the banner-style and fixed-width portions of
`2026-07-14-tui-banner-and-help-design.md`. Help, alignment, filtering, theme
cycling, SSH behavior, and strict persistence remain unchanged unless stated
here.

## Banner Forms

Flux has one logical wordmark with two physical forms. Users do not select the
form directly.

The large form is the original seven-line wordmark:

```text
░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓████████▓▒░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░
```

The compact form is this left-aligned five-line wordmark, with no leading or
trailing padding stored in its rows:

```text
░▒▓████████▓▒░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓██████▓▒░ ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░░▒▓██████▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
░▒▓█▓▒░      ░▒▓█▓▒░     ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░
```

The renderer chooses on every draw:

1. Show the large form when its full width and its seven rows plus the one-row
   separator fit around the usable TUI content.
2. Otherwise show the compact form when its full width and its five rows plus
   separator fit.
3. Otherwise hide the banner completely.

Forms are never clipped, scaled, or partially drawn. A resize immediately
re-evaluates the choice. The saved color mode and alignment do not change when
the physical form changes or disappears.

## Color Modes and Persistence

`Ctrl+B` cycles exactly two logical modes:

1. `ansi` — the existing fixed left-to-right ANSI rainbow
2. `monochrome` — a single color taken from the active `Theme.Text` value

Changing the theme recolors a visible monochrome banner immediately. This means
Dark and Okabe-Ito Dark use white, Light and Okabe-Ito Light use black, and
Matrix uses green with the current theme definitions.

The existing `~/.config/flux/banner` file remains in use. To avoid breaking
settings written by Flux v0.3.0, loading accepts and normalizes legacy values:

- `wordmark-ansi` and `terminal-ansi` become `ansi`
- `wordmark-mono` and `terminal-mono` become `monochrome`

The next `Ctrl+B` change writes only the new stable identifiers. Unknown and
empty values remain hard errors. Alignment persistence and `Ctrl+A` are
unchanged.

## Responsive TUI Width

The TUI window width is recalculated from the current screen size rather than
fixed from worst-case footer combinations.

- Leave at least two terminal columns empty on both the left and right.
- Never make the TUI wider than 100 columns.
- Within those limits, use the natural host-table width, with enough width to
  display a banner form when that form fits the terminal.
- Long border titles and footer strings do not increase the chosen width.
- tview may clip a border title when the selected width is shorter than the
  title; all commands remain discoverable through `Ctrl+O` help.
- Recalculate width and horizontal centering after terminal resize.

The banner remains aligned with `Ctrl+A` relative to the selected TUI width.
If neither banner form fits that width, the banner is hidden rather than
forcing the TUI wider than its limits.

## Footer Wrapping and Vertical Allocation

The footer remains a `TextView`, but its height is derived from the current
footer text and current inner width. It wraps on word boundaries where
possible. A short footer uses one row; longer content uses the number of rows it
needs while space is available.

The complete centered window leaves at least one terminal row above and below.
When vertical space is constrained, allocation priority is:

1. window border, search row, and at least three body rows;
2. as many wrapped footer rows as fit;
3. compact or large banner only when the remaining height permits it.

The host table remains scrollable when fewer body rows than host entries are
visible. Footer wrapping must never widen the TUI, and hidden overflow must not
cause a crash or corrupt selection/search state.

## Components

Pure layout helpers calculate:

- the capped TUI width from screen width and natural table width;
- the wrapped footer row count for a given inner width;
- the available body height;
- the selected large, compact, or hidden banner form.

`runTUI` applies those results in its existing before-draw resize hook. Banner
data remains separate from rendering, and color-mode persistence remains
separate from physical form selection.

## Testing

Tests cover:

- exact large and compact rows, including the absence of leading indentation;
- large-to-compact-to-hidden selection at exact width/height boundaries and one
  column/row below each boundary;
- resize-driven form selection without changing mode or alignment;
- `Ctrl+B` cycling only `ansi` and `monochrome`;
- legacy identifier normalization and strict rejection of unknown/empty files;
- monochrome rendering from `Theme.Text` for every current theme;
- width capped at 100, two-column side margins, and narrow-terminal behavior;
- footer wrapping at representative widths and vertical allocation preserving
  at least three body rows;
- long titles and footer strings not influencing window width;
- existing help, Backspace, filtering, persistence, theme, and alignment tests.

Final verification runs formatting, the complete Go test suite, `go vet`, and a
fresh build.

## Out of Scope

- User-authored banner text
- Terminal font size, weight, or letter-spacing control
- Horizontal clipping or scaling of banner glyphs
- New themes or changes to existing theme palettes
- Changes to SSH execution or host parsing
