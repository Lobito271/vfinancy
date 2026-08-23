---
version: alpha
name: vfinancy ERP Design Language
description: An analysis of the vfinancy ERP desktop design language — a dense enterprise-information interface built from HSL semantic tokens on a 4px grid. Neutral slate chrome carries data surfaces; brand ink-and-paper polarity (`{colors.primary}` / `{colors.background}`) flips between light and dark; the four semantic status colors are rationed to meaning only (success / warning / destructive / info); and a single primary token inverts between modes. Inter at a 14px body, tight-tracked semibold headings, an 8px radius base, 150/200/300ms transitions, and Radix-driven enter/exit motion powered by `vf-enter` / `vf-exit` custom properties.

colors:
  primary: "#0f172a"          # Ink — action fill in light, inverts to paper in dark
  primary-foreground: "#f8fafc"
  background: "#ffffff"
  foreground: "#0f172a"
  card: "#ffffff"
  card-foreground: "#0f172a"
  popover: "#ffffff"
  popover-foreground: "#0f172a"
  muted: "#f1f5f9"
  muted-foreground: "#64748b"
  accent: "#f1f5f9"
  accent-foreground: "#0f172a"
  secondary: "#f1f5f9"
  secondary-foreground: "#0f172a"
  success: "#22c55e"
  success-foreground: "#ffffff"
  warning: "#f59e0b"
  warning-foreground: "#ffffff"
  destructive: "#ef4444"
  destructive-foreground: "#f8fafc"
  info: "#0ea5e9"
  info-foreground: "#ffffff"
  border: "#e2e8f0"
  input: "#e2e8f0"
  ring: "#0f172a"
  dark-background: "#020617"
  dark-foreground: "#f8fafc"
  dark-card: "#0b111e"
  dark-popover: "#0b111e"
  dark-muted: "#131a25"
  dark-muted-foreground: "#94a3b8"
  dark-accent: "#18212f"
  dark-secondary: "#18212f"
  dark-border: "#1d283a"
  dark-input: "#1d283a"
  dark-ring: "#cbd5e1"
  dark-primary: "#f8fafc"
  dark-primary-foreground: "#0f172a"
  dark-destructive: "#7f1d1d"

typography:
  display:
    fontFamily: Inter
    fontSize: 30px
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: -0.015em
  page-title:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: 600
    lineHeight: 1.33
    letterSpacing: -0.015em
  section-title:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: 600
    lineHeight: 1.4
    letterSpacing: -0.015em
  card-title:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: 600
    lineHeight: 1.55
    letterSpacing: -0.015em
  body:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.43
    letterSpacing: 0
  table:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: 400
    lineHeight: 1.54
    letterSpacing: 0
  label:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: 500
    lineHeight: 1.54
    letterSpacing: 0
  caption:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 500
    lineHeight: 1.33
    letterSpacing: 0.025em
  numeric:
    fontFamily: Inter
    fontSize: 30px
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: -0.015em
    fontVariantNumeric: tabular-nums
  mono:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.33
    letterSpacing: 0

rounded:
  none: 0px
  sm: 4px
  md: 6px
  lg: 8px
  xl: 12px
  2xl: 16px
  full: 9999px

spacing:
  px: 1px
  xxs: 2px
  xs: 4px
  sm: 8px
  md: 12px
  lg: 16px
  xl: 24px
  xxl: 32px
  section: 48px

