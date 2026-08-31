# vfinancy — Design Language

This is the canonical visual/interaction spec for the vfinancy desktop UI. Implementation lives in `frontend/src/index.css` (plain CSS3, no Tailwind/PostCSS). The component layer is built on **Base UI** (`@base-ui/react`, unstyled, accessible parts) — we style Base UI parts directly with lean CSS classes and the library's `data-*` state attributes.

---

## 1. Foundations

### Color tokens (OKLCH)

Full-color custom properties (no channel splitting). Alpha is derived with `color-mix(in oklch, …)` when needed.

The palette is **monochrome/grayscale by design** — a terminal/skeuomorphic-square look. The only chromatic token is the destructive/error red; all other state colors (success/warning/info) are muted pairs never used decoratively.

| Token | Light | Dark | Role |
|---|---|---|---|
| `--color-bg` | `oklch(97% 0 0)` | `oklch(14.5% 0 0)` | App background (dark = the fg color) |
| `--color-surface` | `oklch(100% 0 0)` white | `oklch(14.5% 0 0)` | Cards, popups, inputs |
| `--color-muted` | `oklch(97% 0 0)` | `oklch(26.9% 0 0)` | Hover fills, table head, wells |
| `--color-fg` | `oklch(14.5% 0 0)` | white | Primary text, borders, primary accent |
| `--color-muted-fg` | `oklch(43.9% 0 0)` | `oklch(87% 0 0)` | Secondary text |
| `--color-fg-subtle` | `oklch(55.6% 0 0)` | `oklch(70.8% 0 0)` | Hints, placeholders, disabled |
| `--color-border` / `--color-border-strong` | `oklch(14.5% 0 0)` | white | Full-strength 1px chrome (border = fg) |
| `--color-primary` | `oklch(14.5% 0 0)` | white | **Single accent**: since the palette is monochrome, the accent collapses to fg. Filled primary buttons invert (near-black bg / white text in light; white bg / near-black text in dark). |
| `--color-primary-muted` / `-fg` | `oklch(97% 0 0)` / `oklch(14.5% 0 0)` | `oklch(26.9% 0 0)` / white | Active nav pill, selected rows, soft accents |
| `--color-success/-warning/-destructive/-info` (+`-muted`, `-muted-fg`) | semantic pairs | lightened pairs | Status only — never decorative. `destructive` is the sole chromatic: `oklch(50.5% 0.213 27.518deg)` / `oklch(70.4% 0.191 22.216deg)` |
| `--color-ring` | `oklch(14.5% 0 0)` | white | `:focus-visible` outline |
| `--color-overlay` | `oklch(14.5% 0 0 / 0.3)` | `oklch(14.5% 0 0 / 0.7)` | Dialog/drawer backdrops |

**Rules**
- No hardcoded colors anywhere. Components consume tokens only. Every color has an OKLCH `0 0deg` chroma/hue or a muted semantic pair — there is no free-standing hue.
- The accent is **monochrome**: primary, focus ring, borders, and links all resolve to `--color-fg` / its inversion. Blue/indigo is gone.
- `destructive` is the only chromatic token, and only red is allowed to carry chroma.
- Status colors are rationed to meaning (badges, deltas, destructive confirmations).
- Charts read the same tokens (passed as string props to recharts).

### Typography

- **Geist Sans** (`@fontsource/geist-sans` 400/500/600) for all UI. **Geist Mono** for document numbers and numeric cells (`.tabular`).
- Scale (tokens `--text-*`): 12 / 13 / 14 / 16 / 18 / 22 / 28 px. Body and default control text: 14. Tables and dense UI: 13.
- Headings: weight 600, slight negative tracking on page titles (`-0.015em`).
- Money always renders through `formatCurrency(value, currency)`; dates through `formatDate`. Numbers in tables get `.tabular`.

### Shape, depth, motion

- Radii: `--radius-sm` / `--radius` / `--radius-lg` are **all `0`** (square corners — sharp, terminal-like edges); `--radius-full` (`999px`) is reserved for badges, dots, and circular avatars only.
- Depth is **flat with hard-offset shadows** (no blur): `--shadow-sm` `0.0625rem 0.0625rem 0`, `--shadow-md` `0.125rem 0.125rem 0`, `--shadow-lg` `0.25rem 0.25rem 0` over `rgb(0 0 0 / 12–16%)`. Dark mode uses **no shadow** (`none`) — the monochrome inversion relies on strong borders instead.
- **Layering (ascending z-index):** table sticky header 2 → topbar 40 → drawer 50 → dialog 60 → menu/select 70 → tooltip 80 → toaster 90. Overlays are portals into `body`; only the topbar/table participate in page stacking, so every overlay must sit above the topbar (40).
- Motion: ~160ms for hover/state, ~200ms for dialogs/toasts, 250ms for the drawer. Easing `--ease` (`cubic-bezier(0.25,0.8,0.35,1)`).
- Enter/exit animations use Base UI's `data-starting-style` / `data-ending-style` attributes. All motion is disabled under `prefers-reduced-motion: reduce`.

