# Sistema de Diseño — vfinancy ERP

> Documento canónico. Todo componente, pantalla y módulo debe cumplirlo.
> Cualquier desviación requiere justificación escrita y aprobación.
>
> Los tokens tienen **dos representaciones sincronizadas**:
> 1. **CSS variables** en `src/index.css` (lo que consume Tailwind y la UI en runtime).
> 2. **TypeScript** en `src/design-system/` (lo que consume el código TS para type safety y valores no-CSS como z-index, breakpoints, durations).
>
> Al cambiar un valor, actualizar **ambos**. Las claves son idénticas.

Inspirado en: **SAP Fiori**, **Microsoft Dynamics 365**, **Odoo Enterprise**, **Linear**, **Vercel Dashboard**.

---

## 1. Principios

1. **Claridad ante decoración.** Cada elemento visible tiene una función.
2. **Densidad equilibrada.** Información suficiente sin saturar.
3. **Consistencia radical.** Misma acción = mismo color, misma forma, misma posición.
4. **Acción reversible.** Acciones destructivas requieren confirmación.
5. **Accesible por defecto.** WCAG 2.1 AA como mínimo.
6. **Respuesta inmediata.** Toda interacción muestra resultado en < 100 ms.
7. **Sin estado de carga invisible.** Toda espera > 200 ms muestra `Spinner` o `Skeleton`.

---

## 2. Paleta de Colores

### 2.1 Modo Claro

| Token              | Valor HSL  | Uso                                       |
|--------------------|------------|-------------------------------------------|
| `--background`     | `0 0% 100%`    | Fondo de página                           |
| `--foreground`     | `222 47% 11%`  | Texto principal                           |
| `--muted`          | `210 40% 96%`  | Fondo de zonas secundarias                |
| `--muted-foreground` | `215 16% 47%` | Texto secundario / placeholders         |
| `--card`           | `0 0% 100%`    | Fondo de tarjetas / dialogs               |
| `--card-foreground`| `222 47% 11%`  | Texto sobre tarjetas                      |
| `--popover`        | `0 0% 100%`    | Fondo de popovers / dropdowns             |
| `--border`         | `214 32% 91%`  | Bordes 1 px                               |
| `--input`          | `214 32% 91%`  | Borde de inputs                           |
| `--ring`           | `222 47% 11%`  | Anillo de focus                           |
| `--primary`        | `222 47% 11%`  | Marca / acción principal                  |
| `--primary-foreground` | `210 40% 98%` | Texto sobre `--primary`                 |

### 2.2 Modo Oscuro

| Token              | Valor HSL  |
|--------------------|------------|
| `--background`     | `222 47% 6%`   |
| `--foreground`     | `210 40% 98%`  |
| `--muted`          | `217 33% 11%`  |
| `--muted-foreground` | `215 20% 65%` |
| `--card`           | `222 47% 8%`   |
| `--card-foreground`| `210 40% 98%`  |
| `--popover`        | `222 47% 8%`   |
| `--border`         | `217 33% 17%`  |
| `--input`          | `217 33% 17%`  |
| `--ring`           | `213 27% 84%`  |
| `--primary`        | `210 40% 98%`  |
| `--primary-foreground` | `222 47% 11%` |

### 2.3 Semánticos

| Token                       | Claro (HSL)          | Oscuro (HSL)         | Significado              |
|-----------------------------|----------------------|----------------------|--------------------------|
| `--success`                 | `142 71% 45%`        | `142 71% 45%`        | Confirmación / pagado    |
| `--success-foreground`      | `0 0% 100%`          | `0 0% 100%`          |                          |
| `--warning`                 | `38  92% 50%`        | `38  92% 50%`        | Pendiente / alerta       |
| `--warning-foreground`      | `0 0% 100%`          | `0 0% 100%`          |                          |
| `--destructive`             | `0 84% 60%`          | `0 63% 31%`          | Error / eliminar         |
| `--destructive-foreground`  | `210 40% 98%`        | `210 40% 98%`        |                          |
| `--info`                    | `199 89% 48%`        | `199 89% 48%`        | Información              |
| `--info-foreground`         | `0 0% 100%`          | `0 0% 100%`          |                          |

