# Guía de Componentes

> Inventario completo de los componentes del design system, su API y cuándo usarlos.
> Para reglas de diseño, ver `DESIGN_SYSTEM.md`. Para arquitectura, `ARCHITECTURE.md`.

**Convención:** todos los componentes se importan desde `@/components/<categoría>`, nunca desde el archivo individual.

```ts
// ✓
import { Button } from '@/components/button';
import { Card, CardHeader, CardTitle } from '@/components/card';

// ✗
import { Button } from '@/components/button/Button';
```

---

## button

### `<Button>`

```tsx
<Button variant="default" size="md" loading={isSubmitting}>
  Guardar
</Button>
```

| Prop      | Tipo                                                                           | Default      | Notas                                  |
|-----------|--------------------------------------------------------------------------------|--------------|----------------------------------------|
| `variant` | `default` \| `secondary` \| `outline` \| `ghost` \| `link` \| `destructive` \| `success` | `default` | Solo **un** `default` por pantalla     |
| `size`    | `sm` \| `md` \| `lg` \| `icon` \| `icon-sm`                                     | `md`         | `icon` para toolbar, `lg` para CTAs    |
| `loading` | `boolean`                                                                     | `false`      | Reemplaza contenido con `<Spinner />`  |
| `asChild` | `boolean`                                                                     | `false`      | Pasa estilos a un hijo (Radix Slot)    |
| `disabled`| `boolean`                                                                     | `false`      | Auto-disabled cuando `loading`          |

Reglas: las acciones destructivas siempre usan `variant="destructive"` Y van dentro de un `AlertDialog`.

---

## input

### `<Input>`

```tsx
<Input type="text" placeholder="..." invalid={hasError} />
```

Props estándar de `<input>` + `invalid?: boolean` para estado de error.

### `<Textarea>`

```tsx
<Textarea rows={3} maxLength={500} />
```

### `<SearchInput>`

```tsx
<SearchInput value={q} onChange={(e) => setQ(e.target.value)} onClear={() => setQ('')} />
```

Input con icono de búsqueda y botón limpiar (cuando hay valor). Usa `type="search"`.

### `<Field>`, `<Label>`, `<FieldDescription>`, `<FieldError>`

Solo se usan **directamente** cuando el field personalizado no se puede componer con los de `components/form/`. En formularios, prefiere `TextField`, `EmailField`, etc.

---

## form

Campos listos para usar con `react-hook-form`. Requieren un `<Form>` o `FormProvider` como padre.

| Componente                  | Tipo / Validación                          | Notas                          |
|-----------------------------|--------------------------------------------|--------------------------------|
| `<TextField>`               | text, email, tel, url, password            | `register` directo             |
| `<NumberField>`             | number                                     | `valueAsNumber: true`          |
| `<MoneyField>`              | número con formato de moneda               | Símbolo de moneda a la izq.    |
| `<PercentageField>`         | 0–100                                      | Sufijo %                       |
| `<CurrencyField>`           | selector PEN/USD/EUR/...                   | Lista de `Currencies`          |
| `<DateField>`               | date (nativo HTML)                         | `min` / `max` soportados       |
| `<DateRangeField>`          | dos `DateField`                            | `fromName` / `toName`          |
| `<DateTimeField>`           | datetime-local                             |                                |
| `<EmailField>`              | email                                      |                                |
| `<PhoneField>`              | tel, valida formato PE                     | `+51` prefijo, 9 dígitos       |
| `<PasswordField>`           | password, con botón mostrar/ocultar        |                                |
| `<SearchField>`             | search                                     | Auto-clear                      |
| `<TextareaField>`           | multilínea                                 | `rows` configurable             |
| `<CheckboxField>`           | boolean                                    | Label a la derecha              |
| `<SelectField>`             | opción de lista                            | `options: SelectOption[]`       |
| `<AsyncSelectField>`        | opciones cargadas async                    |                                |
| `<CustomerSelectField>`     | wrapper sobre AsyncSelect + service        |                                |
| `<SupplierSelectField>`     |                                            |                                |
| `<ProductSelectField>`      |                                            |                                |
| `<WarehouseSelectField>`    |                                            |                                |
| `<CategorySelectField>`     |                                            |                                |
| `<BrandSelectField>`        |                                            |                                |
| `<TaxSelectField>`          |                                            |                                |
| `<CurrencySelectField>`     |                                            |                                |
| `<DocumentTypeSelectField>` |                                            |                                |

Todos aceptan:
- `name: FieldPath<T>`
- `label?: string`
- `description?: string`
- `required?: boolean`
- `disabled?: boolean`
- `className?: string`
- Y el resto de props del input subyacente.

`<MoneyField>` adicional: `currency: CurrencyCode` (default `PEN`).

### `<Form>`

```tsx
<Form<MyValues>
  defaultValues={...}
  onSubmit={async (values) => { ... }}
>
  {() => <TextField name="..." label="..." />}
</Form>
```

---

## select

Wrapper sobre `@radix-ui/react-select`. Raramente se usa directamente — preferir `<SelectField>` (form).

```tsx
<Select value={v} onValueChange={setV}>
  <SelectTrigger><SelectValue placeholder="..." /></SelectTrigger>
  <SelectContent>
    <SelectItem value="a">Opción A</SelectItem>
    <SelectItem value="b">Opción B</SelectItem>
  </SelectContent>
</Select>
```