### Spacing & density

- 8px grid; page gutters 24px (16px < 768px).
- Control heights: `--control-h` **2rem (32px)** (default), `--control-h-sm` 1.875rem (sm), `--control-h-lg` 2.5rem (lg). Default control text is 0.875rem (14px) with 1.25rem line-height.
- Buttons/inputs use `gap 0.5rem`, padding `0 0.75rem` (buttons) / `0 0.5rem` (inputs), matching the reference component sheet.
- CRUD pages: `PageContainer` → `PageHeader` (title/subtitle + Create) → stat-card `Grid` → `DataTable`.

---

## 2. Component system (Base UI)

Base UI parts are styled via `className` + data attributes. Every styled part lives in `index.css` under a numbered section. Never import from `@radix-ui/*` (removed).

| Area | Components (`@/components/…`) | Base UI part |
|---|---|---|
| Actions | `button` — variants: primary, secondary, outline, ghost, destructive; sizes: sm, md, lg, icon, icon-sm; `loading` spinner | `Button` (+ `render` prop for element composition) |
| Inputs | `input` — `Input`, `Textarea`, `Label`, `SearchInput` | native, plain CSS |
| Pickers | `select` — `Select`, `SelectValue`, `SelectTrigger`(`invalid`), `SelectContent`, `SelectItem` | `Select` (Portal → Positioner → Popup; `items` on Root enables labeled trigger values) |
| Overlays | `dialog` — `Dialog`, `DialogContent`(`size` sm/md/lg/xl), Header/Footer/Title/Description; `AlertDialog` (variants: success/warning/destructive/info/confirmation), `ConfirmDialog` (destructive confirm), `CancelDialog`, `RegisterPaymentDialog` | `Dialog` (Portal → Backdrop → Popup + Close) |
| Menus | `misc` — `DropdownMenu*` (items support `onSelect`, `danger`, `inset`; radio groups for theme), `RowActions` (row `⋯` menu from `RowAction[]`), `Tooltip*` (`asChild` → `render` bridged), `Drawer` (controlled side panel w/ swipe) | `Menu`, `Tooltip`, `Drawer` |
| Tabs | `tabs` — `Tabs`, `TabsList`, `TabsTrigger` (`data-active`), `TabsContent` | `Tabs` |
| Feedback | `feedback` — `Spinner`, `EmptyState` (icon+title+description+action), `ErrorState`, `Toaster` (Base UI Toast; imperative API `useNotificationStore.getState().push({title, description?, variant, duration?})`) | `Toast` (Provider/Root/Title/Description/Close/Viewport) |
| Data | `table` — `DataTable<T>` + `Column<T>`, `TablePagination` | plain table + our parts |
| Layout | `layout` — `PageContainer`, `PageHeader`, `Section`, `Grid`; helpers `.stack`, `.hstack`, `.grid-N` | divs |
| Display | `card` — `Card`, Header/Title/Description/Content, `StatCard`; `badge` — 8 variants + status badges | divs |

### Data attributes used for state styling

`[data-pressed]`, `[data-disabled]` (Button/Menu); `[data-open]`, `[data-closed]`, `[data-starting-style]`, `[data-ending-style]` (Dialog/Menu/Select popups, Toast, Drawer); `[data-highlighted]` (Menu.Item, Select.Item); `[data-selected]` (Select.Item); `[data-popup-open]` (Select.Trigger); `[data-active]` (Tabs.Tab); `aria-invalid` (inputs).

---

## 3. App shell & navigation

- `AppLayout`: left `Sidebar` + main column (`Topbar`, `Outlet`) + mobile `Drawer` + `Toaster`.
- **Sidebar**: flat list only (no nesting), icon + label, active state = `--color-primary-muted` pill; collapses to icon rail (persisted in `useSidebarStore`); tooltips (right) only when collapsed; collapse toggle in footer.
- **Responsive**: < 1100px auto-collapses; < 768px the sidebar hides and a Topbar hamburger opens it inside a `Drawer` (`sidebar--mobile`).
- **Topbar**: global search, theme menu (light/dark/system radio), notifications bell (unread count badge), lock button (only when a local password is configured).
- Active route/`end` semantics come from `lib/nav.ts`; **Configuración is always the last item**.

---

## 4. Entry flows