### 2.4 Estados de Inventario / Negocio

| Token                    | Color       | Uso                                  |
|--------------------------|-------------|--------------------------------------|
| Badge `Pendiente`        | warning     | Ventas/compras pendientes            |
| Badge `Pagado`           | success     | Ventas/compras pagadas               |
| Badge `Parcial`          | info        | Pago parcial                         |
| Badge `Cancelado`        | destructive | Anulado                              |
| Badge `Activo`           | success     | Clientes/proveedores activos         |
| Badge `Inactivo`         | muted       | Registros deshabilitados             |
| Badge `Stock Bajo`       | warning     | Inventario bajo mínimo               |
| Badge `Sin Stock`        | destructive | Inventario en cero                   |
| Badge `Remate`           | destructive | Inventario > 25 días (limpieza)      |

### 2.5 Reglas de Uso

- **Un solo color de marca (`--primary`)** por pantalla.
- **Acciones destructivas siempre usan `--destructive`.** Nunca rojo libre.
- **Estados neutros** (`muted`, `muted-foreground`) para texto auxiliar — nunca un gris inventado.
- **Contraste mínimo 4.5:1** entre texto y fondo. Verificar en ambos modos.

---

## 3. Tipografía

### 3.1 Familia

**Inter** (auto-hospedada en `src/assets/fonts/`, peso 400, 500, 600, 700). Fallback al stack del sistema.

```css
font-family:
  'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
```

### 3.2 Escala (Tailwind)

| Uso            | Clase           | Tamaño  | Line-height | Peso  |
|----------------|-----------------|---------|-------------|-------|
| Display        | `text-3xl`      | 30 px   | 36 px       | 600   |
| Título página  | `text-2xl`      | 24 px   | 32 px       | 600   |
| Subtítulo      | `text-xl`       | 20 px   | 28 px       | 600   |
| Sección        | `text-lg`       | 18 px   | 28 px       | 600   |
| Cuerpo         | `text-base`     | 14 px   | 20 px       | 400   |
| Tabla / denso  | `text-sm`       | 13 px   | 20 px       | 400   |
| Etiqueta       | `text-xs`       | 12 px   | 16 px       | 500   |
| Numérico       | `tabular-nums`  | —       | —           | —     |

- **Siempre `tabular-nums`** en columnas de tablas financieras.
- **Tracking**: `-tracking-tight` en `text-2xl` y `text-3xl`.
- **Saltos permitidos**: 1 nivel entre jerarquía vecina (no `text-3xl` → `text-sm` directo).

---

## 4. Espaciado

Escala Tailwind estándar (múltiplos de 4 px). Reglas de uso:

| Token       | Valor     | Uso                                  |
|-------------|-----------|--------------------------------------|
| `space-1`   | 4 px      | Entre icono y texto en línea         |
| `space-2`   | 8 px      | Entre filas de control               |
| `space-3`   | 12 px     | Padding de inputs/buttons            |
| `space-4`   | 16 px     | Padding de cards / separación mayor  |
| `space-6`   | 24 px     | Separación entre secciones de página  |
| `space-8`   | 32 px     | Separación entre bloques principales |
| `space-12`  | 48 px     | Hero / dashboard header              |

- **Cards**: `p-6` (24 px).
- **Tablas**: celdas `px-4 py-3`.
- **Dialogs**: padding interno `p-6`, header `p-6`, footer `p-4`.
- **Sidebar item**: `px-3 py-2` (12 × 8).
- **Topbar**: altura `h-14` (56 px), padding horizontal `px-4`.
- **Sidebar expandido**: ancho `w-64` (256 px).
- **Sidebar colapsado**: ancho `w-16` (64 px).

---

## 5. Radios (Bordes Redondeados)

| Token  | Valor | Uso                                     |
|--------|-------|-----------------------------------------|
| `sm`   | 2 px  | Badges / chips                          |
| `md`   | 6 px  | Inputs, botones (default)               |
| `lg`   | 8 px  | Cards, dialogs, popovers                |
| `xl`   | 12 px | Modales de confirmación                 |
| full   | 9999px| Avatares, switches                      |

