# vfinancy Frontend

React + TypeScript desktop UI. Built with Vite, styled with a hand-rolled plain CSS3 system in `src/index.css`. Talks to the Go backend exclusively through Wails bindings (no HTTP).

## Stack

- **React 19** + **TypeScript 5**, **Vite 8**
- **Plain CSS3** design system in `src/index.css` — OKLCH tokens + semantic component classes (no Tailwind/PostCSS)
- **Base UI** (`@base-ui/react`) — unstyled, accessible primitives wrapped in `src/components/`
- **React Router 7** for client routing (hash)
- **TanStack Query 5** for server-state caching
- **Zustand 5** for local UI / session state (with `persist` middleware)
- **React Hook Form 7** + **Zod 4** for forms / validation
- **lucide-react** for icons, **recharts** for charts
- **@fontsource/geist-sans** / **@fontsource/geist-mono** for self-hosted Geist fonts

## Folder Structure

```
src/
  app/               # App.tsx (route table), Providers.tsx, ErrorBoundary.tsx
  pages/             # route screens (1 per module + SetupWizard + Welcome)
  features/          # feature-based modules (dashboard/, customers/, sales/, ...)
  components/        # category folders, not by feature
    button/          # Button (Base UI) — variants + sizes + loading
    input/           # Input, Textarea, Label, SearchInput, PasswordInput
    select/          # Select (Base UI)
    form/            # Form (RHF + zod) + field components
    table/           # DataTable, TablePagination
    dialog/          # Dialog + Body/Header/Footer, AlertDialog, ConfirmDialog, CancelDialog
    card/            # Card, StatCard
    badge/           # Badge + status-specific variants
    navigation/      # Sidebar, Topbar, Breadcrumbs, nav config (nav.ts)
    layout/          # AppLayout, PageContainer, PageHeader, Section, Grid
    misc/            # DropdownMenu, Tooltip, Drawer, RowActions
    feedback/        # Spinner, EmptyState, ErrorState, Toaster
    charts/          # LineChart, BarChart (recharts wrappers, token colors)
    tabs/            # Tabs (Base UI)
  services/          # one folder per business domain + queryKeys.ts, bindings.ts, wails-types.ts
  stores/            # Zustand: theme, sidebar, notification
  hooks/             # useDebounce
  constants/         # routes (English slugs), currencies, countries, payment methods
  utils/             # cx, format, storage
  locales/           # es-PE translation dictionary + t() helper
  types/             # shared domain types (Customer, Product, Supplier, Sale, ...)
  main.tsx           # Vite entrypoint
```

Every `components/<category>/` has an `index.ts` barrel — **import from `@/components/<category>`**, not from individual files.

## Commands

```bash
pnpm install         # install dependencies
pnpm dev             # Vite dev server (frontend only)
pnpm build           # production build → dist/ (Wails embeds this)
pnpm check           # tsc --noEmit (type check)
```

Wails-specific:

```bash
wails dev           # run Vite + Go together, hot-reload
wails build         # produce desktop binary in build/bin/
```

## Conventions

- The frontend **must not** access the database directly. All calls go through Wails bindings exposed by the Go `App` and `bindings.App` structs.
- All UI text is in **Spanish (es-PE)**. Route slugs are English (see `@/constants/routes`).
- All numbers / dates / currency use `Intl.*` helpers in `@/utils/format`. **Never** use `toFixed` for money or `toLocaleString` ad-hoc.
- Path alias `@/*` resolves to `src/*`.
- All style tokens (CSS variables) and semantic component classes live in `src/index.css`. Use the component classes (`.btn`, `.card`, `.input`, …) — never hardcode colors. This is the only stylesheet.
- Use `cx()` from `@/utils/cx` for conditional class composition. Don't write raw string concatenation.
- Destructive actions go through `<AlertDialog variant="destructive">` or `<ConfirmDialog>`.
- Forms use `react-hook-form` + `zod`.