components:
  button-default:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    typography: "{typography.label}"
    rounded: "{rounded.md}"
    height: 40px
    padding: "{spacing.lg} {spacing.lg}"
    hover: "hsl(var(--primary) / 0.9)"
  button-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.secondary-foreground}"
    typography: "{typography.label}"
    rounded: "{rounded.md}"
    height: 40px
    padding: "{spacing.lg}"
    hover: "hsl(var(--secondary) / 0.8)"
  button-outline:
    backgroundColor: "{colors.background}"
    borderColor: "{colors.input}"
    textColor: "{colors.foreground}"
    typography: "{typography.label}"
    rounded: "{rounded.md}"
    height: 40px
    padding: "{spacing.lg}"
    hover: "accent wash"
  button-ghost:
    backgroundColor: transparent
    textColor: "{colors.foreground}"
    typography: "{typography.label}"
    rounded: "{rounded.md}"
    height: 40px
    padding: "{spacing.lg}"
    hover: "accent wash"
  button-link:
    backgroundColor: transparent
    textColor: "{colors.primary}"
    typography: "{typography.label}"
    underline-offset: 4px
    hover: "underline"
  button-destructive:
    backgroundColor: "{colors.destructive}"
    textColor: "{colors.destructive-foreground}"
    typography: "{typography.label}"
    rounded: "{rounded.md}"
    height: 40px
    padding: "{spacing.lg}"
    hover: "hsl(var(--destructive) / 0.9)"
  button-success:
    backgroundColor: "{colors.success}"
    textColor: "{colors.success-foreground}"
    typography: "{typography.label}"
    rounded: "{rounded.md}"
    height: 40px
    padding: "{spacing.lg}"
    hover: "hsl(var(--success) / 0.9)"
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.card-foreground}"
    borderColor: "{colors.border}"
    rounded: "{rounded.lg}"
    padding: "{spacing.xl}"
    shadow: subtle
  card-title:
    typography: "{typography.card-title}"
    color: "{colors.card-foreground}"
  card-description:
    typography: "{typography.table}"
    color: "{colors.muted-foreground}"
  stat-card:
    backgroundColor: "{colors.card}"
    borderColor: "{colors.border}"
    rounded: "{rounded.lg}"
    padding: "{spacing.xl}"
    valueTypography: "{typography.numeric}"
    trendUp: "{colors.success}"
    trendDown: "{colors.destructive}"
    trendFlat: "{colors.muted-foreground}"
  badge-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.full}"
    padding: "{spacing.xs} {spacing.md}"
  badge-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.secondary-foreground}"
    rounded: "{rounded.full}"
    padding: "{spacing.xs} {spacing.md}"
  badge-success:
    backgroundColor: "hsl(var(--success) / 0.15)"
    textColor: "{colors.success}"
    rounded: "{rounded.full}"
    padding: "{spacing.xs} {spacing.md}"
  badge-warning:
    backgroundColor: "hsl(var(--warning) / 0.15)"
    textColor: "{colors.warning}"
    rounded: "{rounded.full}"
    padding: "{spacing.xs} {spacing.md}"
  badge-destructive:
    backgroundColor: "hsl(var(--destructive) / 0.15)"
    textColor: "{colors.destructive}"
    rounded: "{rounded.full}"
    padding: "{spacing.xs} {spacing.md}"
  badge-info:
    backgroundColor: "hsl(var(--info) / 0.15)"
    textColor: "{colors.info}"
    rounded: "{rounded.full}"
    padding: "{spacing.xs} {spacing.md}"
  badge-muted:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.muted-foreground}"
    rounded: "{rounded.full}"
    padding: "{spacing.xs} {spacing.md}"
  input:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    borderColor: "{colors.input}"
    typography: "{typography.table}"
    rounded: "{rounded.md}"
    height: 40px
    padding: "{spacing.sm} {spacing.md}"
    placeholder: "{colors.muted-foreground}"
    focus: "ring in {colors.ring} + --vf-offset 1px"
    invalid: "destructive border + ring"
  field-label:
    typography: "{typography.label}"
    color: "{colors.foreground}"
    requiredMarker: "{colors.destructive}"
  field-description:
    typography: "{typography.caption}"
    color: "{colors.muted-foreground}"
  field-error:
    typography: "{typography.caption}"
    color: "{colors.destructive}"
  select-trigger:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    borderColor: "{colors.input}"
    typography: "{typography.table}"
    rounded: "{rounded.md}"
    height: 40px
    padding: "{spacing.sm} {spacing.md}"
  select-content:
    backgroundColor: "{colors.popover}"
    textColor: "{colors.popover-foreground}"
    rounded: "{rounded.md}"
    shadow: medium
    borderColor: "{colors.border}"
    maxHeight: 384px
  data-table:
    container: "8px radius, 1px border"
    headerBackground: "{colors.muted}"
    headerTypography: "{typography.caption}"
    headerColor: "{colors.muted-foreground}"
    headerHeight: 40px
    headerPadding: "{spacing.lg} {spacing.lg}"
    cellPadding: "{spacing.md} {spacing.lg}"
    rowBorder: "{colors.border}"
    rowHover: "muted wash at 30%"
    stickyHeader: true
  dialog-overlay:
    backgroundColor: "rgba(0, 0, 0, 0.6)"
    backdropFilter: "blur(4px)"
  dialog-content:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.lg}"
    padding: "{spacing.xl}"
    shadow: large
    sizes: "sm 384px · md 448px · lg 672px · xl 1024px"
  alert-dialog:
    success: "icon CheckCircle2 {colors.success} on success tint at 10%"
    warning: "icon AlertTriangle {colors.warning} on warning tint at 10%"
    destructive: "icon XCircle {colors.destructive} on destructive tint at 10%"
    info: "icon Info {colors.info} on info tint at 10%"
    confirmation: "icon AlertTriangle {colors.primary}"
  sidebar:
    backgroundColor: "{colors.card}"
    borderColor: "{colors.border}"
    widthCollapsed: 64px
    widthExpanded: 256px
    itemTypography: "{typography.table}"
    itemColor: "{colors.muted-foreground}"
    itemActive: "accent wash, accent-foreground text, medium weight"
    itemPadding: "{spacing.sm} {spacing.md}"
    itemRounded: "{rounded.md}"
    brand: "vfinancy wordmark link, primary color"
  topbar:
    backgroundColor: "background token at 95% with backdrop blur"
    borderColor: "{colors.border}"
    height: 56px
    padding: "{spacing.lg}"
  widget-shell:
    backgroundColor: "{colors.card}"
    borderColor: "{colors.border}"
    rounded: "{rounded.lg}"
    shadow: subtle
    padding: "{spacing.xl}"
  money-input:
    typography: "{typography.table}"
    tabularNums: true
    textAlign: right
    currencyPrefixColor: "{colors.muted-foreground}"
    prefixPaddingLeft: "{spacing.md}"