Variable CSS: `--radius: 0.5rem` (8 px, base para `lg`).

---

## 6. Sombras

```css
/* Sutil — usado para cards en reposo */
--shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);

/* Card elevada / dropdown */
--shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.08), 0 2px 4px -2px rgb(0 0 0 / 0.05);

/* Dialog / popover / topbar con scroll */
--shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.08), 0 4px 6px -4px rgb(0 0 0 / 0.05);

/* Notificaciones flotantes / dialogs críticos */
--shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.10), 0 8px 10px -6px rgb(0 0 0 / 0.05);
```

Reglas:
- Cards en reposo: `--shadow-sm` o sin sombra (`border` 1 px basta).
- Dropdown / popover: `--shadow-md`.
- Dialog: `--shadow-lg`.
- Toast: `--shadow-xl`.

---

## 7. Iconos

- Librería: **lucide-react** exclusivamente.
- Tamaño: 16 px (dentro de botones), 20 px (sidebar/topbar), 24 px (destacados).
- Stroke: 1.75 (más fino que default 2 — look enterprise moderno).
- Color: hereda `currentColor`. No colorear íconos manualmente.
- Cada elemento de menú del sidebar debe tener un ícono (ver lista en §14).

---

## 8. Botones

### 8.1 Variantes

| Variante      | Uso                                       | Estilo                                       |
|---------------|-------------------------------------------|----------------------------------------------|
| `default`     | Acción primaria de la pantalla            | `bg-primary text-primary-foreground`         |
| `secondary`   | Acción secundaria                         | `bg-secondary text-secondary-foreground`     |
| `outline`     | Acción terciaria                          | `border bg-background`                       |
| `ghost`       | Acción sutil (en tablas, toolbar)         | sin fondo, hover `bg-accent`                 |
| `link`        | Navegación inline                         | `text-primary underline`                     |
| `destructive` | Eliminar / cancelar definitivamente       | `bg-destructive text-destructive-foreground` |
| `success`     | Confirmar positiva (pagar, cobrar)        | `bg-success text-success-foreground`         |

### 8.2 Tamaños

| Tamaño  | Altura | Padding        | Texto    | Uso                          |
|---------|--------|----------------|----------|------------------------------|
| `sm`    | 32 px  | `px-3`         | `text-sm`| Tablas, espacios densos      |
| `md`    | 40 px  | `px-4`         | `text-sm`| Default                      |
| `lg`    | 44 px  | `px-6`         | `text-base` | CTAs principales           |
| `icon`  | 40 px  | `h-10 w-10`    | —        | Toolbar / cards              |

### 8.3 Reglas

- **Un solo botón `default` por pantalla** (la acción principal). El resto `outline` / `ghost`.
- **Acciones destructivas siempre con `variant="destructive"`** y dentro de un `AlertDialog` de confirmación.
- Botón de carga: `disabled` + `<Spinner />` interior + texto permanece.
- Icono + texto: `gap-2`. Icono nunca reemplaza al texto en acciones críticas (excepto toolbar).

---

## 9. Inputs

### 9.1 Tipos Soportados

`text`, `number`, `email`, `password`, `search`, `tel`, `url`, `date`, `datetime-local`, `time`, `textarea`, `select` (single), `combobox`, `checkbox`, `radio`, `switch`, `money`, `file`.

### 9.2 Estructura (Field)

```tsx
<Field>
  <FieldLabel required>Nombre</FieldLabel>
  <Input ... />
  <FieldDescription>Texto de ayuda opcional</FieldDescription>
  <FieldError>{errors.name?.message}</FieldError>
</Field>
```

### 9.3 Reglas

- Altura input: `h-10` (40 px).
- Borde: `border-input`; en focus `ring-2 ring-ring ring-offset-2`.
- Label **siempre visible** (no placeholder-only).
- Validación inline: `FieldError` con `text-destructive` debajo del input.
- Inputs monetarios: alineados a la derecha, prefijo de moneda visible, `tabular-nums`.
- Inputs de fecha: usar calendario nativo (HTML `date`) estilado.
- **Nunca** deshabilitar el copy/paste en inputs.