---

## checkbox

`<Checkbox>`, `<Switch>`, `<RadioGroup>` + `<RadioGroupItem>`. Wrappers sobre Radix.

---

## table

### `<DataTable>` — el framework

Ver `ARCHITECTURE.md` § "Tablas (DataTable Framework)" para capacidades completas.

```tsx
<DataTable
  columns={columns}
  data={rows}
  keyField="id"
  loading={isLoading}
  error={error}
  onRetry={refetch}
  onSelectionChange={(rows) => setSelected(rows)}
  bulkActions={(rows) => <BulkDeleteButton ids={rows.map(r => r.id)} />}
  toolbarLeft={<FilterSelect />}
  toolbarRight={<ExportButton />}
  preferencesKey="customers"          // persiste visibilidad + pageSize
  exportFilename="clientes.csv"
  globalSearch
  exportable
  stickyFirstColumn
  ariaLabel="Listado de clientes"
/>
```

### `<TablePagination>`

Solo se usa si necesitas paginación custom (DataTable ya la incluye).

---

## dialog

### `<Dialog>`

```tsx
<Dialog open={open} onOpenChange={setOpen}>
  <DialogContent size="md">
    <DialogHeader>
      <DialogTitle>Título</DialogTitle>
      <DialogDescription>Descripción</DialogDescription>
    </DialogHeader>
    {/* contenido */}
    <DialogFooter>
      <Button variant="outline" onClick={() => setOpen(false)}>Cancelar</Button>
      <Button onClick={handleConfirm}>Confirmar</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

`size`: `sm` (24rem) | `md` (28rem) | `lg` (42rem) | `xl` (64rem).

### `<AlertDialog>` (5 variantes)

```tsx
<AlertDialog
  open={open}
  onOpenChange={setOpen}
  variant="destructive"     // success | warning | destructive | info | confirmation
  title="¿Eliminar cliente?"
  description="Esta acción no se puede deshacer."
  confirmLabel="Eliminar"
  cancelLabel="Cancelar"
  onConfirm={handleDelete}
  loading={isDeleting}
/>
```

### `<ConfirmDialog>`

Atajo sobre `<AlertDialog variant="confirmation">` para confirmaciones no destructivas.

---

## card

`<Card>`, `<CardHeader>`, `<CardTitle>`, `<CardDescription>`, `<CardContent>`, `<CardFooter>` — primitivas de tarjeta.

### `<StatCard>`

Para KPIs del dashboard:

```tsx
<StatCard
  label="Ventas del Mes"
  value={formatCurrency(125000)}
  icon={ShoppingCart}
  change={12.4}             // % vs periodo anterior (opcional)
  changeLabel="vs. mes anterior"
/>
```

---

## badge

`<Badge variant="default | secondary | outline | success | warning | destructive | info | muted">`.

Wrappers de estado:
- `<SaleStatusBadge status="paid" />` — paid/pending/partial/cancelled
- `<CustomerStatusBadge status="active" />` — active/inactive/blocked
- `<StockBadge stock={x} minStock={y} isClearance={z} />` — incluye remate

---

## navigation

### `<Sidebar>`, `<Topbar>`, `<Breadcrumbs>`

Usados por `<AppLayout>`. No se usan directamente en pages.

---

## feedback

| Componente      | Cuándo                                            |
|-----------------|---------------------------------------------------|
| `<Spinner>`     | Carga puntual (botón, sección)                     |
| `<Skeleton>`    | Carga de página, lista, tabla                      |
| `<ProgressBar>` | Progreso conocido (import, backup)                 |
| `<EmptyState>`  | Carga completada sin resultados                    |
| `<ErrorState>`  | Carga completada con error                         |
| `<Toaster>`     | Una sola instancia en `<AppLayout>`, no se duplica |

---

## charts

Wrappers de `recharts` con tokens del design system.

```tsx
<LineChart data={[{ label: 'Lun', value: 100 }, ...]} />
<BarChart data={...} formatY={(v) => formatCurrency(v)} colors={['#...', ...]} />
<PieChart data={...} />
```

Datos: `{ label: string, value: number }[]`.

---

## layout

- `<PageContainer>` — wrapper estándar con `p-6` y `max-w-7xl`.
- `<PageHeader title subtitle actions>` — header de página.
- `<Section title description actions>` — bloque de contenido.
- `<Stack direction gap>` — flexbox.
- `<Grid cols>` — grid de 12.

---

## money

- `<MoneyField>` — input de moneda (ver `form`).
- `<MoneyDisplay value signed />` — span formateado, `signed` colorea verde/rojo.

---

## date

- `<DateField>` — input date (ver `form`).
- `<DateRangeField>` — dos `DateField`.

---

## misc

- `<DropdownMenu>` + sub-componentes — menú contextual sobre Radix.
- `<Separator>` — línea horizontal/vertical.
- `<Tooltip>` + `<TooltipContent>` + `<TooltipTrigger>` + `<TooltipProvider>` — tooltip Radix.

---

## auth

- `<Can permission anyPermission role fallback>` — render condicional por permisos.
- `<PermissionGate permission fallback>` — atajo para `<Can permission>`.

```tsx
<Can permission={Permissions.Customers.Delete} fallback={<DisabledDelete />}>
  <Button variant="destructive">Eliminar</Button>
</Can>
```
