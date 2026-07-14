# Flux TUI Banner and Help Design

## Goal

Add a configurable banner directly above Flux's centered host-picker window and
add in-TUI help for all commands and options. Banner style, alignment, and theme
remain independent settings. The host picker must remain usable in terminals
that are too short to display the banner.

## Banner Styles

Flux ships four banner styles in this cycle order:

1. Compact wordmark with fixed ANSI rainbow colors: `▓▒░ FLUX ░▒▓`
2. Compact wordmark in one theme-derived color
3. Four-line terminal wordmark with fixed ANSI rainbow colors
4. Four-line terminal wordmark in one theme-derived color

The four-line shape is:

```text
█▀▀▀  █     █  █  ▀█▄█▀
█▀▀   █     █  █    █
█     █     █  █  ▄█▀█▄
▀     ▀▀▀▀   ▀▀   ▀   ▀
```

ANSI styles use a fixed left-to-right terminal rainbow and do not change when
the UI theme changes. Monochrome styles use the active theme's existing header
accent color: aqua for Dark, navy for Light, green for Matrix, sky blue for
Okabe-Ito Dark, and blue for Okabe-Ito Light. A theme change recolors a visible
monochrome banner immediately.

`Ctrl+B` advances to the next style and wraps from style 4 to style 1. The
selected style is saved immediately and restored on the next launch.

## Alignment

The banner is aligned relative to the left and right edges of the TUI window,
not relative to the full terminal. Flux supports this cycle order:

1. Left
2. Center
3. Right

`Ctrl+A` advances to the next alignment and wraps from Right to Left. Alignment
is independent of banner style and theme, is saved immediately, and is restored
on the next launch.

## Layout and Small-Terminal Behavior

The banner and existing TUI window form one vertically stacked, horizontally
centered layout block. There is one blank row between the banner and the window.
The selected alignment determines the banner's position within a row whose
width equals the calculated TUI window width.

Flux shows the banner only when the current screen can accommodate the complete
banner, separating row, and existing TUI window. It never clips or partially
draws a banner. If the screen is too short, Flux hides only the banner and gives
all available space to the existing host picker. Resizing re-evaluates this
decision, so the banner returns when sufficient height becomes available. The
saved style and alignment do not change when the banner is temporarily hidden.

The existing centered-window behavior, selection, filtering, mouse support,
and scrolling remain unchanged.

## Help View

`Ctrl+O` opens an in-TUI options/help view. The help replaces the host table within the
same centered window; it is not a second process or external pager. Opening help
preserves the active main/filter mode, query, table selection, banner style,
alignment, and theme.

The help lists:

- typing, Backspace, Escape, arrow keys, Home, End, Enter, mouse click, and
  mouse-wheel behavior;
- `Ctrl+E` for the filter view;
- `Ctrl+T` for theme cycling, including all theme display names;
- `Ctrl+B` for banner cycling, including all four banner style names;
- `Ctrl+A` for alignment cycling, including Left, Center, and Right;
- `Ctrl+O` for opening and closing options/help.

While help is open, `Ctrl+O` and `Escape` close it and restore the preserved
view. Other input is consumed so it cannot mutate the hidden search, filter,
selection, theme, banner, or alignment state. The help content may scroll when
the window cannot show every line.

## Titles and Status

The main and filter titles include `^O: Optionen`; the help title says that
`Ctrl+O` or `Esc` closes it. To keep the border title readable, banner and
alignment shortcuts are documented in help rather than adding every shortcut
to the already long main title.

The footer continues to show the active theme and additionally shows the active
banner style and alignment. Existing host-detail and filter-status text remains
intact.

## Configuration and Errors

Banner style and alignment follow the existing theme persistence pattern and
live in separate files under Flux's existing configuration directory:

- `~/.config/flux/banner`
- `~/.config/flux/banner-alignment`

A missing file selects the first value in its cycle as the default. A file must
contain exactly one known, non-empty identifier after surrounding whitespace is
trimmed. Empty or unknown values, read failures other than a missing file, and
write failures are hard errors with actionable messages. Parent directories and
file permissions follow the existing theme writer.

No migration is required because both settings are new and missing files are
valid.

## Components and Data Flow

Banner definitions are data rather than conditionals scattered through the TUI.
Each definition has a stable persisted identifier, help/display name, rows, and
color mode. Alignment likewise has a stable identifier and display name.

Startup loads and validates theme, banner style, and alignment before running
the TUI. The selected banner is rendered into a dedicated primitive above the
existing content. Theme, banner, and alignment changes update the primitive,
footer, and layout immediately, then persist the changed setting. Any save
failure stops the application and is returned to the caller, matching current
theme behavior.

Help is a distinct view state layered over the existing main/filter mode rather
than a third value in that mode. This lets closing help restore the exact prior
mode without reconstructing it.

## Testing

Tests cover behavior at unit boundaries without requiring an interactive real
terminal:

- banner and alignment identifier validation, defaults, load/save behavior,
  permissions where testable, and strict error cases;
- deterministic cycle order and wraparound for all four styles and three
  alignments;
- fixed ANSI colors versus theme-derived monochrome colors;
- left, center, and right padding relative to the computed TUI width;
- visibility calculations for both banner heights, exact-fit height, one row
  too short, and resize re-evaluation;
- help opening/closing and preservation of mode, query, and selection;
- input suppression while help is visible;
- title, footer, and help content exposing the required commands and option
  names.

The final verification runs formatting, the complete Go test suite, `go vet`,
and a build of the application.

## Out of Scope

- User-authored banner definitions or arbitrary banner configuration
- Editing settings from the help view
- Animations or timed color changes
- Scaling or partially clipping banner glyphs
- Changing existing theme definitions or SSH behavior