---

## 10. Tablas

### 10.1 Estructura

```tsx
<DataTable
  columns={columns}      // definición de columnas (id, header, cell, sortable, align)
  data={items}           // filas
  loading={isLoading}    // muestra <Skeleton> en filas
  empty={<EmptyState />} // componente cuando data=[]
  onRowClick={...}       // opcional
  selection             // checkbox por fila
  pagination={{...}}     // ver §10.4
/>
```

### 10.2 Anatomía

- **Header**: `text-xs uppercase tracking-wider text-muted-foreground`, fondo `bg-muted/50`.
- **Filas**: `hover:bg-muted/30`, separador `border-b border-border`.
- **Celdas**: `px-4 py-3`, `text-sm`.
- **Numéricos**: alineados a la derecha, `tabular-nums`.
- **Acciones de fila**: íconos `ghost` alineados a la derecha, en columna fija de 80 px.

### 10.3 Funcionalidades obligatorias

- [x] Ordenamiento por columna (click en header).
- [x] Paginación (25, 50, 100 por página).
- [x] Filtros por columna (popover) — opcional, decisión por módulo.
- [x] Visibilidad de columnas (menú desde header).
- [x] Selección múltiple con checkbox de cabecera.
- [x] Acciones en bloque (toolbar contextual superior cuando hay selección).
- [x] Estado de carga: skeleton con misma estructura.
- [x] Estado vacío: ilustración + mensaje + acción primaria.

### 10.4 Paginación

Controles: `Anterior` `[1] 2 3 ... 10` `Siguiente`. Selector de tamaño de página a la izquierda. Rango "Mostrando 1–25 de 247" a la derecha.

---

## 11. Dialogs / Modales

### 11.1 Variantes Semánticas

| Variante        | Uso                                          | Botón principal          |
|-----------------|----------------------------------------------|--------------------------|
| `confirmation`  | Confirmar acción reversible                 | `default`                |
| `delete`        | Eliminar registro                            | `destructive` "Eliminar" |
| `success`       | Resultado exitoso (pago, guardado)            | `default` "Aceptar"      |
| `warning`       | Alerta antes de acción riesgosa              | `default`                |
| `error`         | Error de operación                           | `default` "Cerrar"       |
| `info`          | Información / ayuda                          | `default` "Entendido"    |

### 11.2 Anatomía

```
┌─────────────────────────────────────────┐
│ [Icono semántico]  Título               │  ← header
│ Descripción opcional                    │
├─────────────────────────────────────────┤
│                                         │
│ Contenido (form, texto, tabla, etc.)    │
│                                         │
├─────────────────────────────────────────┤
│              [Cancelar] [Confirmar]     │  ← footer
└─────────────────────────────────────────┘
```

- Ancho por tipo: pequeño 400 px, mediano 560 px, grande 800 px, full 90 vw.
- Cierre: click en backdrop, tecla `Esc`, botón `X`.
- Confirmación con Enter, cancelación con Esc.
- Trap de foco dentro del dialog.

---

## 12. Notificaciones (Toasts)

### 12.1 Tipos

`success`, `warning`, `destructive`, `info`. Mismo sistema visual que `AlertDialog`.

### 12.2 Reglas

- Posición: bottom-right (es-PE) — área no obstructiva.
- Duración: 4 s (success/info), 6 s (warning), persistente (destructive) con botón cerrar.
- Apilamiento vertical: 8 px entre toasts, máx 3 visibles, FIFO.
- Título + descripción opcional. Botón de acción opcional.
- Cerrar con `X` o click en el toast.

---

## 13. Estados de Carga

| Componente      | Cuándo usar                                       |
|-----------------|---------------------------------------------------|
| `Spinner`       | Acción puntual (botón, sección)                   |
| `Skeleton`      | Carga de página / lista / tabla                   |
| `ProgressBar`   | Operación con progreso conocido (import, backup)  |
| `EmptyState`    | Carga completada sin resultados                   |
| `ErrorState`    | Carga completada con error                        |

