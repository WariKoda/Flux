# ANSI Regular Optical Balance Design

## Goal

Correct the apparent rightward offset of the middle ANSI Regular row without
moving any complete row or changing the alignment of the `FLUX` wordmark.

## Artwork Change

ANSI Regular keeps its existing five-row form. Only row 3 changes from a
five-cell full-block F crossbar to five full blocks followed by a left half
block:

```text
███████ ██      ██    ██ ██   ██
██      ██      ██    ██  ██ ██
█████▌  ██      ██    ██   ███
██      ██      ██    ██  ██ ██
██      ███████  ██████  ██   ██
```

The following two spaces after `▌` preserve the existing starting column of
the `L`, `U`, and `X`. No whole row receives leading whitespace.

## Banner3

Banner3 remains exactly as currently implemented. All seven rows start at the
same column, and row 4 keeps `######` without a half block. No row shifting or
glyph correction is applied to Banner3.

## Behavior and Compatibility

Rainbow coloring, monochrome theme coloring, alignment, responsive layout,
catalog order, persistence, shortcuts, and all other banner artwork remain
unchanged.

## Testing

Update the exact ANSI Regular artwork assertion to cover the half-block row and
retain the exact Banner3 assertion as regression coverage. Run formatting, the
full Go test suite, vet, build, and whitespace checks before completion.