---

## Overview

vfinancy is a **dense enterprise information surface**, not a marketing page. Every pixel is spent on legibility, density, and the reliable signaling of business state. Where a consumer site leads with photography or decoration, this interface leads with **structured data**: tables, KPIs, forms, and status. The whole system is built on a small set of *semantic tokens* — `{colors.primary}`, `{colors.background}`, `{colors.muted}`, `{colors.border}`, `{colors.success}` / `{colors.warning}` / `{colors.destructive}` / `{colors.info}` — exposed as HSL channel triplets in `src/index.css` and consumed by the semantic component classes defined in that same file (`.btn--primary`, `.badge--success`, `.input--invalid`) plus a mirrored TypeScript registry in `src/design-system/`. There is no hardcoded color anywhere; both representations are kept synchronized.

The system runs on **ink-and-paper polarity**. In light mode the canvas is `{colors.background}` paper and the action/ink token `{colors.primary}` is near-black; in dark mode they **invert** — `{colors.primary}` becomes paper and the canvas becomes `{colors.dark-background}` ink. This single polarity flip is what makes dark mode consistent: semantic accents (`{colors.success}`, `{colors.destructive}`) hold their hue in both modes, while all neutral and action surfaces swap. Hover states are always **alpha dims of the fill** (`hsl(var(--primary) / 0.9)`, `hsl(var(--secondary) / 0.8)`) rather than new colors — an intentional economy that keeps the palette small and predictable.

Status color is **rationed to meaning**. Success is paid/active/in-stock; warning is pending/low-stock; destructive is cancelled/blocked/error; info is partial/neutral-progress. Badges render these as translucent tints (`hsl(var(--success) / 0.15)` fill with `hsl(var(--success))` text) so they read as tags on any surface rather than solid blocks competing with buttons. The four accents are never used decoratively.

Motion is restrained and Radix-native: overlays and popovers animate through the `vf-enter` / `vf-exit` keyframes (driven by `--vf-enter-*` / `--vf-exit-*` custom properties), with 150ms micro-interactions, 200ms dialog transitions, and 300ms page transitions, all disabled under `prefers-reduced-motion`.

**Key Characteristics:**
- **Token-driven neutrals.** Every surface is a semantic token (`card`, `popover`, `muted`, `accent`); text is `foreground` / `muted-foreground`; structure is `border` / `input`. No literal hex in components.
- **4px spacing grid.** Component paddings, heights, and gaps are multiples of 4px (0.25rem), enforced by `src/design-system/spacing.ts` and baked into the semantic classes.
- **8px radius base.** 6px (`--radius-sm`) for controls, 8px (`--radius`) for cards/dialogs, full only for badges and avatars. Sharpness is the default mood.
- **Alpha-dim hovers.** Hover fills are alpha dims of the base token (`hsl(var(--primary) / 0.9)`, `hsl(var(--accent) / 0.5)`, `hsl(var(--muted) / 0.3)`) — depth on hover is a tint change, never a new color.
- **Uppercase caption chrome.** Table headers and small labels use `{typography.caption}` (12px, 500, +0.025em tracking, uppercase) — the "column legend" voice of the data grid.
- **Tabular numerics.** Money, quantities, and KPI values always render with the `.tabular` class; money goes through `formatCurrency(value, 'PEN')` and dates through `formatDate(value)` — never `toFixed` or ad-hoc `toLocaleString`.
- **Layered motion via custom properties.** Radix open/close animations compile into `vf-enter` / `vf-exit`; focus rings use `--vf-offset` / `--vf-offset-color`.
- **Dense data chrome.** `DataTable` owns sorting, filtering, column visibility, selection, pagination, and CSV export; the sticky uppercase header sits on a translucent muted wash over a rounded bordered container (`.datatable`).

## Colors

