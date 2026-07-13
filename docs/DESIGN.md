# Design

## 1. Aesthetic direction

**Provenance Check is an exhibit binder.** The page reads like a legal exhibit dossier: a
parchment page, serif headings set like a brief, and every result rendered as an exhibit
card stamped with a rotated verdict badge — CLEAR / CAUTION / RESTRICTED — the way a
courtroom exhibit gets a sticker and a case number. Quoted clause text sits in monospace,
like a transcript excerpt, so the reader can tell "the tool's words" from "the license's
words" at a glance. This direction hasn't been used across recent ships (which have leaned
blueprint/technical, paper-and-ink case-file, risograph, and vapor-gradient) — exhibit-binder
is close in spirit to paper-and-ink but distinct in palette (warm parchment, not cream-folder
neutral) and signature motif (the rotated stamp, not a dossier tab).

## 2. Tokens

**Color**

| Token | Value | Use |
|---|---|---|
| `--bg` | `#F4EFE2` | page background — warm parchment |
| `--surface-1` | `#FBF8F0` | exhibit card surface |
| `--surface-2` | `#EAE1CC` | recessed areas: input panel, table header |
| `--text` | `#2A241C` | body text — warm near-black ink |
| `--text-muted` | `#6E6353` | secondary text, captions, metadata |
| `--accent` | `#1F3A5F` | ink navy — links, focus ring, primary CTA |
| `--accent-support` | `#A9812E` | brass — secondary emphasis, dividers, active tab |
| `--success` | `#2F6B4F` | CLEAR verdict stamp |
| `--warning` | `#B8842C` | CAUTION verdict stamp |
| `--danger` | `#B23A2E` | RESTRICTED verdict stamp |

**Type**

- Display: `"Source Serif 4"` (Google Fonts) — wordmark, page headings, section titles.
  Fallback: `Georgia, "Times New Roman", serif`.
- UI/data: `"IBM Plex Mono"` (Google Fonts) — body copy, URLs, verdict labels, quoted clause
  text, buttons. Fallback: `"SFMono-Regular", Consolas, monospace`.
- Scale: 0.8rem / 1rem / 1.25rem / 1.563rem / 1.953rem / 2.441rem (1.25 ratio from a 1rem base).

**Spacing & shape**

- Spacing unit: 8px scale (8/16/24/32/48/64).
- Corner radius: 3px on cards and inputs (crisp, paper-like — never pill-shaped); verdict
  stamps are circular/oval, an intentional exception to the square-corner rule.
- Shadow: a stacked-paper shadow — `0 1px 2px rgba(42,36,28,0.08), 0 4px 12px rgba(42,36,28,0.10)`
  offset down-right, like a card resting slightly above the page. No glow; this direction is
  matte, not glassy.
- Motion: UI transitions 150–200ms ease-out. A verdict stamp "hits" the card with a
  120ms scale(1.3→1) + rotate flourish when its result resolves — the one game-adjacent
  moment in an otherwise static tool.

## 3. Layout intent

**Hero:** the exhibit grid — the list of URLs the user pasted, each rendered as an exhibit
card with its stamp, SPDX identifier, and (for non-green results) the quoted clause. The
paste input is secondary chrome, not the hero.

- **Desktop (1440×900):** two-column frame — a narrow (≈320px) sticky left panel holding the
  paste textarea, `Check` button, and a running count ("5 of 5 resolved"); the remaining
  ≈70% of the viewport is the exhibit grid (2-column card grid), scrollable independently.
  The grid is the majority of the screen at all times, satisfying the ≥60% hero rule.
- **Phone (390×844):** stacked — paste panel on top (collapsible once results exist, so the
  grid gets the screen), exhibit cards full-width below in a single column.
- No dead space: an empty state before the first check fills the grid area with a large
  illustrated "case file" empty state (see signature detail), not blank parchment.

## 4. Signature detail

The **verdict stamp**: a rotated (−8°) circular badge with a double-ring border and
letterpress-style condensed caps ("CLEAR" / "CAUTION" / "RESTRICTED"), rendered as inline SVG
so it's crisp at any size and reused as the favicon (a single navy-on-parchment ring monogram
"PC"). The stamp "hits" each card on resolve with the scale+rotate flourish described above —
one memorable, on-brand flourish rather than decorative particles.

## 5. Interaction & feedback plan (non-game, but interactive)

- **Empty state:** illustrated case-file icon + "Paste a list of dataset or repo URLs to
  begin" — designed, not blank.
- **Loading state:** each pending card shows a shimmer placeholder shaped like the final card
  (badge outline, two text lines) so the grid never collapses or reflows as results land.
- **Success/resolve:** the stamp-hit flourish (120ms) plus the clause quote fading in
  (150ms ease-out).
- **Error state:** a card styled like a "returned to sender" stamp (danger-adjacent but
  visually distinct from RESTRICTED) for unreachable URLs or fetch failures, with a retry
  action.
- Respect `prefers-reduced-motion`: stamp flourish becomes an instant opacity swap, shimmer
  becomes a static tone.

Every later build/QA run follows this file. Changes to direction or tokens get their own
commit explaining why.
