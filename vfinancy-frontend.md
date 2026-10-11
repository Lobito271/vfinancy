# vfinancy: Frontend y diseño (v5)

Interfaz de la app de escritorio (Wails). Las reglas de negocio, datos y API están en `vfinancy-backend.md`.

## 1. Stack

- React + TypeScript + Vite (plantilla de Wails). Estado local simple; sin librerías de estado ni de UI externas.
- Gráfico de barras propio en SVG.
- **Sin tests en el proyecto:** no se incluyen Vitest, Jest, Playwright ni archivos de prueba. Cualquier verificación es temporal y se elimina.
- Idioma español (Perú): fechas `DD/MM/YYYY`, meses en español, montos `S/ 1,250.00` y `$ 330.00`.

## 2. Diseño visual

**Estilo:** flat, minimalista, escandinavo. Bordes cuadrados (`border-radius: 0`), sin sombras, sin adornos, cuadrícula de 4 px, bordes de 1 px.

**Paleta principal:** negro `#0a0a0a` y amarillo `#f5c518`.

| Token | Tema oscuro | Tema claro |
|---|---|---|
| Fondo | `#0a0a0a` | `#fafafa` |
| Superficie | `#141414` | `#ffffff` |
| Borde | `#2a2a2a` | `#e2e2e2` |
| Texto | `#f2f2f2` | `#0a0a0a` |
| Texto secundario | `#9a9a9a` | `#6b6b6b` |
| Acento | `#f5c518` | `#f5c518` |

- El amarillo es el único acento: botón primario, ítem activo, foco (outline de 2 px) y selección. **Siempre con texto `#0a0a0a`** encima; nunca texto amarillo sobre fondo claro.
- Tema: claro u oscuro. Se cambia en Ajustes. (Se eliminó el modo automático; el tema por defecto es claro.)

**Tipografía** (solo sans-serif y monospace, gestionadas con **Fontsource**, empaquetadas en la app, sin CDN):
- Títulos: `@fontsource-variable/montserrat` (`'Montserrat Variable'`).
- Interfaz: `@fontsource-variable/figtree` (`'Figtree Variable'`), con respaldo `system-ui, sans-serif`.
- Códigos, montos, TC y SKU: `@fontsource-variable/roboto-mono` (`'Roboto Mono Variable'`), con respaldo `ui-monospace, monospace`, números tabulares.
- Pesos: 400 / 500 / 600 / 700. Escala: 11 / 12 / 14 (base) / 16 / 20 / 24 px.

**Estados visuales** (ver sección 5): 6 colores fijos, independientes del acento.

## 3. Layout y responsive

- **Barra lateral** (izquierda): Inicio, Compras, Inventario, Ventas, Envíos, Tarjetas, Clientes, Proveedores, Ajustes y **Bloquear**. Ítem activo con acento.
- **Topbar:** título de la página y **switch de moneda** `S/ | $` (control segmentado). No hay búsqueda global ni TC en el topbar.
- **Contenido:** ancho máximo 1100 px, centrado.
- **Responsive:**
  - ≥ 1100 px: barra lateral expandida.
  - 760–1100 px: barra lateral solo con iconos.
  - < 760 px: barra lateral oculta tras un botón de menú; formularios a una columna; el resumen de cada accordion pasa a dos líneas.
  - Ancho mínimo de ventana: 480 px.
- **Atajos:** `Ctrl+Enter` guarda, `Esc` cierra un modal, `Ctrl+M` cambia la moneda, `Ctrl+N` crea un registro en la página actual.

## 4. Moneda de visualización

- El switch del topbar es persistente y cambia al instante todos los montos de la app: listas, detalles, totales de formularios, dashboard, kardex, tarjetas y cuentas por cobrar.
- Los montos se piden al backend con `displayCurrency`; el frontend **no calcula conversiones**.
- Si la moneda del switch coincide con la del registro, se muestra el monto exacto. Si no, el convertido, con el original en texto secundario (`$ 330.00` / `S/ 1,250.00 · TC 3.7900`).
- Los campos de entrada conservan la moneda del documento. Subtotales, extras, total y saldo del formulario se muestran en la moneda del switch.
- **Dinero sin `float`:** el texto se convierte a céntimos con manipulación de cadenas y se formatea desde enteros. El `<MoneyInput>` solo acepta hasta 2 decimales y un máximo de 99,999,999.99.