The palette is **two polar modes over one shared semantic set**. All tokens are stored as HSL channel triplets (e.g. `--primary: 222 47% 11%`) so any class can apply opacity with `hsl(var(--primary) / 0.9)`. The light and dark variants live in `:root` and `.dark` in `src/index.css`; `src/design-system/colors.ts` mirrors them for TypeScript.

### Mode Polarity (Light)
- **Canvas** — `{colors.background}` (#ffffff) paper behind everything.
- **Ink / Action** — `{colors.primary}` (#0f172a): the near-black slate that fills the primary button and the sidebar logo plate. Doubles as the focus ring `{colors.ring}`.
- **Foreground** — `{colors.foreground}` (#0f172a) primary text on light; `{colors.muted-foreground}` (#64748b) secondary text, labels, placeholders.
- **Structure** — `{colors.border}` / `{colors.input}` (#e2e8f0) hairlines for cards, tables, and inputs.
- **Tint fills** — `{colors.muted}` / `{colors.accent}` / `{colors.secondary}` (#f1f5f9) for table headers, hover washes, and ghost-hover states.

### Mode Polarity (Dark)
- **Canvas** — `{colors.dark-background}` (#020617) ink canvas.
- **Inverted Action** — `{colors.dark-primary}` (#f8fafc): *paper becomes the action fill* so the primary button still reads as "the one button." Text on it is `{colors.dark-primary-foreground}` (#0f172a).
- **Surfaces** — `{colors.dark-card}` / `{colors.dark-popover}` (#0b111e) one step above canvas; `{colors.dark-muted}` / `{colors.dark-accent}` / `{colors.dark-secondary}` (#18212f) for headers and hovers; `{colors.dark-border}` (#1d283a) hairlines.
- **Text** — `{colors.dark-foreground}` (#f8fafc); `{colors.dark-muted-foreground}` (#94a3b8).
- **Ring** — `{colors.dark-ring}` (#cbd5e1) focus ring, lighter than the light-mode ring so it reads on dark.

### Semantic (same hue in both modes)
- **Success** (`{colors.success}` — #22c55e): paid, active, in-stock, positive trends.
- **Warning** (`{colors.warning}` — #f59e0b): pending, low-stock, caution.
- **Destructive** (`{colors.destructive}` — #ef4444, #7f1d1d in dark): cancelled, blocked, errors, sign-out.
- **Info** (`{colors.info}` — #0ea5e9): partial payments, neutral progress, guidance.

### Brand & Character
- **`vfinancy` brand wordmark** — the app name as a primary-colored link at the top of the sidebar; user initials render in a 28px circular avatar in the topbar.
- **Destructive is rationed** to actions that destroy or block (delete confirmations, cancelled sales, blocked customers, sign-out). It is never a "brand" color.

## Typography

### Font Family
**Inter** is the entire UI voice (self-hosted via `@fontsource/inter`). No display face, no serif — character comes from *weight and size contrast*, not typeface. **JetBrains Mono** (`{typography.mono}`) is reserved for codes and identifiers (SKU, document numbers, chart-of-accounts codes) where monospace alignment matters. Body copy is 14px Inter; tables and dense controls drop to 13px.

### Hierarchy

| Token | Size | Weight | Line Height | Tracking | Use |
|---|---|---|---|---|---|
| `{typography.display}` | 30px | 600 | 1.2 | -0.015em | KPI values (`tabular-nums`), dashboard stat value |
| `{typography.page-title}` | 24px | 600 | 1.33 | -0.015em | Page headers (`PageHeader`) |
| `{typography.section-title}` | 20px | 600 | 1.4 | -0.015em | Section titles in a page |
| `{typography.card-title}` | 18px | 600 | 1.55 | -0.015em | Card titles, dialog titles |
| `{typography.body}` | 14px | 400 | 1.43 | 0 | Default UI text |
| `{typography.table}` | 13px | 400 | 1.54 | 0 | Table cells, inputs, buttons, sidebar items |
| `{typography.label}` | 13px | 500 | 1.54 | 0 | Form labels, button text |
| `{typography.caption}` | 12px | 500 | 1.33 | +0.025em | Table headers (uppercase), metadata, timestamps |
| `{typography.numeric}` | 30px | 600 | 1.2 | -0.015em | KPI / money display (`tabular-nums`) |

### Principles
- **Tight tracking on display levels only.** `-0.015em` on 18px and up; body and table text are optically neutral.
- **Uppercase + tracking = data-column voice.** Table headers use `{typography.caption}` uppercase — the grid's legend — never sentence-case body text.
- **Semibold, not bold.** The weight ceiling is 600; bold (700) exists in the token set but is used sparingly (active nav item).
- **Money and dates are formatted, never computed.** `formatCurrency(value, 'PEN')`, `formatDate(value)`, `formatPercent` from `@/utils/format` (Intl, es-PE locale); tabular numerals are mandatory in every money column and KPI.
- **Dialog titles follow `{typography.card-title}`** with tight tracking, matching card and modal hierarchy.

## Layout

### Spacing System
- **Base unit**: 4px — the token ladder `{spacing.xxs}` 2px · `{spacing.xs}` 4px · `{spacing.sm}` 8px · `{spacing.md}` 12px · `{spacing.lg}` 16px · `{spacing.xl}` 24px · `{spacing.xxl}` 32px · `{spacing.section}` 48px, matching `src/design-system/spacing.ts`.
- Controls are 40px tall; small buttons 32px, large 44px. Inputs 12px horizontal padding, cards 24px padding, table cells `16px 12px`, dialog padding 24px.
- **Semantic sizes** come from `semanticSpacing`: input height 40px, card padding 24px, table cell `16px 12px`, sidebar 64/256px, topbar 56px, page max-width 80rem.

### Grid & Container
- **App shell** (`.app-shell`): fixed `Sidebar` (`.sidebar`, collapsed 64px / expanded 256px, card background with right hairline) + sticky `Topbar` (`.topbar`, 56px, translucent background wash) + scrollable main.
- **Page**: `PageContainer` (`.page-container` — centered, max-width 80rem, 24px padding, 24px vertical rhythm between blocks), then `PageHeader` (title + subtitle + actions), then `Section` blocks and `Grid`.
- **Responsive grid**: `Grid` collapses from 1 column on mobile to 2 on `sm` and the target count on `lg`/`xl`. Dashboard `DashboardGrid` (`.widget-grid`) uses a 12-column pattern where each widget spans 12/6/4/3 columns by size.
- **Table**: a bordered rounded container with a sticky uppercase header and a `TablePagination` footer (showing X–Y of Z, page-size select 10/25/50/100).

### Breakpoints
| Name | Min Width | Key Changes |
|---|---|---|
| `sm` | 640px | Grids go 2-col; page headers row-align; dialog footers turn horizontal |
| `md` | 768px | Tablet vertical |
| `lg` | 1024px | Grids go 3–5-col; sidebar expanded by default |
| `xl` | 1280px | Desktop target (app optimized for 1280×800 min) |
| `2xl` | 1536px | Wide desktop (1920px) |

Mobile-first `min-width` queries; below `sm` everything stacks to one column and the sidebar collapses to icons with hover tooltips.

### Whitespace Philosophy
Whitespace is **reading room for data**, not decoration. Page containers give 24px gutters, cards 24px internal padding, and tables 12px row height — dense but never cramped. The goal is maximum rows per screen while keeping every cell legible.

### Responsive Strategy
- **Density-preserving collapse**: sidebar → icon rail (64px) with tooltips; page grids stack to one column; table toolbars wrap.
- **Sticky chrome**: topbar `sticky top-0`, table header `sticky`, sidebar independently scrollable — the header/footer chrome never scrolls out of reach.
- **Touch targets**: buttons and inputs are ≥ 32px tall (40px default), meeting the 44px target when icon buttons and small controls are used in touch contexts.

### Image Behavior
No photography. Charts (recharts wrappers: `LineChart`, `BarChart`, `PieChart`) render with theme-aware colors (`hsl(var(--border))` grid, `hsl(var(--primary))` series) and adapt to light/dark automatically. Skeleton loaders (`Skeleton`, `Spinner`, `ProgressBar`) are used during data fetches; `EmptyState` and `ErrorState` fill result-less and failed states.

## Elevation & Depth

Depth is **surface layering via token tints and hairline borders**, not shadow stacks. The system keeps one shadow vocabulary (subtle cards, medium popovers/dropdowns, large dialogs, deepest toasts) from `src/design-system/shadows.ts` and never invents a new elevation color.

| Level | Treatment | Use |
|---|---|---|
| 0 — Canvas | background token | Page surface |
| 1 — Card | card token + border + subtle shadow | Cards, widgets, tables |
| 2 — Popover | popover token + border + medium shadow, 400px max | Selects, dropdowns, menus |
| 3 — Dialog | background token + border + large shadow over a 60% black blurred overlay | Dialogs, alert dialogs |
| 4 — Toast | deepest shadow at `z-50` | Toasts/notifications |

### Decorative Depth
- **Alpha hovers** (accent wash, muted wash at 30%, primary at 90%) are the tactile cue — a fill change, not a shadow change.
- **Overlay scrim**: 60% black wash with blur dims and blurs the app behind every dialog.
- **Focus rings** are a two-layer box-shadow: `--vf-offset` (1–2px of `--vf-offset-color`, defaulting to `background`) plus a 2px ring in `{colors.ring}` — driven by the custom properties in every interactive class.
- **z-index ladder** is explicit in `src/design-system/zIndex.ts`: dropdown 1000 → topbar 1300 → drawer 1500 → modal 1700 → popover 1800 → tooltip 1900 → toast 2000 → notification 2100.

## Shapes

### Border Radius Scale
| Token | Value | Use |
|---|---|---|
| `{rounded.none}` | 0px | Chips, dividers, tight cells |
| `{rounded.sm}` | 4px | Small chips, skeleton blocks |
| `{rounded.md}` | 6px | Buttons, inputs, selects, list items |
| `{rounded.lg}` | 8px | Cards, dialogs, tables, stat cards |
| `{rounded.xl}` | 12px | Larger framed surfaces |
| `{rounded.full}` | 9999px | Badges, avatars, switches, radio dots |

`--radius` is `0.5rem` in `src/index.css`; `sm`/`md`/`lg`/`xl`/`2xl` are derived as `calc(var(--radius) ∓ Npx)` in `src/design-system/radius.ts`.

The signature is **functional radius**: controls are subtly rounded (6px), big surfaces rounder (8px), and pill shapes reserved for badges, avatars, and toggles. No decorative blob shapes, no chamfered panels — every radius maps to a physical control class.

### Geometry Rules
- **Buttons**: 6px radius across all variants; icon buttons are exact squares (40×40 / 32×32).
- **Badges**: full-radius pills, 10px/2px padding, 12px text.
- **Cards / dialogs / tables**: 8px radius; dialog sizes `sm` 384px · `md` 448px · `lg` 672px · `xl` 1024px.
- **Avatars / logo plate**: perfect squares/circles, 28px.
- **Inputs**: 6px radius, 40px tall, 12px horizontal padding; invalid state swaps border+ring to destructive.

## Components

> Each spec covers Default and, where it exists, Hover/Active. Variants are separate `components:` entries. No hover state introduces a new color — only alpha dims of the fill.

### Navigation

**`sidebar`** — Primary navigation rail
- Card token fill with right border hairline; collapses 256px → 64px with a 200ms width transition and a ghost toggle button at the foot. Brand wordmark link in a 56px header. Nav rows: `{typography.table}` `{colors.muted-foreground}`, 6px radius, 12px/8px padding; active row = accent wash, accent-foreground text, medium weight, with a 6px primary dot at the rail end; hover = accent wash at 50%. Items without the user's `*.view` permission are hidden via `<Can>`. Collapsed rows show a right-side Radix tooltip on hover.

**`topbar`** — Global command strip
- Sticky 56px translucent background wash with bottom hairline; 16px horizontal padding. Left: `SearchInput` (global search, max-width 36rem). Right: theme dropdown (light/dark/system with radio items), notification bell (badge counter, 20rem dropdown), a 1px vertical divider, and the user menu — 28px round initials avatar + name/company + logout in destructive color. `z-30`.

### Buttons

**`button-default`** — Primary action
- `{colors.primary}` fill, `{colors.primary-foreground}` text in `{typography.label}`, 6px radius, 40px tall with 16px horizontal padding and 8px icon gap. Hover: `hsl(var(--primary) / 0.9)`. The single dominant action per surface. Loading swaps in a `Loader2` spinner. Class: `.btn--primary`.

**`button-secondary`** — Neutral action
- `{colors.secondary}` fill, `{colors.secondary-foreground}` text, 6px radius, 40px. Hover: `hsl(var(--secondary) / 0.8)`. For secondary actions beside a primary. Class: `.btn--secondary`.

**`button-outline`** — Bordered action
- Background token fill with a 1px `{colors.input}` border and `{colors.foreground}` text. Hover: accent wash. For less-prominent actions (view, details). Class: `.btn--outline`.

**`button-ghost`** — Toolbar action
- Transparent fill; hover accent wash. Used in topbar icon buttons and tight toolbars. Class: `.btn--ghost`.

**`button-link`** — Inline action
- `{colors.primary}` text with 4px underline offset; hover underline. For in-text actions. Class: `.btn--link`.

**`button-destructive`** — Irreversible action
- `{colors.destructive}` fill, `{colors.destructive-foreground}` text; hover `hsl(var(--destructive) / 0.9)`. Reserved for delete/block/cancel; also the sign-out row in the user menu. Class: `.btn--destructive`.

**`button-success`** — Affirmative action
- `{colors.success}` fill, `{colors.success-foreground}` text; hover `hsl(var(--success) / 0.9)`. Class: `.btn--success`.

Sizes: `sm` 32px · `md` 40px · `lg` 44px · `icon` 40×40 · `icon-sm` 32×32 (classes `--sm`, `--md`, `--lg`, `--icon`, `--icon-sm`). Disabled = 50% opacity, pointer-events none. Composition via `Slot` (asChild); variants are plain CSS classes joined with `cx()`.

### Cards & Widgets

**`card`** — Content surface
- Card token background, 1px border, 8px radius, subtle shadow. Anatomy: `CardHeader` (24px padding, `CardTitle` = `{typography.card-title}`, `CardDescription` = `{typography.table}` `{colors.muted-foreground}`) + `CardContent` (24px padding, no top padding) + `CardFooter`. A `Card` is the base of `StatCard` and the dashboard `WidgetShell`. Classes: `.card`, `.card-header`, `.card-title`, `.card-description`, `.card-content`, `.card-footer`.

**`stat-card`** — KPI tile
- 8px radius, card background, 24px padding, subtle shadow. Header row: label (`{typography.table}` `{colors.muted-foreground}`) + icon (16px, `{colors.muted-foreground}`). Value: `{typography.numeric}` (30px semibold tabular tight). Trend chip (`.trend--up/down/flat`): pill with alpha tint — up success tint with `ArrowUpRight`, down destructive tint with `ArrowDownRight`, flat muted with `Minus`, plus `changeLabel` in `{colors.muted-foreground}`.

**`widget-shell`** — Dashboard widget frame
- Full-height `Card`; header row (`.card-header--row`) with title (`.card-title--sm`) + description + actions; body handles `loading` (`Spinner`), `error` (`.error-text`), or children.

### Inputs & Forms

**`input`** — Text / search field
- Background token fill, 1px `{colors.input}` border, `{typography.table}`, 6px radius, 40px tall, 12px horizontal padding. Placeholder `{colors.muted-foreground}`. Focus: ring in `{colors.ring}` with a 1px `--vf-offset` gap. Invalid: destructive border + ring + `aria-invalid` (`.input--invalid`). Disabled: 50% opacity, cursor not-allowed. `Textarea` shares the chassis (min-height 80px).

**`field-label`** — Form label
- `{typography.label}` `{colors.foreground}` with a `*` required marker in `{colors.destructive}`. Label sits above the input inside the `.field` column (6px gap).

**`field-description` / `field-error`**
- Description: `{typography.caption}` `{colors.muted-foreground}`. Error: `{typography.caption}` `{colors.destructive}` with `role="alert"` (announced to screen readers). Error replaces description when present.

**`select`** — Radix dropdown
- Trigger mirrors `input` (40px, 6px radius, border, 14px text) with a chevron. Content: popover tokens, 1px border, 6px radius, shadow, max-height 24rem, animating `vf-enter` on open. Items: 14px text, 6px vertical padding, checkmark slot at left, semibold selected label.

**`money-input`** — Currency field
- `input` chassis with a currency prefix (`S/` for PEN, `$` for USD) absolutely positioned left, right-aligned tabular numerals; sanitizes non-numeric input. Paired with `MoneyDisplay` for read-only formatted amounts.

**Form framework** — `Form` + 16+ field components
- `TextField`, `NumberField`, `TextareaField`, `CheckboxField`, `MoneyField`, `PercentageField`, `CurrencyField`, `DateField`, `DateRangeField`, `DateTimeField`, `EmailField`, `PhoneField`, `PasswordField`, `SearchField`, `SelectField`, `AsyncSelectField`, plus domain selects (`CustomerSelectField`, `SupplierSelectField`, `ProductSelectField`, `WarehouseSelectField`, `CategorySelectField`, `BrandSelectField`, `TaxSelectField`, `CurrencySelectField`, `DocumentTypeSelectField`). All use React Hook Form + Zod.

### Tables & Data

**`data-table`** — Enterprise grid
- Bordered rounded container (`.datatable-scroll`) with horizontal scroll. Header (`.datatable-table thead`): sticky, translucent muted wash, 40px cells in `{typography.caption}` uppercase `{colors.muted-foreground}`, sortable with `ChevronUp`/`ChevronDown`, optional column-visibility dropdown (`Settings2`) and CSV export (`Download`). Rows: bottom hairline, hover muted wash; selected rows highlighted; voided rows dimmed via `.voided`; loading shows skeleton rows; error shows `ErrorState` with retry; empty shows `EmptyState`. First column can be sticky. State (sort, filter, page, page size, visible columns) persists to `localStorage`.

**`table-pagination`** — Footer pager
- 16px/12px padding (`.table-pagination`), wraps on mobile. Summary "Mostrando X – Y de Z filas" at 14px with foreground numerals; page-size `Select` (10/25/50/100); numbered page buttons with first/last and ellipsis window.

### Dialogs & Feedback

**`dialog`** — Modal surface
- Overlay: 60% black wash with blur (`.dialog-overlay`), fades in/out. Content: background token fill, 8px radius, large shadow, 24px padding (`.dialog-content`), centered, sizes `sm` 384 · `md` 448 · `lg` 672 · `xl` 1024px, animating `vf-enter`/`vf-exit`. Header (`DialogTitle` = `{typography.card-title}`) + description (`{typography.table}` `{colors.muted-foreground}`) + footer (column on mobile, right-aligned row from `sm`). Close button top-right, 16px `X`.

**`alert-dialog`** — Semantic confirm
- Five variants, each with a themed icon chip (`.alert-icon--*`): `success` CheckCircle2 on success tint, `warning` AlertTriangle on warning tint, `destructive` XCircle on destructive tint, `info` Info on info tint, `confirmation` AlertTriangle in `{colors.primary}`. Used for all destructive confirmations (with `ConfirmDialog` wrapper) — never a bare `window.confirm`.

**`badge`** — Status tag
- Pill (`.badge`): full radius, 1px border, 10px/2px padding, 12px medium text. Variants: `primary` (primary fill), `secondary` (secondary fill), `outline` (borderless text), and the four semantic tints `success`/`warning`/`destructive`/`info` (15% tint fill + solid text) plus `muted`. Domain wrappers: `SaleStatusBadge` (paid→success, pending→warning, partial→info, cancelled→destructive), `CustomerStatusBadge` (active/inactive/blocked), `StockBadge` (inStock/lowStock/outOfStock/clearance).

**`toaster` / `spinner` / `skeleton` / `progress-bar` / `empty-state` / `error-state`**
- Toasts: deep shadow, top-right stack (`.toaster`), auto-dismiss, types success/info/warning/error. Loading: `Spinner`, `Skeleton` (pulse over muted token), `ProgressBar`. Result-less data → `EmptyState` (icon + title + description + optional action); failed → `ErrorState` (destructive tone + retry).

### Charts

**`line-chart` / `bar-chart` / `pie-chart`** — recharts wrappers
- Theme-aware: grid and axis lines use `hsl(var(--border))`, tick labels `hsl(var(--muted-foreground))`, tooltip `hsl(var(--popover))`, primary series `hsl(var(--primary))` with 10% alpha fills. Adapt automatically to light/dark since they read the CSS variables.

## Do's and Don'ts

### Do
- Build every surface from **semantic tokens** consumed through the component classes in `src/index.css` — never hardcode a hex in a component.
- Make hover states **alpha dims of the fill** (`hsl(var(--x) / N)`), never brand-new colors.
- Format money with `formatCurrency(value, 'PEN')` and dates with `formatDate(value)`, always tabular numerals (`.tabular`) in money columns and KPIs — never `toFixed` or ad-hoc `toLocaleString`.
- Let **status color carry meaning only**: success/warning/destructive/info are semantic states, not decoration. Badges use the tinted variants.
- Give destructive actions a **destructive affordance** — `Button variant="destructive"` or `AlertDialog`/`ConfirmDialog`, never a bare confirm.
- Keep the **uppercase caption voice** for table headers and metadata (`{typography.caption}`).
- Use `DataTable`, `Form`, and the category component barrels (`@/components/<categoría>`) instead of hand-rolled tables/forms — and import from the barrel, never the individual file.
- Keep dialogs on the four size steps (384/448/672/1024px) and motion on the three durations (150/200/300ms), respecting `prefers-reduced-motion`.

### Don't
- Don't introduce a new color or shadow for hover/active states — the system's economy is alpha dims and the four semantic accents only.
- Don't break the **ink-paper polarity**: in dark mode `primary` is paper and canvas is ink; adding a saturated "dark primary" breaks theme consistency.
- Don't put raw money/date formatting in components — every number goes through `@/utils/format` with the es-PE locale.
- Don't use badges as solid color blocks that compete with buttons; the 15% tints are the badge voice.
- Don't add utility-style class names (`bg-*`, `text-sm`, `p-4`, `flex items-center gap-2`) — use the semantic component classes and layout helpers (`.stack`, `.hstack`, `.grid-N`); one-off values go in `style={{...}}`.
- Don't soften the system into uniformly rounded cards; radius is functional (controls 6px, surfaces 8px, pills only for badges/avatars/toggles).
- Don't hide permission-lacking buttons — **disable them**; the sidebar hides items without `*.view`, but in-page actions stay visible-but-disabled.
- Don't render emojis or non-localized strings; all UI text goes through `t('key')` (es-PE) and icons come from `@/design-system/icons`, never imported ad-hoc from `lucide-react`.
