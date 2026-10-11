# vfinancy: Backend y base de datos (v5)

ERP de escritorio para un negocio pequeño de importaciones. Se compra en el extranjero (mayormente USD) y se vende en soles (PEN). Un solo almacén, un solo usuario. La interfaz está en `vfinancy-frontend.md`.

## 1. Stack y estructura

- **Go + Wails** (bindings hacia el frontend) y **SQLite** local. Driver sin CGO (`modernc.org/sqlite`) para compilar fácil en Windows.
- SQLite con `foreign_keys=ON`, `journal_mode=WAL` y `busy_timeout`.
- Migraciones SQL versionadas y embebidas en el binario, aplicadas al abrir la app.
- Fechas como `TEXT` ISO (`YYYY-MM-DD`); marcas de tiempo ISO UTC.
- Paquetes sugeridos: `money`, `tc`, `codes`, `backup`, `store`, `service`. Sin capas adicionales.
- **Tests:** ninguno forma parte de la app. Las verificaciones (p. ej. de conversión de dinero) se hacen con archivos temporales fuera del repositorio o en una rama descartable y se eliminan al cumplir su propósito. No hay archivos `_test.go`, fixtures ni dependencias de testing en el proyecto.

## 2. Seguridad y acceso

- Contraseña opcional, guardada con hash (argon2id). Sin contraseña, la app abre directo.
- "Recuérdame" (ajuste local): al abrir, la app inicia desbloqueada. "Bloquear" vuelve a pedir la contraseña.
- Mientras la app está bloqueada, el backend rechaza todas las llamadas de datos (el bloqueo se aplica en el backend, no solo en la interfaz).
- La base de datos y los backups no se cifran; la app lo informa al activar una contraseña.

## 3. Dinero, monedas y TC

**Dinero**
- SQLite: `INTEGER` en céntimos (`*_cents`). Go: `int64` (`type Money int64`). **Nunca `float`** en base de datos, backend ni frontend.
- Máximo 2 decimales. Límite por monto: 99,999,999.99 (evita desbordes en `int64` al convertir).
- Columnas con `CHECK (x >= 0)` y `CHECK (currency IN ('PEN','USD'))`.

**TC (tipo de cambio)**
- Siempre expresado como **soles por 1 dólar**, guardado como `INTEGER` escalado ×10 000 (3.7512 → `37512`) en la columna `tc`. Rango válido: 0.5000 a 20.0000.
- **Cada registro con montos guarda su propio `tc`**, incluso si está en PEN (se necesita para mostrarlo en USD). Las líneas de un documento (ítems, extras) usan el `tc` del documento.
- Registros con `tc`: compras, lotes, ventas/envíos, pagos (cobros y adelantos), gastos de tarjeta y valores de referencia (costo y precio por defecto del producto, límite de tarjeta). En estos últimos el `tc` se toma del TC por defecto vigente al guardar y no se muestra en el formulario.
- No existe un TC global de la app. Cambiar un `tc` solo afecta a ese registro.

**Conversión**
- Función única `Convert(amount, from, to, tc)`: USD→PEN = `amount × tc / 10000`; PEN→USD = `amount × 10000 / tc`. Redondeo half-up (away from zero) solo en esta función, con aritmética entera.
- Mismo `from` y `to`: devuelve el monto exacto.
- Los agregados (dashboard, saldos, totales mensuales, deuda de tarjetas) **convierten cada registro con su `tc`** y luego suman. Los listados y totales aceptan `displayCurrency` y devuelven los montos ya convertidos junto con el original.

**TC por defecto (API)**
- Fuente: API gratuita sin clave (`open.er-api.com/v6/latest/USD`, tasa `PEN`), consultada desde el backend con timeout de 5 s. Se redondea a 4 decimales.
- Se guarda el último valor correcto (`tc_last`, `tc_last_at`) en `settings`.
- **Valor por defecto** de un registro nuevo: último valor de la API. Si nunca hubo uno, el **TC de respaldo** configurable (`tc_fallback`, por defecto 3.7500).
- `FetchTC()`: consulta la API en el momento (la usa el botón de cada formulario). Si falla, devuelve error y el frontend rellena el campo con el respaldo solo si está vacío.
- **Consulta periódica:** al iniciar la app y luego según la frecuencia configurada (`tc_refresh`: Manual, 1 h, 6 h por defecto, 12 h, 24 h). Un fallo no interrumpe la app ni cambia el último valor.
- Los registros existentes **nunca** se actualizan con nuevos valores de la API.

