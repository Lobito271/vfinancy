# vfinancy design system

## Direction

vfinancy is an operations-focused ERP for purchasing, sales, inventory, treasury, and reporting. Its interface is flat, functional, quiet, and economical with visual framing. Use whitespace, alignment, typography, and fine separators to make dense business information easy to scan. Surfaces exist only where they clarify a distinct task or interaction.

Keep the existing color palette exactly as defined in `frontend/src/index.css`, including light/dark values and semantic colors. Do not add colors, gradients, or new aliases. `sample.html` is a static historical reference and must not be modified.

## Foundations

### Color and themes

CSS variables in `frontend/src/index.css` are the sole source of palette values. Preserve all current hex values and light/dark theme behavior. Use semantic tokens for feedback and keep status text legible without relying on color alone. Avoid tinted panels as decoration.

### Typography

- **Figtree** is used for interface text, controls, labels, and tabular data.
- **Montserrat** is used sparingly for page titles and key figures.
- Use sentence case for page and section titles. Reserve uppercase labels for compact table headings where it improves scanning.
- Keep hierarchy restrained: clear title, readable body text, muted supporting details, and tabular numerals for quantities and money.

### Geometry and spacing

- Keep the existing square geometry; do not use rounded corners.
- Use the existing 4px/8px spacing rhythm. Provide generous page margins and consistent gaps between distinct content groups.
- Prefer 1px separators and subtle borders over heavy outlines. Do not use decorative shadows or color blocks to simulate depth.
- Align headings, controls, and content edges. Keep table density practical and maintain readable touch targets.

## Layout

The app shell retains its collapsible navigation, topbar, and internally scrolling workspace. Preserve its existing keyboard, theme, and responsive behavior.

Use the layout archetype that fits each task:

| Archetype | Structure |
|---|---|
| Dashboard | Page header followed by a responsive grid of metrics, charts, and operational lists |
| Dense list | Page header and actions, optional KPI row, then table and its toolbar |
| Multiple lists | Page header followed by separated, titled list blocks |
| Settings | Page header followed by clearly titled settings groups and compact forms |
| Standalone flow | Focused setup, welcome, and recovery screens without app navigation |

Keep the canvas visually open. Do not wrap every group in a card. Separate sibling groups with whitespace or a hairline rule, and reserve bounded surfaces for interactive overlays, standalone forms, and data tables where the boundary improves usability.

## Components

### Surfaces and sections

- Use at most one visually bounded surface per content zone.
- Never nest cards, sections, or framed panels of equivalent weight.
- A list section uses a frameless heading directly above its table. The table owns its own boundary.
- Dashboard widgets use simple separators or a light boundary only when needed to distinguish independent content; KPI values do not need individual decorative framing.
- Dialogs and drawers may use a clear surface boundary, with internal groups separated by spacing or hairlines.

### Navigation

- Keep primary navigation legible in expanded and collapsed modes, with a clear active state and visible focus.
- Use the existing yellow accent selectively for active navigation and primary actions.
- Maintain the mobile navigation drawer and its existing interaction behavior.

### Controls and forms

- Keep controls easy to identify with consistent heights, clear labels, and subtle full borders.
- Use yellow for primary actions and active states; secondary controls remain quiet until interacted with.
- Keep validation text adjacent to the field and preserve visible keyboard focus.
- Group related fields with spacing and alignment, not nested boxes.

### Tables and data

- Use the shared `DataTable` component for tabular data.
- Use subtle row separators, a legible header, and restrained hover feedback.
- Keep numeric values aligned consistently and use tabular numerals for financial and quantity data.
- Preserve sticky headers/columns, sorting, filtering, pagination, row actions, and keyboard-operable rows.

### Feedback and status

- Pair every status color with a text label.
- Keep empty, loading, and error states concise and aligned with the content they describe.
- Preserve existing focus rings, validation announcements, dialog focus management, and keyboard support.

## Responsive behavior

Preserve the existing breakpoints and behavior around 1200px, 800px, and 600px. Grids should collapse cleanly, toolbars may wrap, forms should become single-column, and tables should retain a usable horizontal scroll area. Avoid fixed widths that cause overflow on narrow screens.

## Implementation rules

- Plain CSS in `frontend/src/index.css`; no utility CSS framework or new styling dependency.
- Use semantic component classes and existing CSS tokens. Do not add one-off utility classes or hardcoded colors.
- Keep the existing palette values exactly unchanged.
- Preserve content, functionality, accessibility, and interactions while adjusting visual presentation.
- `sample.html` is not part of implementation changes.
