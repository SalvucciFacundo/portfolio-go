# Design Spec — Portfolio

**Status:** Draft v0.1
**Anchor:** Swiss — dark editorial variant
**Differentiator:** *Monochrome until touched* — the page is grayscale/black-white at rest; color returns on hover (icons, photos, links).

---

## 1. Direction

A single-page developer portfolio with a dark, editorial, precise aesthetic.
The identity lives in one rule: **nothing is in color until the user interacts with it.**
Hover is the reward — it re-introduces color to icons, project photos, and links.

Why this pairing over the safe alternative: the reference image already points to a
dark editorial look with a single red accent. The Industrial/terminal direction
(monospace, 1px borders everywhere) would fight the huge sans-serif scale the
reference shows. Swiss-dark keeps the editorial confidence, and the hover-reveal
differentiator gives the monochrome base a reason to exist.

## 2. Palette (tokens)

| Token | Value | Usage |
|---|---|---|
| `--surface` | `#0A0A0A` | Page background (near-black, not pure #000) |
| `--surface-raised` | `#141414` | Cards, raised panels |
| `--ink` | `#FFFFFF` | Headings, primary text |
| `--ink-muted` | `#9A9A9A` | Secondary text, metadata |
| `--line` | `#262626` | Hairlines, borders, grid rules |
| `--accent` | `#E4002B` | Swiss Red — the single color that returns on hover |

Rules:
- **One accent only.** No gradients, no secondary hues, no warm paper.
- Photos and icons ship in grayscale by default; the accent (or the object's real
  color, limited to one hue family) returns on hover.
- Never center typography. Left-aligned, asymmetric balance.

## 3. Typography

- **Family:** Inter Variable (sans display + body, one family).
  Fallback: `system-ui, -apple-system, sans-serif`. No serifs, no mono as display.
- **Hero scale:** oversized Black/900 heading (`clamp(3rem, 9vw, 8rem)`), tight
  line-height (`0.95`), negative letter-spacing.
- **Body:** Regular/400, 16–18px, generous line-height (`1.7`).
- **Metadata:** small caps / uppercase tracking on dates, stack tags, section
  numbers — used sparingly and always carrying real information.
- **Numerals as composition:** folio numbers, dates set large and light where
  structure benefits (project index, section markers).

## 4. Structure

Single-page scroll, generous margins, hairline grid.

1. **Navbar (fixed)** — brand left, section links right. Text links, no buttons.
   Hover: underline hairline + accent color.
2. **Hero (full viewport)** — giant name/role heading, one supporting line,
   asymmetric. Right side: a large visual with an **organic mask**
   (SVG/clip-path blob), grayscale at rest, color on hover.
3. **Selected Work** — project cards in a grid. Each card: grayscale thumbnail,
   title, year, stack tags. Hover: image returns to color (or accent overlay),
   subtle scale, hairline frame.
4. **About** — short bio, real facts only (no invented metrics).
5. **Contact / Footer** — email link + social icons (grayscale, accent on hover),
   hairline top border, folio numbering.

## 5. Interactions

| Element | Rest | Hover |
|---|---|---|
| Navbar links | `--ink-muted` | `--accent`, 1px underline |
| Project thumbnails | `filter: grayscale(1)` | `grayscale(0)`, `scale(1.02)`, 300ms ease |
| Icons (social/tech) | `--ink-muted`, no fill | `--accent` fill |
| Buttons (if any) | 1px border, transparent bg | accent bg, black text |

- Transitions: 250–350ms `ease`, no spring physics, no parallax noise.
- Respect `prefers-reduced-motion`: disable scale/glow, keep color swap.

## 6. Texture & Material

- Flat surfaces, 1px hairlines instead of shadows.
- No grain, no scanlines, no glassmorphism.
- Organic blob mask in hero is the single non-rectangular shape, used once.

## 7. Content Discipline

- Every string names real information (real name, real projects, real links).
- No fabricated metrics, no filler mono-caps kickers, no unicode-glyph icons.
  Use a real icon set (SVG) or nothing.
- Standard copy for standard actions (`Contact`, `Email me`).
- **Open slots (fill with real data before build):**
  - Real name / role line
  - Project list: title, year, stack, repo/live URL, thumbnail
  - Social links (GitHub, email, LinkedIn if used)

## 8. Non-Goals (v0.1)

- No light mode.
- No CMS, no blog, no admin.
- No animations beyond the hover reveal and gentle section transitions.
- No third-party UI kit; tokens above are the whole system.

## 9. Tech Constraints (Go)

- Server-rendered Go app (templ + HTMX or plain `html/template` — TBD in logic spec).
- CSS custom properties above as single source of truth; no Tailwind config
  unless it maps 1:1 to these tokens.
- Static assets (fonts, images) local, no CDN dependency.
