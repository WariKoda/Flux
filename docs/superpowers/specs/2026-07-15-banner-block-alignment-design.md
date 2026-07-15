# Banner Block Alignment Design

## Goal

Align each multi-line banner as one rectangular artwork block so rows retain
their authored relative positions, and reduce only the main TUI title to the
minimum discoverable options hint.

## Root Cause

`alignedBannerText` currently calculates external left padding independently
from every row's display width. With centered or right alignment, shorter rows
therefore receive more padding and move right relative to longer rows. ANSI
Regular and Banner3 visibly bend even though their stored rows all start at
column zero.

## Block Alignment

The maximum display width of all rows in the selected `BannerForm` defines the
banner block width. External padding is calculated once from the available TUI
width and this block width:

- left: zero spaces;
- center: `(available width - block width) / 2` spaces;
- right: `available width - block width` spaces.

The same external padding is prepended to every rendered row. Short rows are
not independently centered or right-aligned, and the stored banner artwork is
not modified or padded with leading spaces. Existing Unicode display-width
handling and rainbow phase behavior remain unchanged because external padding
is applied only after color rendering.

This rule applies to every banner family, preventing the same distortion in
future uneven-width artwork. It preserves the selected block alignment: the
widest row determines the left/center/right position of the complete artwork.

## Main Title

Only the main-view title changes to:

```text
 Flux · ^O: Optionen 
```

The Filter and Options titles remain byte-for-byte unchanged because their
contextual controls are useful while those modes are active. The search bar,
footer details/settings status, help content, shortcuts, and README shortcut
documentation remain unchanged.

## Testing

Automated tests cover uneven-width rows for left, center, and right block
alignment; Unicode display widths; unchanged rainbow phase; exact ANSI Regular
and Banner3 artwork; the compact main title; and unchanged Filter/Options
titles. Fresh formatting, full tests, vet, build, and whitespace checks are
required before completion.

## Out of Scope

- Editing ANSI Regular, Banner3, or any other stored glyph rows
- Adding leading spaces to artwork definitions
- Changing banner selection, colors, themes, responsive sizing, or persistence
- Shortening the Filter or Options titles
- Changing footer content or help text