- **Not configured** → `/configuracion-inicial` wizard: 3 steps (Empresa → Regional → Acceso), step list + "paso N de 3" text indicator (never a progress bar or percentage), Back enabled from step 2, per-step zod validation, single `SetupWorkspace` submit.
- **Configured + password set + locked** → `/bienvenida`: full-screen card with the "vfinancy" text logo and the password form (unlock). 
- **Configured, no password (or unlocked)** → straight into the app.
- Lock is available from the Topbar and from Settings → Seguridad ("Bloquear ahora").

---

## 5. CRUD pages (standard)

Every module page follows the same skeleton:

1. `PageHeader` — title, subtitle, **Create** button.
2. Optional stat-card `Grid`.
3. `DataTable` toolbar — `SearchInput` + relevant `Select` filters (left), optional secondary actions (right).
4. Table — column sorting (click header), pagination (page-size select + pages), sticky first column, loading skeleton, error state with retry.
5. **Row actions** — `RowActions` menu (`⋯`): Edit / domain actions / destructive Delete last, in `--danger` styling.
6. **Create & Edit share one dialog** per entity (feature-local `*FormDialog`) with RHF + zod validation, loading (submit spinner, disabled cancel), error (toast) and success (toast + close) states.
7. **Delete** uses `ConfirmDialog` (destructive `AlertDialog`) that names the record and states irreversibility.
8. **Empty states** carry a useful message and a Create CTA when the list is empty (`action` prop).

Feature settings: only Inventory has feature-scoped settings → a **Drawer** ("Reglas": clearance days + warning days) with an explicit **Save** button. Everything else is app-wide and lives in Configuración.

---

## 6. General Settings (`/configuracion`)

Sections, each with an explicit **Save** (or explicit action buttons):

1. **Empresa** — fiscal info (`updateBusinessInfo`).
2. **Operaciones** — document number prefixes (sale / purchase / journal).
3. **Apariencia y perfil** — theme (applies instantly, persisted on Save) + read-only profile info.
4. **Seguridad** — create/update/remove the local password (Argon2id-backed) + "Bloquear ahora".

---

## 7. States

| State | Pattern |
|---|---|
| Loading (page) | `.page-loader` centered `Spinner` |
| Loading (table) | 5 skeleton rows (`.skel-cell`) |
| Loading (button) | `loading` prop: spinner + disabled |
| Loading (form/dialog) | submit `loading`; cancel disabled |
| Empty | `EmptyState` — icon, message, optional Create CTA |
| Error (query) | `ErrorState` with retry (also DataTable `error` prop) |
| Error (mutation) | destructive toast with backend message |
| Success | success toast, dialog closes |
| Danger confirm | `ConfirmDialog` — names the record, explains consequences, "Eliminar/Anular" in destructive button |

---

## 8. Accessibility

- Focus: `:focus-visible` 2px `--color-ring` outline, 2px offset (inputs swap to border+shadow ring).
- Dialogs/menus/selects/drawer: focus trap, ESC close, portal rendering, focus return — provided by Base UI.
- Icon-only buttons require `aria-label`; destructive actions require explicit confirmation; toasts are polite live regions.
- Contrast: body text ≥ 4.5:1 in both themes; muted text reserved for secondary info.
- `prefers-reduced-motion` disables all transitions/animations.

---

## 9. i18n & formatting

- UI copy is Spanish (es-PE). `t()` from `@/locales` is used in shared components; page-level strings are inline Spanish today — migrate opportunistically, never block a redesign change on it.
- All money via `formatCurrency(value, 'PEN' | 'USD')`, dates via `formatDate`, percentages via `formatPercent`, quantities via `formatNumber`. No `toFixed`, no ad-hoc `toLocaleString`.

---

## 10.CSS architecture (`src/index.css`)

Single self-contained stylesheet, numbered sections:

1. Tokens (`:root` / `.dark`) — 2. Base/reset + utilities — 3. Button — 4. Input/Label/Field — 5. Card/StatCard — 6. Badge — 7. Dialog — 8. Menu — 9. Tooltip — 10. Select — 11. Tabs — 12. Drawer — 13. NumberField/LineItems — 14. Toast — 15. Spinner/Skeleton/Empty/Error — 16. App shell (sidebar/topbar) — 17. Page layout — 18. Forms — 19. DataTable — 20. Setup wizard — 21. Welcome — 22. Misc/crash screen — 23. Animations & accessibility — 24. Responsive shell.

Rules: semantic class names (BEM-ish `.block__element--modifier`), no utility frameworks, no per-component CSS files, one-off inline `style={{}}` allowed sparingly. Class join via `cx()` from `@/utils/cx`.