## 5. Estados visuales

**Reglas**
- Cada registro muestra **un solo estado** o ninguno, el que entrega el backend.
- Se representa con una **badge** (texto + borde y texto del color, sin relleno) y el fondo de la fila teñido.
- Los anulados van en gris con texto tachado. Los datos secundarios (saldo, progreso "3/10 vendidas") van como texto junto a la badge.

| Color | Significado | Claro | Oscuro |
|---|---|---|---|
| Gris | Inactivo, cerrado, anulado | `#6b6b6b` | `#9a9a9a` |
| Ámbar | Pendiente, requiere seguimiento | `#b45309` | `#fb923c` |
| Azul | En curso | `#1d4ed8` | `#6ea8ff` |
| Verde | Completado | `#15803d` | `#4ade80` |
| Rojo | Atrasado o vencido | `#c62828` | `#ff6b6b` |
| Violeta | Reservado para cliente | `#6d28d9` | `#b79cff` |

| Entidad | Estados |
|---|---|
| Compra | Anulada (gris), Pendiente (ámbar), Vendida (verde), Reservada (violeta), Recibida (azul) |
| Lote | Agotado (gris), Reservado (violeta), En remate (rojo), Activo (sin badge) |
| Venta | Anulada (gris), Atrasada (rojo), Saldo pendiente (ámbar), Pagada (verde) |
| Envío | Anulado (gris), Preparado (ámbar), Enviado (azul), Entregado (verde); con saldo, texto ámbar "Saldo S/ X" |
| Adelanto | Devuelto (gris), Aplicado (verde), Vigente (azul) |
| Ciclo de tarjeta | Pagado (verde), Vencido (rojo), Por pagar (ámbar), Abierto (azul) |
| Cliente | "Con saldo" (ámbar); si no, sin badge |

Sin estado visual: productos, proveedores, cobros posteriores, gastos manuales y backups.
El filtro de estado de cada página usa estas mismas badges.

## 6. Cómo se muestran los registros

**Agrupación por mes:** todos los listados se agrupan por mes (del más reciente al más antiguo) con un encabezado fijo `Octubre 2026 · N registros · total`. Dentro del mes, ordenados por fecha descendente. Se cargan los últimos meses y un botón "Cargar meses anteriores" trae más. Los encabezados de mes se pueden contraer.

**Elección del componente**

| Información | Componente | Páginas |
|---|---|---|
| Documentos con detalle (ítems, extras, pagos, acciones) | **Lista accordion**: el resumen muestra código, fecha, contraparte, total y badge; al hacer click se expande el detalle | Compras, Ventas, Envíos, Ciclos de tarjeta, Encargos ("Para clientes") |
| Elementos con progreso o vida útil | **Accordion con barra de progreso** (vendido/total) y días restantes | Lotes |
| Entidades de consulta con ficha propia | **Lista simple de filas** que abre su página de detalle | Productos, Clientes, Proveedores |
| Libro de movimientos numéricos con saldo acumulado | **Tabla** (única excepción justificada) | Kardex en el detalle de producto |
| Pocos elementos con métricas | **Tarjetas** | Tarjetas de crédito (deuda y límite), resumen del dashboard |
| Líneas de ítems y extras dentro de un detalle | **Filas compactas** con alineación numérica | Detalle de documentos |

- El accordion permite varios paneles abiertos. Los códigos, clientes, proveedores, productos y lotes son **enlaces**: llevan a la página correspondiente y abren el registro (por ejemplo `/ventas?abrir=V-0012`).
- Cada accordion muestra sus acciones al expandirse: Editar, Anular/Restaurar y las específicas (Marcar recibida, Registrar cobro, Registrar adelanto, Vender a cliente, Avanzar estado de envío, Imprimir comprobante). Editar aparece en todo registro no anulado; el backend valida qué se puede modificar (por ejemplo, en compras los ítems vendidos no se alteran).
- Los registros anulados permanecen en la lista, en gris.

## 7. Búsqueda y filtros por página

- **No hay búsqueda global.** Cada página tiene su propia barra de filtros, construida con `GetFilterSchema`.
- La barra tiene: **[Campo ▾] [Operador ▾] [Valor]**. El tipo de campo determina el operador y el control del valor:

| Tipo | Operadores | Control del valor |
|---|---|---|
| Texto | contiene, es igual | Campo de texto |
| Número / monto | =, >, <, entre | Campo numérico (los montos se comparan en la moneda del switch) |
| Fecha | en el mes, antes, después, entre | Selector de mes o de fecha |
| Opción (estado, método, origen) | es, es alguno de | Selector con badges |

- Campo por defecto: el código (o el nombre). Se pueden sumar más filtros con **"+ Filtro"**; cada uno se muestra como chip removible, combinados con AND. Hay un botón "Limpiar".
- Los filtros se aplican al escribir (con espera corta) y se mantienen al cambiar de página dentro de la sesión.

| Página | Campos filtrables |
|---|---|
| Compras | Código, Proveedor, Cliente (para cliente), Fecha, Total, Método de pago, Estado |
| Ventas | Código, Cliente, Producto, Fecha, Total, Saldo, Estado |
| Envíos | Código, Cliente, Región, Fecha, Total, Saldo, Estado |
| Lotes | Código, Producto, Fecha de ingreso, Disponible, Costo unitario, Días restantes, Origen, Estado |
| Productos | Nombre, SKU, Proveedor, Stock, Costo, Precio |
| Clientes | Nombre, Teléfono, Saldo |
| Proveedores | Nombre, Teléfono |
| Tarjetas (ciclos) | Tarjeta, Fecha de corte, Total, Estado |
| Kardex | Fecha, Movimiento, Cantidad |

## 8. Formularios

**Reglas generales**
- Cada campo tiene un **label visible** sobre el campo. El placeholder nunca sustituye al label. Los obligatorios llevan `*`.
- Sin descripciones ni textos de ayuda bajo los campos. Solo un mensaje breve de error en línea cuando el valor no es válido (el error viene del backend por campo).
- Valores por defecto: fecha de hoy, código siguiente, TC por defecto, FIFO, último método de pago, moneda del documento (compras USD; ventas y envíos PEN).
- Todos los registros (compras, ventas, envíos, lotes manuales, cobros, adelantos, gastos y entidades auxiliares) se crean y editan en **modales**. El modal tiene encabezado y pie fijos con el cuerpo desplazable; las acciones (Cancelar / Guardar) van en el pie, nunca dentro del formulario.
- `Ctrl+Enter` guarda. Los formularios son de una columna; en pantallas anchas, dos columnas para campos cortos.

**Componente por tipo de dato**

| Dato | Componente |
|---|---|
| Entidad (cliente, proveedor, producto, tarjeta, dirección) | **Autocomplete** con "Agregar nuevo" |
| Lote por ítem | Select de lotes con stock, el más antiguo preseleccionado (sin "Agregar nuevo") |
| 2–3 opciones fijas (moneda, método de pago, tipo de cobro) | Control segmentado |
| Opciones fijas más largas | Select |
| Concepto de extra | Campo de texto con sugerencias de conceptos usados |
| Monto | `<MoneyInput>` con símbolo de la moneda del registro |
| TC | `<TcInput>` (ver abajo) |
| Fecha | Selector de fecha nativo |
| Cantidad | Campo numérico entero |
| Código | Campo de texto en fuente monoespaciada, precargado y editable |
| Nota | Área de texto |
| Opción sí/no ("Para cliente", ajustes) | Checkbox o interruptor |

**Autocomplete con "Agregar nuevo"**
- Todo campo que selecciona una entidad que se puede crear en ese contexto muestra al final de la lista la opción **"Agregar nuevo"** (con el texto escrito precargado como nombre).
- Abre el **modal de creación** de esa entidad. Al guardar, el modal se cierra, la entidad queda seleccionada y el formulario actual **conserva todo lo ingresado**.
- Se aplica a: proveedor, cliente, producto y tarjeta (en compras, ventas, envíos y lotes manuales) y dirección de envío (en envíos).

**`<TcInput>` (reutilizable)**
- Label `TC (S/ por $)`, campo de texto de hasta 4 decimales y botón **"Consultar"** que llama a `FetchTC()`. Editable a mano siempre.
- Estado de carga en el botón; si la API falla, error breve en línea y el campo se llena con el TC de respaldo solo si estaba vacío.
- Se precarga con el TC por defecto del backend al abrir el formulario.
- Aparece en todo formulario con montos: compra, venta/envío, cobro, adelanto, gasto de tarjeta e ingreso manual de lote. No aparece en productos ni tarjetas.