## 4. Códigos

- Todo identificador se **autogenera** y es editable: compra `C-0001`, venta `V-0001`, envío `E-0001`, lote manual `L-0001`, producto (SKU) `P-0001`, cobro `R-0001`.
- Lotes de compra: código de la compra + sufijo (`C-0001-1`, `C-0001-2`).
- `NextCode(entity)` devuelve el siguiente código libre (sin reservarlo) para precargar el campo del formulario. Al guardar, el código debe ser único; si ya existe, se devuelve un error de validación sobre ese campo.
- Tabla `sequences(prefix, last)`: se actualiza al guardar un código con formato `prefijo-número` mayor que `last`. Pueden existir saltos.
- Los códigos de contacto (clientes, proveedores, tarjetas) no existen; se identifican por nombre.
- **Código de seguridad** del envío: 4 dígitos aleatorios (`0000`–`9999`), editable, texto plano.

## 5. Modelo de datos

Todas las tablas tienen `id INTEGER PRIMARY KEY`, `created_at` y `updated_at`. Las entidades con documentos tienen `voided_at` (nulo = vigente): **no hay borrado físico**.

**Configuración**
- `settings(key, value)`: `password_hash`, `remember_me`, `theme`, `display_currency`, `tc_last`, `tc_last_at`, `tc_fallback`, `tc_refresh`, `lot_countdown_days` (0 = desactivado), `overdue_days` (30), `backup_dir`, `backup_keep` (14).
- `sequences(prefix, last)`.

**Contactos y catálogo**
- `clients(name, phone, note)` y `client_addresses(client_id, region, address)`.
- `suppliers(name, phone, note)`.
- `products(sku UNIQUE, name, default_cost_cents, default_cost_currency, default_cost_tc, default_price_cents, default_price_currency, default_price_tc)` y `product_suppliers(product_id, supplier_id)` (se vincula al comprar).
- `cards(name, last4, cut_day, due_day, limit_cents NULL, limit_currency, limit_tc)`.

**Compras**
- `purchases(code UNIQUE, supplier_id, date, payment_method ['card','cash','wallet'], card_id NULL, currency, tc, status ['pending','received'], for_client_id NULL, received_at NULL, voided_at)`.
- `purchase_items(purchase_id, product_id, description, qty, unit_cost_cents)`.
- `purchase_extras(purchase_id, concept, amount_cents, currency)`: usa el `tc` de la compra si la moneda difiere.

**Inventario**
- `lots(code UNIQUE, product_id, purchase_item_id NULL, source ['purchase','manual'], supplier_id NULL, qty_initial, unit_cost_cents, currency, tc, entry_date, countdown_days NULL, countdown_start, reserved_client_id NULL, voided_at)`.
  - `countdown_days`: NULL = usa el valor global; 0 = desactivado para este lote.
  - Lotes de compra: costo, moneda y `tc` se copian de la compra al recibirla y se sincronizan si la compra se edita.
  - Lotes manuales: se ingresan sin compra (ver 6.3).
- Disponible = `qty_initial` − unidades en ventas/envíos vigentes. Se calcula en la vista `lot_stock`; no se guarda.
- `kardex` es una **vista** (entradas: lotes; salidas: ítems de venta/envío vigentes) con saldo acumulado por producto mediante window function. No hay tabla de movimientos.

**Ventas y envíos** (misma tabla, separados por `kind`)
- `sales(code UNIQUE, kind ['sale','shipment'], client_id NULL, date, currency, tc, shipment_status NULL ['prepared','sent','delivered'], security_code NULL, ship_region NULL, ship_address NULL, voided_at)`.
- `sale_items(sale_id, product_id, lot_id, description, qty, unit_price_cents)`.
- `sale_extras(sale_id, concept, amount_cents, currency)`.

**Pagos**
- `payments(code UNIQUE, kind ['payment','advance'], client_id NULL, sale_id NULL, purchase_id NULL, date, method, amount_cents, currency, tc, advance_status NULL ['active','applied','refunded'], voided_at)`.
  - `payment`: cobro de una venta/envío, en la moneda de la venta. El "Pagado ahora" del registro crea un `payment` con la fecha de la venta.
  - `advance`: adelanto de un encargo (cliente y compra). Al aplicarse se vincula a la venta; si las monedas difieren, se convierte con su propio `tc` para el saldo.

