# Flux Banner-First Terminal Layout Design

## Goal

Use the terminal width more efficiently and keep the selected banner visible
at constrained heights by making the hosts table scroll sooner. Preserve the
existing banner catalog, color rendering, alignment, footer wrapping, theme
behavior, and 100-column maximum.

This design supersedes the width-selection and vertical banner-allocation rules
in `2026-07-14-responsive-banner-layout-design.md` and
`2026-07-14-taag-banner-catalog-design.md`. All other requirements from those
documents remain unchanged.

## Width Allocation

The TUI always uses all horizontal space available inside the existing
two-column margins, capped at 100 columns:

```text
width = max(1, min(100, screenWidth - 2 * horizontalMargin))
```

Natural table width and selected banner width no longer shrink the window.
They remain relevant only to content rendering and determining whether the
selected banner form fits without horizontal clipping.

Consequences:

- a 120-column terminal produces a 100-column TUI;
- an 80-column terminal produces a 76-column TUI;
- narrow host data no longer produces a narrow floating window;
- footer text wraps against the full selected TUI width;
- banners continue to obey left, center, and right alignment within that width.

## Vertical Priority

The layout prioritizes a selected banner before expanding the hosts body. The
vertical allocation order is:

1. top and bottom terminal margins;
2. one complete selected banner form and its one-row gap;
3. window borders and search row;
4. at least three visible hosts-body rows;
5. wrapped footer rows;
6. additional hosts-body rows up to the preferred content height.

The body begins scrolling as soon as its content exceeds the allocated body
height. Existing table navigation and mouse-wheel behavior provide the scroll
mechanism; no new scroll widget or shortcut is introduced.

## Banner Form Selection

A form is eligible only when its complete display width fits the calculated TUI
width and its height can coexist with:

- the banner gap;
- window borders;
- the search row; and
- the three-row minimum body.

Footer height is not required for banner eligibility. At extreme heights, the
footer may shrink to zero before a selected banner is hidden.

For BlurVision, forms remain ordered large then compact:

1. show the seven-row form when it fits with the minimum usable window;
2. otherwise show the corrected five-row form when it fits;
3. otherwise hide the banner temporarily.

Every fixed-form family is shown completely when eligible and hidden otherwise.
No form is clipped, scaled, or replaced by a form from another family. A
temporarily hidden banner remains selected and reappears after the terminal is
enlarged.

The explicit `none` choice reserves no banner height or gap, so all otherwise
usable height remains available to the normal window calculation. The window
still stops at its preferred content height rather than adding empty body rows.

## Window and Footer Allocation

After selecting a banner, subtract its height and gap from the available
terminal height before calculating the window.

Within the remaining window space:

- borders and the search row stay fixed where physically possible;
- the body receives its preferred height when possible, but is reduced toward
  three rows before the banner is affected;
- wrapped footer rows receive remaining space after the three-row body floor;
- when banner, fixed window rows, and the three-row body fit but the complete
  footer does not, the footer is truncated to the remaining row count,
  including zero;
- for terminals too small even for fixed window rows, existing nonnegative
  tiny-terminal reconciliation remains in effect.

The layout must reconcile exactly: allocated banner, gap, window, margins,
body, footer, borders, and search rows never exceed the screen height and no
component receives a negative size.

## TUI Integration

`runTUI` continues to apply one complete `tuiLayout` in its before-draw hook.
Resizing the terminal therefore updates full-width geometry, banner form,
body height, footer height, and scrollable viewport atomically.

The banner remains physically above the Flux window. Help replaces the hosts
body inside the same viewport and remains independently scrollable.

## Testing

Automated tests cover:

- full available width at large, medium, and tiny screen widths;
- width independence from natural table width and selected banner width;
- footer wrapping against the full width;
- BlurVision large, compact, and hidden boundaries while preserving a
  three-row body;
- exact-height fixed-form visibility with a three-row body;
- footer reduction to zero before banner hiding;
- preferred body reduction causing earlier scrolling;
- `none` reserving no banner height or gap while retaining preferred window
  height;
- complete height reconciliation and nonnegative sizes on tiny terminals;
- unchanged selected-family, alignment, help, and table navigation behavior.

Fresh Go formatting, tests, vetting, build, and whitespace checks are required
before completion.

## Out of Scope

- Removing or raising the 100-column width cap
- Changing the two-column horizontal margins
- Clipping or horizontally scaling banners
- Overlaying a banner on the Flux window
- New scrolling shortcuts or widgets
- Changes to banner glyphs, colors, catalog order, persistence, or themes