**Formularios específicos**
- **Compra:** código, proveedor, fecha, método de pago, tarjeta (si el método es Tarjeta), moneda, TC, "Para cliente" y cliente (si aplica), ítems (producto, cantidad, costo unitario), extras (concepto, monto, moneda). Total en la moneda del switch.
- **Venta / Envío:** código, cliente (opcional, "Cliente general"), fecha, moneda, TC, ítems (producto, lote, cantidad, precio), extras, "Pagado ahora". Si la cantidad supera el lote, se agregan líneas con los siguientes lotes. Envío añade dirección, región y código de seguridad (editable).
- **Lote manual:** código, producto, cantidad, costo unitario, moneda, TC, fecha de ingreso, proveedor. La cuenta regresiva de lotes usa el valor global de Ajustes (no se define por lote).
- **Cobro / adelanto:** código, monto, fecha, método, moneda, TC.
- **Gasto de tarjeta:** concepto, monto, moneda, fecha, TC.
- **Producto:** nombre, SKU, costo por defecto (monto y moneda), precio por defecto (monto y moneda).
- **Cliente / Proveedor:** nombre, teléfono, nota; cliente también direcciones (región, dirección).
- **Tarjeta:** nombre, últimos 4 dígitos, día de corte, día de pago, límite (opcional).

## 9. Páginas

- **Inicio (dashboard):** selector `Mes | Trimestre | Año | Total` con anterior/siguiente.
  - Tarjetas: Ventas, Costos y Utilidad.
  - Desglose de costos (base y extras) y extras por concepto.
  - Gráfico de barras Ventas − Costo base − Costos extras = Utilidad, con detalle al pasar el cursor. En períodos de más de un mes, barras agrupadas por mes.
  - Cuentas por cobrar y Lotes en remate (listas con badge y enlace, agrupadas por mes las primeras).
  - Alertas ligeras: compras Pendientes antiguas, ciclos Por pagar o Vencidos y ventas Atrasadas.
- **Compras:** pestañas "Todas" y "Para clientes" (encargos). Acordeón por mes.
- **Inventario:** pestañas "Lotes" (con botón "Ingresar lote") y "Productos". Lotes con filtro de estado y días restantes visibles.
- **Producto (detalle):** stock total, lotes con badge y kardex agrupado por mes (tabla).
- **Ventas / Envíos:** acordeón por mes. En envíos, botón para avanzar de estado e imprimir el comprobante con el código de seguridad (`window.print()` con hoja de estilos de impresión).
- **Tarjetas:** una tarjeta por cada tarjeta de crédito (deuda y límite) y sus ciclos como accordion agrupado por mes, con los cargos, gastos manuales ("Agregar gasto") y "Marcar como pagado".
- **Clientes y Proveedores:** lista y ficha con historial por mes; en clientes, saldo por cobrar y direcciones.
- **Ajustes:** tema, contraseña y "Recuérdame", moneda por defecto del switch, TC (frecuencia de actualización, TC de respaldo y último valor consultado), días de cuenta regresiva global, días de atraso, backups (carpeta, cantidad, "Respaldar ahora", "Restaurar").
- **Pantalla de bloqueo:** campo de contraseña y botón; también al abrir la app si hay contraseña y no está activo "Recuérdame".

## 10. Decisiones clave

| Tema | Decisión |
|---|---|
| Agrupación | Por mes, con total mensual |
| Registros | Accordion por defecto; tabla solo para el kardex |
| Búsqueda | Por página, con campo, operador y valor según tipo de dato |
| Moneda | Switch solo de visualización; conversión en el backend con el TC de cada registro |
| TC | Componente `<TcInput>` en cada formulario con montos |
| Entidades | Autocomplete + "Agregar nuevo" en modal que conserva el formulario |
| Acento | Amarillo con texto negro; el ámbar de estados es otro tono |
| Tipografía | Montserrat (títulos) + Figtree (interfaz) + Roboto Mono (códigos y montos) vía Fontsource |
| Tests | No hay tests en el proyecto |

## 11. Fuera de alcance (v1)

Búsqueda global, más de dos monedas, reportes avanzados, importar/exportar masivo, multiusuario, nube y personalización de colores.
