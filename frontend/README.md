# vfinancy Frontend

React + TypeScript desktop UI. Built with Vite, styled with TailwindCSS + Shadcn UI primitives. Talks to the Go backend exclusively through Wails bindings (no HTTP).

## Stack

- **React 18** + **TypeScript 5**
- **Vite 5** for dev server / build
- **TailwindCSS 3** + `tailwindcss-animate` for styling
- **Shadcn UI** patterns (Radix primitives + `class-variance-authority`)
- **React Router 6** for client routing
- **TanStack Query 5** for server-state caching
- **Zustand 4** for local UI / session state (with `persist` middleware)
- **React Hook Form 7** + **Zod 3** for forms / validation
- **lucide-react** for icons
- **recharts** for charts
- **@fontsource/inter** for self-hosted Inter (latin subset)

## Phase 0.5 — UI/UX Foundation

The current state of the frontend is the **UI/UX foundation** — full Spanish (es-PE) shell with all module placeholders, design system, theme, mock data. **No business logic or DB connections yet.** See `../DESIGN_SYSTEM.md` for the canonical design rules.

## Folder Structure

```
src/
  components/        # category-folders, not by feature
    button/          # Button + variants (default/secondary/outline/ghost/link/destructive/success)
    input/           # Input, Textarea, Field, Label, SearchInput
    select/          # Select (Radix)
    checkbox/        # Checkbox, RadioGroup, Switch
    table/           # DataTable, TablePagination
    dialog/          # Dialog, AlertDialog (5 variants), ConfirmDialog
    card/            # Card, StatCard
    badge/           # Badge + 3 status-specific variants
    navigation/      # Sidebar (collapsible), Topbar, Breadcrumbs
    feedback/        # Spinner, Skeleton, ProgressBar, EmptyState, ErrorState, Toaster
    charts/          # LineChart, BarChart, PieChart (recharts wrappers)
    layout/          # PageContainer, PageHeader, Section, Stack, Grid
    money/           # MoneyInput, MoneyDisplay
    misc/            # DropdownMenu, Separator, Tooltip
  pages/             # route screens (Dashboard + 1 per module + Login)
  layouts/           # AppLayout (sidebar + topbar + breadcrumbs)
  stores/            # Zustand: theme, session, sidebar, ui, notification
  locales/           # es-PE translation dictionary + t() helper
  lib/               # cn(), formatCurrency, formatDate, nav routes
  data/              # deterministic mock data (seeded RNG, identical every load)
  hooks/             # (reserved)
  services/          # (reserved for Wails bindings in Phase 1+)
  types/             # (reserved)
  assets/            # static files
  main.tsx           # Vite entrypoint
  App.tsx            # router config
  index.css          # Tailwind + design tokens + Inter font
```

Every `components/<category>/` has an `index.ts` barrel — **import from `@/components/<category>`**, not from individual files.

## Commands

```bash
npm install         # install dependencies
npm run dev         # Vite dev server (frontend only)
npm run build       # production build → dist/ (Wails embeds this)
npm run check       # tsc --noEmit (type check)
```

Wails-specific:

```bash
wails dev           # run Vite + Go together, hot-reload
wails build         # produce desktop binary in build/bin/
```

## Conventions

- The frontend **must not** access the database directly. All calls go through Wails bindings exposed by the Go `App` and `bindings.App` structs.
- All UI text is in **Spanish (es-PE)** via `t('key')` from `@/locales`. No hardcoded strings in components.
- All numbers / dates / currency use `Intl.*` helpers in `@/lib/utils`. **Never** use `toFixed` for money or `toLocaleString` ad-hoc.
- Path alias `@/*` resolves to `src/*`.
- Tailwind tokens (CSS variables) live in `src/index.css`. Use `bg-primary`, `text-muted-foreground`, etc. — never hardcode colors.
- Use `cn()` for class composition. Don't write raw string concatenation.
- Destructive actions go through `<AlertDialog variant="destructive">` or `<ConfirmDialog>`.
- Forms use `react-hook-form` + `zod` (when forms are added in later phases).
- After running `wails dev`/`wails build`, real generated Wails types appear in `src/wailsjs/go/main/` — import from those, not from the placeholder `AppBindings` interface in `src/vite-env.d.ts`.