---

## 14. Navegación (Sidebar)

Estructura de módulos y sus íconos (lucide-react):

| Ruta              | Etiqueta         | Ícono                          |
|-------------------|------------------|--------------------------------|
| `/`               | Inicio           | `LayoutDashboard`              |
| `/clientes`       | Clientes         | `Users`                        |
| `/proveedores`    | Proveedores      | `Truck`                        |
| `/productos`      | Productos        | `Package`                      |
| `/inventario`     | Inventario       | `Warehouse`                    |
| `/compras`        | Compras          | `ShoppingCart`                 |
| `/ventas`         | Ventas           | `Receipt`                      |
| `/tesoreria`      | Tesorería        | `Landmark`                     |
| `/contabilidad`   | Contabilidad     | `BookOpen`                     |
| `/reportes`       | Reportes         | `BarChart3`                    |
| `/configuracion`  | Configuración    | `Settings`                     |
| `/administracion` | Administración   | `Shield`                       |

---

## 15. Tema (Light / Dark)

### 15.1 Implementación

- `class` strategy en Tailwind (`darkMode: ['class']`).
- Clase `dark` en `<html>` activada por `themeStore`.
- Persistencia: `localStorage` key `vfinancy.theme`. Default: `system` (respeta `prefers-color-scheme`).
- Cambio instantáneo sin parpadeo (aplicar clase antes de hidratar React).

### 15.2 Estados

`light` | `dark` | `system`. Solo `light` y `dark` se exponen al usuario; `system` se resuelve internamente.

---

## 16. Layout

### 16.1 Estructura de Pantalla

```
┌────────────────────────────────────────────────────────┐
│ Topbar (h-14, sticky)                                  │ 56 px
├──────┬─────────────────────────────────────────────────┤
│      │ Breadcrumbs                                      │
│ Side │                                                 │
│ bar  │  Título de página                                │
│      │  Subtítulo / acciones                           │
│ w-64 │ ─────────────────────────────────────────────── │
│ or   │                                                 │
│ w-16 │  Contenido (p-6, max-w-7xl mx-auto)            │
│      │                                                 │
│      │                                                 │
└──────┴─────────────────────────────────────────────────┘
```

- **Sidebar**: fijo, scroll independiente.
- **Main**: scroll vertical propio. Header de página sticky dentro del main.
- **Container**: `max-w-7xl` por defecto; tablas a `max-w-full`.

### 16.2 Breakpoints (Tailwind)

| Breakpoint | Mínimo | Soporte                                |
|------------|--------|----------------------------------------|
| `sm`       | 640 px | Móvil básico                           |
| `md`       | 768 px | Tablet vertical                        |
| `lg`       | 1024 px| Tablet horizontal / desktop compacto   |
| `xl`       | 1280 px| Desktop estándar                       |
| `2xl`      | 1536 px| Desktop amplio (1920 px)               |

App optimizada para **1280 × 800 mínimo**. Funciona desde 1024 px con sidebar colapsado.

---

## 17. Accesibilidad (WCAG 2.1 AA)

- [x] Contraste mínimo 4.5:1 en texto normal, 3:1 en texto grande.
- [x] Foco visible en todo elemento interactivo (`ring-2 ring-ring`).
- [x] Navegación completa por teclado (Tab / Shift+Tab / Enter / Esc / flechas).
- [x] Roles ARIA correctos en dialogs, menús, listas, tablas.
- [x] `aria-label` en íconos sin texto.
- [x] Estados de carga anunciados con `aria-live="polite"`.
- [x] Errores de formulario asociados con `aria-describedby`.
- [x] Respeto a `prefers-reduced-motion` (animaciones de carga).

---

## 18. Animación

- Duración: 150 ms (transiciones hover), 200 ms (entradas/salidas de dialog), 300 ms (cambios de página).
- Easing: `cubic-bezier(0.4, 0, 0.2, 1)` (Tailwind default).
- `prefers-reduced-motion: reduce` → deshabilitar todas las transiciones.

---

## 19. Iconografía Numérica (KPIs)