**Tarjetas**
- `card_expenses(card_id, date, concept, amount_cents, currency, tc, voided_at)`.
- `card_cycle_payments(card_id, cycle_end, paid_date)`: los ciclos se calculan; solo se guarda el pago.

## 6. Reglas de negocio

### 6.1 Compras
- Total = subtotal de ítems + extras (convertidos a la moneda de la compra con su `tc`), **sin prorrateo**.
- Solo se puede recibir con al menos un ítem. Al marcar Recibida (en una transacción): se crea un lote por ítem, el costo por defecto del producto toma el último costo, se vinculan proveedor y producto, y si la compra es para un cliente los lotes quedan reservados.
- Extras editables en cualquier estado. Si la compra ya fue recibida, el frontend avisa del efecto en utilidades.
- No se puede reducir una cantidad por debajo de lo ya vendido ni anular una compra con lotes vendidos (primero se anulan esas ventas).

### 6.2 Ventas y envíos
- Total = suma de ítems + extras (convertidos con el `tc` del documento).
- Saldo = total − pagos vigentes (− adelantos aplicados).
- Lote por ítem: por defecto el más antiguo con stock (FIFO). Si la cantidad supera el lote, `SuggestLots(product, qty)` devuelve la división por lotes siguientes; el frontend la propone como líneas adicionales.
- Validar en transacción que la cantidad no supere el disponible del lote. Un lote reservado solo se puede vender al cliente reservado.
- Descuenta stock al crear (también envíos). Anular devuelve stock.
- Envío: región/dirección obligatorias, estado inicial `prepared`, código de seguridad generado.
- "Vender a cliente" (encargo): crea la venta/envío con el lote reservado y los adelantos vigentes del cliente para esa compra aplicados.
- Cancelar encargo: el adelanto pasa a `refunded` manualmente y el lote deja de estar reservado.

### 6.3 Lotes manuales
- Para stock que no viene de una compra registrada: producto, cantidad, costo unitario, moneda, `tc`, fecha de ingreso, proveedor (opcional) y cuenta regresiva opcional.
- Aparecen en el kardex como "Ingreso manual". No tienen costos extras.
- Son editables mientras no tengan ventas; con ventas, solo se editan datos que no alteren el stock vendido.

### 6.4 Cuenta regresiva y remate
- Días efectivos = `countdown_days` del lote, o el global si es NULL. Si es 0, no aplica.
- **En remate** se calcula por fecha: `hoy ≥ countdown_start + días` y stock disponible > 0. No hace falta un proceso en segundo plano.
- "Quitar remate" fija `countdown_days = 0`. "Reiniciar contador" fija `countdown_start = hoy`.

### 6.5 Tarjetas
- Cada compra con tarjeta (no anulada) entra al ciclo de su fecha: si el día de la compra ≤ día de corte, al ciclo que cierra ese mes; si no, al del mes siguiente. Los gastos manuales siguen la misma regla. Cada cargo se convierte con su `tc`.
- Fecha de pago del ciclo: el primer `due_day` posterior al corte.
- Deuda actual = suma de ciclos no pagados.

### 6.6 Utilidad (dashboard)
- Se reconoce en la **fecha de la venta** (devengado).
- Ventas = total de ventas/envíos vigentes del período.
- Costo base = Σ (unidades vendidas del lote × costo unitario del lote), cada uno convertido con el `tc` del lote.
- Costo extra = por cada compra con ventas en el período: extras de la compra × (unidades vendidas en el período / unidades totales de la compra), convertidos con el `tc` de la compra. Los lotes manuales no tienen extras.
- Utilidad = ventas − costo base − costo extra. Los gastos manuales de tarjeta **no** entran.
- Cada componente se convierte con el `tc` de su propio registro (no con el de otro documento).

### 6.7 Estados derivados
Se calculan en las consultas (no se guardan) y devuelven un solo estado por registro, con el primero que se cumpla:

| Entidad | Estados |
|---|---|
| Compra | Anulada → Pendiente → Vendida (todos sus lotes con stock 0) → Reservada (para cliente y sin unidades vendidas) → Recibida (incluye "vendidas X/Y") |
| Lote | Agotado → Reservado → En remate → Activo |
| Venta | Anulada → Atrasada (saldo > 0 y fecha anterior a `overdue_days`) → Saldo pendiente → Pagada |
| Envío | Anulado → Preparado → Enviado → Entregado; además `saldo` si es mayor a 0 |
| Adelanto | Devuelto → Aplicado → Vigente |
| Ciclo de tarjeta | Pagado → Vencido (hoy > fecha de pago) → Por pagar (hoy > corte) → Abierto |
| Cliente | "Con saldo" si su saldo > 0 |

## 7. API (bindings de Wails)

- **Listados** `List<Entidad>(query)`: `query` = `{filters: [{field, op, value}], displayCurrency, monthsLimit, cursorMonth}`. Devuelve **meses** (`month`, `count`, `total` convertido, `rows`), del más reciente al más antiguo. Se cargan los últimos meses y se piden más con el cursor.
  - Cada entidad define una lista blanca de campos filtrables con su tipo (`text`, `number`, `date`, `option`) y los operadores permitidos:
    - `text`: contiene, es igual.
    - `number`: =, >, <, entre. Los montos se comparan ya convertidos a `displayCurrency`.
    - `date`: en el mes, antes, después, entre.
    - `option`: es, es alguno de.
  - Los filtros se combinan con AND. Las consultas usan parámetros, nunca texto concatenado.
- `GetFilterSchema(entity)`: devuelve los campos filtrables, tipos y operadores para construir la barra de filtros.
- CRUD por entidad, con `Void`/`Restore` en lugar de `Delete`. Operaciones específicas: `ReceivePurchase`, `AdvanceShipment`, `AddPayment`, `AddAdvance`, `RefundAdvance`, `SellToClient`, `RestartCountdown`, `DismissAuction`, `MarkCyclePaid`.
- `Quick-create` (cliente, proveedor, producto, tarjeta, dirección): mínimos campos, devuelve la entidad creada.
- `Suggest(entity, text)`: autocompletado por texto (incluye conceptos de extras ya usados).
- `NextCode(entity)`, `FetchTC()`, `GetDashboard(period, anchor, displayCurrency)`.
- Todas las validaciones devuelven errores por campo (`{field, message}`).

## 8. Dashboard

`GetDashboard` devuelve, para el período (mes, trimestre, año o total) y la moneda de visualización:
- Ventas, costos (base y extras), utilidad y extras por concepto.
- Serie por mes (en trimestre, año y total) con ventas, costo base, costo extra y utilidad.
- Cuentas por cobrar: ventas y envíos con saldo, agrupados por mes.
- Lotes en remate: producto, stock y días en remate.
- Alertas: compras Pendientes con más de `overdue_days`, ciclos Por pagar o Vencidos, ventas Atrasadas.

## 9. Backups

- Automático **una vez al día al cerrar la app**, en `backup_dir`, con `VACUUM INTO` (copia consistente). Nombre `vfinancy-YYYYMMDD-HHMM.db`. Se conservan los últimos `backup_keep` (14).
- "Respaldar ahora" ejecuta el mismo proceso.
- "Restaurar": el usuario elige un archivo; la app crea antes un backup de seguridad, cierra la base, reemplaza el archivo y reinicia.

## 10. Decisiones clave

| Tema | Decisión |
|---|---|
| Dinero | `INTEGER` en céntimos, `int64` en Go, sin `float` |
| TC | Entero ×10 000 por registro; sin TC global; API solo para valores por defecto |
| Conversión | Con el TC de cada registro; agregados suman montos ya convertidos |
| Códigos | Autogenerados y editables; únicos |
| Utilidad | Por fecha de venta; extras proporcionales a lo vendido |
| Remate | Calculado por fecha; solo una etiqueta |
| Encargo | Atributo de compra y venta |
| Estados | Derivados, uno por registro |
| Borrado | Nunca físico; se anula |
| Tests | Solo temporales, no se conservan |

## 11. Fuera de alcance (v1)

Multi-almacén, IGV y facturación electrónica, más de dos monedas, reportes avanzados, importar/exportar masivo, multiusuario, nube, cifrado de la base y prorrateo de costos.
