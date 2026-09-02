# gotsx UI — design system for the demos

One system for `site`, `shop`, `admin`, `example` and everything `gotsx new` scaffolds. It is modeled on
**shadcn/ui's neutral (zinc) theme**: semantic tokens instead of raw palettes, a single neutral accent
(near-black in light, white in dark), hairline borders, small radii, generous whitespace, no gradients.

| File | What |
|---|---|
| `gotsx.css` | Tailwind v4 layer: tokens (`:root` / `.dark`), `@theme inline` mapping, base styles, component classes (`.btn`, `.input`, `.card`, `.badge`, `.table`, …) |
| `plain.css` | the same tokens and classes as hand-written CSS (no Tailwind) — used by `example` and the default scaffold |
| `design.go` | embeds both so the CLI can write them into new apps |

## Tokens

`background` `foreground` · `card` `card-foreground` · `muted` `muted-foreground` · `border` `input` `ring` ·
`primary` `primary-foreground` · `secondary` `secondary-foreground` · `accent` `accent-foreground` ·
`destructive` `destructive-foreground` · `success` `warning` · `--radius` (0.5rem).

In Tailwind they are colors: `bg-background text-foreground border-border bg-card text-muted-foreground bg-primary
text-primary-foreground hover:bg-accent text-destructive …`, and radii `rounded-sm/md/lg/xl` derive from `--radius`.

## Rules (what makes it "refined")

1. **Only tokens.** No `zinc-500`, `slate-*`, `brand-*`, hex values or gradients in app code. If something needs a
   color, it is one of the tokens above.
2. **One accent.** The primary button is `bg-primary` (black / white). Status uses `success` / `warning` /
   `destructive` at 10% tint for badges, full strength for text. Nothing else is colored.
3. **Hairlines and whitespace do the layout.** Cards are `border border-border bg-card shadow-xs` — no heavy
   shadows, no colored panels. Sections are separated by `.separator` or spacing, not backgrounds.
4. **Grid: 4px/8px.** Control heights: 36px (`h-9`) default, 32px small, 40px large. Page container:
   `container-page` (max-w-6xl, 24px gutters). Header 56px sticky with `page-header`.
5. **Type: 14px base**, headings tight (`tracking-tight`), `text-muted-foreground` for secondary text,
   monospace for code/ids. Two weights: 400 and 500/600. No 800/900.
6. **Icons are inline SVG**, 16px (`.icon`), stroke 2, `currentColor` (lucide style). Emoji are content
   (a product image, an empty-state illustration), never UI chrome.
7. **Motion is short**: `fade-up` (250ms) for page content, `pop-in` (150ms) for overlays, `transition-colors`
   on hover. Nothing bounces.
8. **Dark mode is free**: everything uses tokens, so `.dark` on `<html>` flips the whole app (`site` has a toggle).
9. **Components, not one-offs.** Use `.btn .btn-primary|secondary|outline|ghost|destructive [.btn-sm|.btn-lg|.btn-icon]`,
   `.input`, `.select`, `.card` (+ `.card-header/.card-title/.card-desc/.card-body`), `.badge-*`, `.table`,
   `.nav-link` / `.nav-link-active`, `.skeleton`, `.kbd`, `.separator`, `.muted`, `.link`. Add Tailwind utilities
   for layout (flex/grid/gap/spacing) and typography only.