Cada tarjeta de estadística usa:
- Título (etiqueta) arriba, `text-sm text-muted-foreground`.
- Valor grande, `text-3xl font-semibold tabular-nums`.
- Indicador de tendencia opcional: `TrendingUp` verde / `TrendingDown` rojo / `Minus` gris.
- Subtítulo: comparación vs período anterior, `text-xs`.

---

## 20. Reglas Inquebrantables

1. **Cero colores hardcodeados.** Solo tokens semánticos.
2. **Cero `style={{...}}` para colores / espaciado / tipografía.** Solo clases Tailwind o variables CSS.
3. **Cero `px` sueltos en CSS.** Usar escala de Tailwind.
4. **Toda acción destructiva pasa por `AlertDialog`.**
5. **Todo formulario usa `react-hook-form` + `zod`.** Sin validaciones ad-hoc.
6. **Todo dato monetario formatea con `formatCurrency(value, 'PEN')`.** Sin `toFixed`.
7. **Toda fecha muestra con `formatDate(value)` (es-PE).** Sin `toLocaleString` ad-hoc.
8. **Sin emojis en la UI** salvo que el usuario lo pida explícitamente.
9. **Textos en español (es-PE).** Comentarios y nombres de variables en inglés.
10. **Componentes reutilizables primero.** Si un patrón se repite 2 veces, se extrae a `components/`.

---

## 21. Inventario (Regla de Negocio Visible)

El ERP aplica la regla de remate:

```
Fecha Máxima de Venta = Fecha de Ingreso + 25 días
```

Cuando hoy > Fecha Máxima, el producto aparece con badge `Remate` y entra al dashboard. Esta regla es **visible en la UI** mediante un campo calculado en cada item de inventario y un filtro global "Mostrar solo remate".

---

## 22. Estructura de Componentes

```
src/components/
  button/         # Button, IconButton, ButtonGroup
  input/          # Input, Textarea, Field, FieldLabel, FieldError, FieldDescription
  select/         # Select, Combobox, MultiSelect
  checkbox/       # Checkbox, RadioGroup, Switch
  table/          # DataTable, TablePagination, ColumnVisibilityMenu, TableSkeleton
  dialog/         # Dialog, AlertDialog (success/warning/destructive/info), Sheet
  card/           # Card, CardHeader, CardTitle, CardContent, StatCard, SummaryCard, FinancialCard
  badge/          # Badge, StatusBadge
  navigation/     # Sidebar, Topbar, Breadcrumbs, NavLink, Tabs
  feedback/       # Toast, Spinner, Skeleton, ProgressBar, EmptyState, ErrorState
  charts/         # LineChart, BarChart, PieChart (wrappers de recharts)
  layout/         # PageHeader, PageContainer, Section, Grid, Stack
  money/          # MoneyInput, MoneyDisplay
  date/           # DatePicker (wrapper del nativo), DateRangePicker
  misc/           # Kbd, Separator, Tooltip, DropdownMenu
```

---

## 23. Catálogo de Módulos (Placeholders de Fase 0.5)

| Módulo              | Ruta              | Estado          |
|---------------------|-------------------|-----------------|
| Inicio              | `/`               | Implementado    |
| Clientes            | `/clientes`       | Placeholder     |
| Proveedores         | `/proveedores`    | Placeholder     |
| Productos           | `/productos`      | Placeholder     |
| Inventario          | `/inventario`     | Placeholder     |
| Compras             | `/compras`        | Placeholder     |
| Ventas              | `/ventas`         | Placeholder     |
| Tesorería           | `/tesoreria`      | Placeholder     |
| Contabilidad        | `/contabilidad`   | Placeholder     |
| Reportes            | `/reportes`       | Placeholder     |
| Configuración       | `/configuracion`  | Placeholder     |
| Administración      | `/administracion` | Placeholder     |

Cada placeholder muestra: título, descripción del módulo futuro, y 2–3 cards con KPIs de ejemplo. **No conecta a base de datos.**

---

_Fin del documento. Última actualización: Phase 0.6 — agregado registro de tokens TypeScript en `src/design-system/`._
