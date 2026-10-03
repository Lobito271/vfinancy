# vfinancy — ERP para negocio de importación

vfinancy es un ERP de escritorio para una empresa de importación. Reúne en un solo lugar todo el
ciclo comercial: desde la **compra al proveedor en el exterior** hasta la **entrega al
cliente en el Perú**, pasando por el ingreso de mercadería al inventario, el control de remates, la
cobranza y el pago de tarjetas.

Toda la información se guarda **en tu equipo**, en una base local. La aplicación funciona sin
internet: solo necesitas conexión cuando se consulta el tipo de cambio del día o cuando decides
sincronizar con un servidor en la nube (opcional, ver [§14](#14-copias-de-seguridad-y-nube)).

No hay usuarios, roles ni permisos. Hay **una sola persona usuaria** y, si quieres, una contraseña
local que protege el acceso a la información.

---

## Índice

1. [Qué resuelve vfinancy](#1-qué-resuelve-vfinancy)
2. [Mapa de las áreas](#2-mapa-de-las-áreas)
3. [Primer arranque: configuración inicial](#3-primer-arranque-configuración-inicial)
4. [Acceso, contraseña y seguridad](#4-acceso-contraseña-y-seguridad)
5. [Inicio: el panel](#5-inicio-el-panel)
6. [Proveedores](#6-proveedores)
7. [Productos](#7-productos)
8. [Compras: pedidos al proveedor exterior](#8-compras-pedidos-al-proveedor-exterior)
9. [Inventario](#9-inventario)
10. [Ventas](#10-ventas)
11. [Tesorería: tarjetas de crédito y ciclos](#11-tesorería-tarjetas-de-crédito-y-ciclos)
12. [Envíos](#12-envíos)
13. [Ajustes](#13-ajustes)
14. [Copias de seguridad y nube](#14-copias-de-seguridad-y-nube)
15. [El ciclo completo de una importación](#15-el-ciclo-completo-de-una-importación)
16. [Estados y transiciones](#16-estados-y-transiciones)
17. [Cálculos que hace el sistema por ti](#17-cálculos-que-hace-el-sistema-por-ti)
18. [Validaciones, bloqueos y dependencias](#18-validaciones-bloqueos-y-dependencias)
19. [Cómo se conectan las áreas](#19-cómo-se-conectan-las-áreas)
20. [Límites conocidos](#20-límites-conocidos)

---

## 1. Qué resuelve vfinancy

El problema típico de una importadora es que la información vive repartida: la compra al proveedor
en un correo, el flete en un chat, el costo de internación en una planilla, la mercadería en una
hoja de cálculo y la deuda del cliente en la cabeza del dueño. vfinancy reemplaza todo eso con un
único registro conectado.

Lo que el sistema hace por ti:

- **Una sola verdad por mercadería.** Cada venta dice de qué lote salió, a qué costo y por qué
  compra llegó ese lote. Puedes seguir el rastro completo hacia atrás.
- **Costo real en soles, no solo precio en dólares.** Las compras se capturan en dólares, pero cada
  lote que entra al inventario ya guarda su costo unitario convertido a soles con el tipo de cambio
  congelado en esa compra. Aparte, la compra muestra un **costo real** en soles que sí incluye el
  factor de importación que tú configuras.
- **Control de antigüedad.** Un lote que pasa los días de venta máxima aparece marcado como
  **remate** en el inventario y en el panel, para que lo saques antes de que pierda valor.
- **Cobranza y tarjetas conectadas.** La venta genera deuda del cliente; el pago de la tarjeta se
  descuenta del ciclo de facturación correspondiente.
- **Control de límites.** El sistema avisa cuando una compra supera el tope que definiste y cuando
  una venta pide mercadería que no tienes.

---

## 2. Mapa de las áreas

| Área | Para qué la usas | Registros principales |
|------|------------------|------------------------|
| **Inicio** | Ver el estado del mes: cobrado, utilidad, ventas por estado y lotes en remate | Indicadores consolidados |
| **Compras** | Registrar y dar seguimiento a las órdenes al proveedor exterior | Compra, ítems, costos extras |
| **Inventario** | Ver los lotes que tienes, su costo, su antigüedad y cada movimiento de stock | Lotes, movimientos (kardex) |
| **Ventas** | Registrar ventas al contado o al crédito y cobranzas | Venta, ítems de venta, cobros |
| **Tesorería** | Controlar tarjetas de crédito, sus ciclos de facturación y saldos | Tarjeta de crédito, ciclo, tipo de cambio |
| **Envíos** | Dar seguimiento a las salidas de mercadería con un código | Envío |
| **Ajustes** | Parámetros del negocio, datos de la empresa, seguridad, respaldos y nube | Preferencias, datos de empresa |

Las áreas de **Proveedores**, **Productos** y **Clientes** no tienen pantalla propia: se administran
desde drawers (paneles laterales) que se abren desde Compras, Inventario y Ventas respectivamente.

---

## 3. Primer arranque: configuración inicial

Al abrir la aplicación por primera vez (sin datos de empresa) aparece el asistente
**"Tu operación empieza aquí"**. Es una sola pantalla y solo se muestra una vez.

### Paso 1 — Tu empresa

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Razón social** | **Sí** (mínimo 2 caracteres) | Nombre legal de tu empresa. Se usa en la documentación del negocio |
| Nombre comercial | No | Nombre con el que operas comercialmente |
| RUC | No | Debe tener 11 dígitos y empezar con 10 o 20 |
| Correo electrónico | No | Correo de la empresa |
| Dirección fiscal | No | Domicilio fiscal |
| Teléfono | No | Teléfono de la empresa |
| Web | No | Sitio web de la empresa |
| Contraseña | No | Si la escribes, la aplicación pedirá contraseña al iniciar. Mínimo 8 caracteres |

> Todo lo que no sea obligatorio aquí **se puede completar después** en `Ajustes → Datos de la
> empresa`.

Botón: **Comenzar**.

- Si dejaste la contraseña vacía, entras directo al panel.
- Si definiste una contraseña, el asistente pasa al **paso 2**.

### Paso 2 — Pregunta de seguridad (solo si definiste contraseña)

Sirve para recuperar el acceso si olvidas la contraseña.

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| Pregunta de seguridad | **Sí** | Puedes elegir una de la lista o escribir la tuya |
| Escribe tu pregunta | **Sí**, solo si elegiste la opción de escribir la tuya | Tu pregunta personalizada |
| Tu respuesta | **Sí** (mínimo 3 caracteres) | La respuesta. Se guarda cifrada |

Botón: **Guardar y continuar** → entras al panel.

---

## 4. Acceso, contraseña y seguridad

### Pantalla de bloqueo

Si configuraste contraseña, al abrir la aplicación verás la pantalla de bloqueo:

- **Contraseña** → obligatoria. Botón **Entrar**.
- **¿Olvidaste tu contraseña?** → se despliega un formulario con la pregunta de seguridad y dos
  campos: **Tu respuesta** y **Nueva contraseña** (mín. 8 caracteres). Botón **Recuperar acceso**.
  Al recuperarlo, la pregunta de seguridad se elimina.

**Bloqueo por intentos fallidos:** tras **5 intentos incorrectos** la aplicación se bloquea durante
**15 minutos**. La recuperación con la pregunta de seguridad funciona incluso durante el bloqueo.

### Dónde se gestiona

`Ajustes → Seguridad → Administrar autenticación` abre un panel con dos tarjetas:

**Contraseña local**

| Campo | Obligatorio | Notas |
|-------|-------------|-------|
| Contraseña actual | **Sí**, si ya existe contraseña | Se necesita para cambiarla o quitarla |
| Nueva contraseña | No | Mínimo 8 caracteres, al menos una mayúscula, una minúscula y un dígito |

Acciones: **Crear contraseña** / **Actualizar contraseña**, **Quitar contraseña** (deja el acceso
libre y borra la pregunta de seguridad) y **Bloquear ahora** (te devuelve a la pantalla de
bloqueo).

**Pregunta de seguridad** (solo si hay contraseña)

| Campo | Obligatorio |
|-------|-------------|
| Pregunta de seguridad | **Sí** |
| Escribe tu pregunta | **Sí**, solo si es personalizada |
| Tu respuesta | **Sí** (mín. 3 caracteres) |

Acciones: **Configurar pregunta** / **Cambiar pregunta** y **Eliminar pregunta**.

---

## 5. Inicio: el panel

Es la pantalla que ves al entrar. Tiene cinco bloques:

| Bloque | Qué te muestra |
|--------|----------------|
| **Ventas cobradas del mes** | Suma de todos los cobros del mes actual (excluye ventas anuladas) |
| **Utilidad del mes** | Utilidad del mes en curso: precio de venta − costo total del mismo mes |
| **Estado del mes** | Tres contadores sobre ventas con fecha en el mes actual: **Cobradas**, **Pendientes** (pendientes + parciales), **Anuladas** |
| **Utilidad mensual** | Últimos 12 meses, uno por fila, con los cuatro sumados del mes: **costo base**, **costos extras**, **costo total** (los tres en USD), **precio de venta** y **utilidad** (estos dos en soles) |
| **Productos en Remate** | Lotes que **ya entraron en remate** y todavía tienen existencias, ordenados por días restantes (los más urgentes primero). Cada fila muestra el producto, la cantidad y una etiqueta de estado |

La **utilidad** se calcula así:

```
Utilidad = Precio de venta − (Costo base + Costos extras)
```

El costo de cada compra entra en el mes de su **fecha de compra** (no la de llegada) y se convierte a
soles con el **tipo de cambio de esa compra**, de modo que el panel nunca discrepa de la lista de
Compras. Las compras anuladas y las ventas anuladas quedan fuera del cálculo. Cuando la utilidad es
negativa se muestra en rojo.

Si no hay nada que mostrar, cada bloque tiene su propio mensaje vacío (por ejemplo *"Sin datos de
utilidad"* o *"Sin productos en remate"*).

---

## 6. Proveedores

**Para qué:** registrar a quién le compras en el exterior, para poder seleccionarlo al crear una
compra.

**Dónde:** `Compras → Proveedores` (botón en la cabecera).

### Qué se guarda de cada proveedor

| Campo | Obligatorio al crear | Qué representa |
|-------|----------------------|----------------|
| **Nombre** | **Sí** | Razón del proveedor en el exterior. No puede repetirse |
| Persona de contacto | No | Persona con la que negocias |
| Teléfono | No | Teléfono de contacto |
| Correo | No | Correo de contacto |
| Dirección | No | Dirección del proveedor |

### Qué puedes hacer

- **Crear** (`Nuevo proveedor` en el pie del panel), **Editar** y **Eliminar**.
- **Buscar** por nombre, persona de contacto o teléfono.
- Los proveedores también se pueden **crear en línea** desde el formulario de compra, sin salir de
  él.
- Un proveedor marcado como **Inactivo** aparece con una etiqueta y **no puede seleccionarse en una
  compra nueva compra** (mensaje: *"El proveedor está inactivo"*). La aplicación no incluye un
  control para activar o desactivar proveedores: los que registras quedan activos.

> **No se puede eliminar** un proveedor que tenga compras asociadas (mensaje: *"No se
> puede eliminar un proveedor con compras asociadas"*). Desactívalo en su lugar.

---

## 7. Productos

**Para qué:** mantener el catálogo de lo que vendes, con su costo en dólares y su precio de venta
en soles. Es la referencia que usan las compras y las ventas.

**Dónde:** `Inventario → Productos` (botón en la cabecera). También se abre desde el ingreso de
stock y desde la venta, con creación en línea.

### Qué se guarda de cada producto

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Descripción** | **Sí** | Nombre del producto |
| SKU | No | Código interno. Si lo dejas vacío, se genera automáticamente (`P-` seguido de caracteres). Debe ser único |
| **Unidad de medida** | **Sí** (por defecto `Unidad`) | Unidad en la que se compra y vende |
| Costo | No (0 por defecto) | Costo estándar en **dólares**. Alimenta el costo de las ventas por pedido de cliente |
| Precio de venta | No (0 por defecto) | Precio de venta en **soles** |

Unidades disponibles: `Unidad`, `NIU` (unidad), `KGM` (kilogramo), `LTR` (litro), `MTR` (metro),
`BX` (caja), `PK` (paquete), `SET` (juego), `ZZ` (servicio).

### Qué puedes hacer

- **Crear**, **Editar**, **Eliminar**.
- **Activar / Desactivar** un producto (botón en la fila). Un producto desactivado no aparece en los
  selectores de compra ni de venta, pero su historial de inventario se conserva.
- **Buscar** por SKU o descripción.

> Si creas un producto con una descripción que **ya existe** en una compra, se reutiliza
> el existente y se actualiza su costo en dólares al valor de esa línea.

---

## 8. Compras: pedidos al proveedor exterior

**Para qué:** registrar lo que le compras a un proveedor del exterior, controlarlo mientras llega,
y convertirlo en inventario cuando lo recibes.

**Dónde:** `Compras` (menú lateral). Botón **Nueva compra** en la cabecera.

### 8.1 Qué se guarda de una compra

| Campo | Origen | Qué representa |
|-------|--------|----------------|
| Número | Automático | `PO-AÑO-#####`. Editable después (debe ser único) |
| Tipo | Derivado | **Cliente a pedido** si tiene cliente asociado; **General (stock)** si no |
| Proveedor | **Obligatorio** | Debe estar **activo** |
| Cliente | Opcional | Si lo pones, la compra queda asociada a un cliente para un pedido de cliente |
| Venta de origen | Automático | Se llena solo cuando la compra nació de una venta del tipo "Pedido de Cliente" |
| Forma de pago | **Obligatorio** | `Tarjeta de crédito`, `Efectivo` o `Billetera digital` |
| Tarjeta de crédito | **Obligatorio** si la forma de pago es tarjeta | Si no tienes tarjetas, el formulario te avisa para que las crees en Tesorería |
| Fecha de pedido | **Obligatorio** (hoy por defecto) | Fecha en que se hizo el pedido |
| Fecha estimada | Opcional (hoy por defecto) | Fecha prevista de llegada |
| Tipo de cambio (USD→PEN) | **Obligatorio** (≥ 0.01) | Se autocompleta con el tipo de cambio del día; puedes cambiarlo |
| Unidad de medida | Fijo | Las compras se registran siempre en `Unidad` |
| Notas | Opcional | Observaciones libres |
| Ítems | **Al menos 1** | Ver abajo |
| Costo base (USD) | Calculado | Suma de los ítems |
| Costos extras (USD) | Calculado | Suma de los costos extras ([§8.9](#89-costos-extras)). Empieza en 0 |
| Costo total (USD) | Calculado | **Costo base + Costos extras** |
| Costo real (PEN) | Calculado | Ver [§17](#17-cálculos-que-hace-el-sistema-por-ti) |
| Reintegro (USD) | Calculado | Solo aparece si se generó saldo a favor |

### 8.2 Los ítems

Cada ítem de la compra es una tarjeta. Puedes agregar y quitar líneas libremente.

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Producto** | No, pero o el producto **o** la descripción son obligatorios | Si eliges un producto, el sistema rellena descripción, costo y precio de venta |
| **Descripción** | Condicional | Obligatoria si no elegiste producto. Si no hay producto, **se crea uno nuevo** con esta descripción |
| **Cantidad** | **Sí** (entero, mayor a 0) | Unidades pedidas |
| **Costo** | **Sí** (USD, mayor a 0) | Costo unitario en dólares |
| Precio de venta | No (≥ 0, PEN) | Precio de venta unitario en soles que te sugiere el sistema |

El formulario muestra en vivo el **Costo total (USD)** de la compra.

### 8.3 Crear la compra: el asistente de 3 pasos

**Paso 1 — Proveedor, pago y datos generales.** Como en la tabla de arriba. Incluye los selectores
de proveedor, cliente y tarjeta, cada uno con la opción de **crear el registro en línea** sin salir
del formulario.

- Si el tipo de cambio se obtuvo del valor de respaldo (no hay conexión), aparece una etiqueta
  **"Modo contingencia"** junto al campo, para que sepas que el número no vino del mercado.
- Si la forma de pago no es tarjeta, el campo de tarjeta se oculta y se limpia.

**Paso 2 — Productos, cantidades y precios.** Como en 8.2.

**Paso 3 — Notas y confirmación.** Solo el campo opcional de notas, y un resumen de lo capturado.

Botones: **Atrás**, **Continuar**, **Guardar**.

### 8.4 Límite de compra (aviso)

Al guardar, si el **Costo total (USD)** supera el **Tope de compra** definido en Ajustes
(200 USD por defecto), aparece una advertencia: *"La compra supera S/ X"*. Puedes **Crear de todas
formas** o **Cancelar**. Es un aviso, no un bloqueo.

### 8.5 Qué pasa al crear la compra

- Se genera el número y el estado inicial es **Pendiente**.
- Si la forma de pago es **tarjeta de crédito**, el **costo total en USD se carga inmediatamente al
  saldo de esa tarjeta**. Si el saldo + el monto supera el límite de crédito de la tarjeta, la
  operación se rechaza.
- Cada ítem sin producto crea un producto nuevo con esa descripción.

### 8.6 Estados de la compra

```
Pendiente ──(Marcar como recibido)──▶ Recibida
    │                                     │
    └──────(Anular)───────────────────────┴──────▶ Anulada
```

| Estado | Qué significa | Qué puedes hacer |
|--------|---------------|------------------|
| **Pendiente** | El pedido está en curso. **No hay mercadería en inventario todavía** | Marcar como recibido, anular, agregar costos extras, editar el número |
| **Recibida** | La mercadería llegó y ya está en inventario como lotes | Anular (retira el stock), agregar costos extras |
| **Anulada** | La compra se canceló | No se puede editar nada. Los costos extras quedan bloqueados |

**No existe la eliminación de compras.** Si algo salió mal, se anula.

### 8.7 Marcar como recibido

Acción disponible en la fila de la compra cuando está **Pendiente** y todavía no tiene fecha de
recepción.

| Campo | Obligatorio | Notas |
|-------|-------------|-------|
| **Fecha de llegada** | **Sí** | No puede ser posterior a hoy |

Al confirmar:

1. La compra pasa a **Recibida** y se guarda la fecha de recepción.
2. Se marca **todo** el pedido como recibido. **No hay recepción parcial.**
3. **Cada línea de la compra se convierte en un lote de inventario** (ver [§9](#9-inventario)) con
   fecha de llegada igual a la fecha de recepción.
4. El costo unitario del lote en soles es el costo en dólares de la línea multiplicado por el tipo
   de cambio de la compra.

### 8.8 Anular la compra

Acción disponible en cualquier compra que no esté anulada. **El motivo es obligatorio**; hay cuatro
motivos sugeridos: *Mal estado*, *Error en el ingreso*, *Pedido duplicado*, *Cancelado por el
proveedor* (también puedes escribir tu propio texto).

| Qué anulas | Efecto |
|------------|--------|
| Compra **Pendiente** | Se libera el cargo de la tarjeta. Si el ciclo de esa tarjeta ya estaba liquidado, la compra queda con **Reintegro (USD)** = costo total |
| Compra **Recibida** | Primero se anulan todos sus lotes de inventario (la mercadería sale del stock), luego se libera la tarjeta con la misma regla de reintegro |

**Motivo "Mal estado":** además de anularse, la compra queda marcada como **Defectuoso** y guarda el
motivo. Aparece una etiqueta roja en la fila y el motivo en el detalle. Una compra defectuosa ya no
puede marcarse como recibida.

### 8.9 Costos extras

Adentro del detalle de la compra hay una sección **Costos extras** — *"Gastos adicionales de la
compra: fletes, aranceles, manejo"*.

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Concepto** | **Sí** | Nombre del gasto. Puedes escribirlo o elegir uno de los conceptos ya usados (autocompletado) |
| **Monto** | **Sí** (mayor a 0) | Importe. Puedes alternar el símbolo entre **USD** y **PEN** |
| **Tipo de cambio (USD→PEN)** | **Sí** (≥ 0.01) | Se autocompleta con el tipo de cambio vigente al guardar |

Acciones: **Agregar costo extra**, **Editar**, **Eliminar** (con confirmación). Cada concepto que
guardas queda en una lista reutilizable (se conservan los **20 más recientes**).

> **Los costos extras suman al costo de la compra.** Cada vez que agregas, editas o eliminas uno, el
> sistema recalcula tres cosas: los **costos extras (USD)**, el **costo total (USD)** y el **costo real
> (PEN)**. Siempre se cumple:
>
> ```
> Costo total = Costo base + Costos extras
> ```
>
> Un monto ya en **USD** se suma tal cual; uno en **PEN** se divide por el tipo de cambio guardado en
> esa misma fila para convertirlo a dólares.
>
> Como el costo total alimenta el costo real, los extras también terminan en el **costo unitario del
> lote** y por lo tanto en el **costo y la utilidad de cada venta** que use esa mercadería.
>
> Lo que **no** cambia es el **cargo de la tarjeta**: ese se hizo al crear la compra y sigue siendo solo
> el costo base. Agregar un extra no vuelve a cargar la tarjeta.
>
> En una compra **Anulada** la sección queda deshabilitada: no se puede agregar, editar ni eliminar.

### 8.10 La lista de compras

**Indicadores:** *Compras* (cuántas se listan), *Costo base (USD)*, *Costos extras (USD)*,
*Costo total (USD)*, *Costo real (PEN)*, *Por Pagar* (pendientes o recibidas) y *Anuladas*.

**Columnas:** Fecha de compra · Número · Proveedor · Productos · Costo (PEN | USD) · Estado · acciones.

- **Productos** muestra solo la **primera línea** de la compra, con su cantidad (p. ej. *"3× Mouse"*).
  El nombre se recorta si es muy largo y la cantidad nunca se oculta. El resto de las líneas las ves en
  el detalle.
- **Costo** junta las dos monedas en una sola celda: `Costo: S/ 4.312,50 | $ 1.150,00`, con el costo
  real en soles a la izquierda y el **costo total en dólares** a la derecha.

Cuando **todas** las líneas de una compra ya se vendieron por completo, la fila queda resaltada en
verde y aparece la etiqueta **Vendido** en Estado.

La columna **Estado** puede mostrar varias etiquetas a la vez:
`Pendiente` / `Recibida` / `Anulada`, más `Cliente` (si tiene cliente), `Defectuoso` (si llegó en mal
estado) y `Vendido` (cuando todas sus líneas ya se vendieron completas).

**Buscar:** por número de compra o por las notas.
**Filtrar por estado:** Todos · Pendientes · Recibidas · Anuladas.
**Filtros avanzados** (botón `Filtros`): rango de fechas (desde / hasta) y tarjeta de crédito. El
botón muestra cuántas hay activas.

**Acciones por fila:**

| Acción | Cuándo aparece | Qué hace |
|--------|----------------|----------|
| **Ver detalle** | Siempre | Abre el panel de detalle |
| **Editar número** | Siempre | Cambia el número de la compra (debe ser único) |
| **Marcar como recibido** | Pendiente, sin fecha de recepción y no defectuosa | Abre el diálogo de fecha de llegada |
| **Anular** | Cuando no está anulada | Abre el diálogo de anulación con motivo |

### 8.11 El detalle de una compra

Un panel lateral con el resumen: tipo, **cliente**, proveedor, **venta de origen**, forma de pago,
fecha de compra, fecha estimada, fecha de recepción, tipo de cambio, **costo base (USD)**, **costos
extras (USD)**, **costo total (USD)**, **costo real (PEN)** y reintegro (si existe). Debajo, las notas,
el motivo si fue defectuosa o anulada, y el detalle de cada línea con: descripción, cantidad × costo,
precio de venta, cuántas unidades ya se vendieron sobre el total pedido, y el importe de la línea.

Una línea cuyas unidades ya se vendieron todas aparece **tachada**, con la etiqueta **Agotado** al
lado del importe.

Al final, la sección de costos extras.

---

## 9. Inventario

**Para qué:** ver qué mercadería tienes, de qué lote viene, a qué costo está y cuánto le queda de
vida para venderse.

**Dónde:** `Inventario` (menú lateral).

### 9.1 Cómo nacen los lotes

Un **lote** es un grupo de mercadería con una fecha de ingreso y un costo. Se crea de dos maneras:

| Origen | Cuándo | Costo unitario del lote |
|--------|--------|-------------------------|
| **Recepción de una compra** | Al marcar una compra como recibida | Costo en dólares de la línea × tipo de cambio de la compra |
| **Ingreso manual** | Cuando registras stock que no viene de una compra (muestras, mercadería devuelta, ajustes de inicio) | El costo unitario que tú escribas, en soles |

Cada línea de una compra recibida produce **un lote**. Si la misma mercadería llega en dos compras
distintas, son **dos lotes distintos**, y puedes verlos y venderlos por separado.

### 9.2 La lista de lotes

**Indicadores:**

| Indicador | Qué cuenta |
|-----------|-----------|
| **Lotes activos** | Cuántos lotes hay (sin los anulados) |
| **Unidades en stock** | Suma de todas las cantidades |
| **Valor de inventario** | Suma de (cantidad × costo unitario) |
| **En remate** | Lotes que ya pasaron su fecha límite de venta |
| **Por vencer (5 días)** | Lotes a los que les quedan entre 0 y 4 días |

**Columnas:** Fecha de ingreso · SKU · Producto · Compra (número de compra de origen) · Cantidad ·
Costo unitario · Costo total · **Venta máxima** · **Días restantes** · Estado · acciones.

La columna **Estado** muestra: `Normal`, `REMATE` (cuando el lote ya venció su fecha de venta),
`Agotado` (sin existencias) o `Anulado`.

Los lotes **agotados** aparecen atenuados en la lista. Cuando un lote entra en remate, la columna
**Días restantes** muestra su valor **en rojo** (hasta el 0, que es el último día a precio normal, y
también los días ya vencidos).

**Buscar:** por producto o SKU.
**Filtrar:** Todos los lotes · En remate · Por vencer (5 días) · Anulados.

El ordenamiento por defecto es por fecha de ingreso, del más reciente al más antiguo.

### 9.3 Acciones sobre un lote

| Acción | Qué hace |
|--------|----------|
| **Ver movimientos** | Abre el historial completo de entradas y salidas de ese producto (ver 9.6) |
| **Recibir** | Abre el ingreso de stock con el producto y el costo del lote ya cargados, para sumar más unidades del mismo artículo |
| **Ajustar stock** | Fija la existencia del lote a un número exacto (ver 9.5) |
| **Anular lote** | Deja el lote en cero y lo marca como anulado, con el motivo fijo *"Anulado por error en el ingreso"* |

Un lote **Anulado** no muestra acciones.

### 9.4 Ingreso manual de stock

Botón **Nuevo ingreso** en la cabecera de Inventario.

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Producto** | **Sí** | Puedes crearlo en línea si no existe |
| **Fecha de ingreso** | **Sí** (hoy por defecto) | No puede ser posterior a hoy. Esta fecha determina cuándo empieza a contar la antigüedad del lote |
| **Cantidad** | **Sí** (entero, mayor a 0) | Unidades que ingresan |
| Costo unitario | No (≥ 0, en soles) | Costo por unidad en soles |

Un ingreso manual **no** crea una compra ni consume tarjeta.

### 9.5 Ajuste de existencias

Botón **Ajustar stock** sobre un lote.

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Nueva existencia** | **Sí** (entero, mayor a 0) | La cantidad **final** que quieres que tenga el lote, no la diferencia |
| **Motivo** | **Sí** (máx. 200 caracteres) | Razón del ajuste |

- Si la nueva cantidad es **mayor** que la actual, el sistema registra un ingreso por la diferencia.
- Si es **menor**, registra una salida por la diferencia.
- Si es **igual**, no ocurre nada (no se registra movimiento).

### 9.6 Movimientos (kardex)

Botón **Ver movimientos** sobre cualquier lote. Muestra el historial del **producto**, no solo del
lote: fecha, tipo de movimiento, cantidad con signo (`+` o `−`) y el saldo después del movimiento.

| Tipo de movimiento | Cuándo ocurre |
|--------------------|---------------|
| Ingreso por compra | Cuando una compra se marca como recibida |
| Venta | Cuando se vende mercadería de ese producto |
| Reversión de venta | Cuando una venta se anula y la mercadería vuelve al lote |
| Anulación de compra | Cuando una compra recibida se anula |
| Ajuste (entrada) | Cuando aumentas la existencia de un lote |
| Ajuste (salida) | Cuando reduces la existencia de un lote |

**El historial no se borra ni se edita nunca.** Toda corrección se registra como un movimiento
contrario, de modo que siempre puedes reconstruir qué pasó.

### 9.7 Reglas de remate (antigüedad)

Botón **Reglas** en la cabecera de Inventario.

| Campo | Obligatorio | Qué representa | Por defecto |
|-------|-------------|----------------|-------------|
| **Días para remate** | **Sí** (entre 1 y 365) | Cuántos días desde el ingreso tiene un lote para ser vendido | 25 |
| **Días de aviso previo** | **Sí** (entre 0 y 90) | Ventana de aviso previo al remate | 3 |

Cómo funciona:

- **Fecha de venta máxima = fecha de ingreso + días para remate.**
- El **día de la fecha máxima ya cuenta como remate**: el lote queda marcado desde ese día.
- Un lote en remate **se puede vender con normalidad**; la marca solo es informativa (aparece la
  etiqueta `REMATE` en inventario, y el lote entra en el panel *Productos en Remate*).
- El sistema revisa los lotes **cada minuto**, así que el cambio de estado aparece solo, sin que
  tengas que hacer nada.
- Si cambias los días para remate, el recálculo es inmediato.

> El filtro y el indicador **"Por vencer (5 días)"** de la lista de inventario usan una ventana
> fija de 5 días. El ajuste "Días de aviso previo" se guarda, pero no modifica esa lista.

---

## 10. Ventas

**Para qué:** registrar lo que vendes, ya sea mercadería que tienes en stock o mercadería que
todavía no ha llegado, y llevar el control de la cobranza.

**Dónde:** `Ventas` (menú lateral). Botón **Nueva venta**.

### 10.1 Los dos tipos de venta

| Tipo | Qué significa | Qué exige |
|------|---------------|-----------|
| **General (stock)** | Vendes mercadería que ya tienes en inventario | Que el producto tenga existencias disponibles |
| **Pedido de Cliente** | Vendes mercadería que todavía no has comprado | Nada: el sistema crea la compra por ti |

Si eliges **General (stock)** y algún producto no tiene stock, la venta **no se guarda** y el
formulario te avisa con la lista de productos sin existencias y un botón para cambiar a
**Pedido de Cliente**.

### 10.2 Qué se guarda de una venta

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| Número | Automático | `V-AÑO-#####` |
| **Cliente** | **Sí** | Puedes crearlo en línea |
| **Fecha de venta** | **Sí** (hoy por defecto) | Fecha de la operación |
| **Tipo de venta** | **Sí** | General (stock) o Pedido de Cliente |
| **Condición de pago** | **Sí** | Contado o Crédito |
| **Método de pago** | **Sí** | `Efectivo`, `Transferencia` u `Otro` |
| Fecha de vencimiento | **Sí**, solo si es **Crédito** | No puede ser anterior a la fecha de venta |
| Pago inicial | No, solo si es **Crédito** | No puede superar el total de la venta |
| Notas | No | Observaciones |
| Ítems | **Al menos 1** | Ver abajo |
| Moneda | Fijo | La venta se registra siempre en soles |

### 10.3 Los ítems de venta

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Producto** | **Sí** | Al elegirlo se autocompleta el precio de venta. Puedes crearlo en línea |
| **Cantidad** | **Sí** (entero, mayor a 0) | Unidades a vender |
| **Precio unitario** | **Sí** (mayor a 0) | Precio en soles. Se precarga con el precio del producto |
| Lote | Opcional | Solo aparece si el producto tiene **más de un lote** y la venta es de stock. Si lo dejas vacío, se usa **el lote más antiguo (FIFO)** |

Al elegir el lote verás su fecha de ingreso, cantidad, costo, si está en remate y la compra
de la que vino.

Cada línea muestra su **total** y el formulario muestra el **Total** de la venta.

### 10.4 Cobranza dentro de la venta

**Contado** (por defecto): el sistema registra **automáticamente un cobro por el total completo** de
la venta, con el método de pago elegido y con la fecha de la venta. No puedes registrar una venta
al contado con cobro parcial.

**Crédito**: la venta queda con saldo pendiente. Puedes registrar un **pago inicial** (opcional,
cero o menos, nunca mayor al total), y luego cobrar el resto desde la lista de ventas.

### 10.5 El asistente de 3 pasos

**Paso 1 — Identifica al comprador y la fecha.** Cliente, fecha de venta.
**Paso 2 — Agrega los productos.** Tipo de venta, líneas de productos, totales.
**Paso 3 — Condición de pago y confirmación.** Contado o crédito, método de pago, vencimiento, pago
inicial, notas.

Botones: **Atrás**, **Continuar**, **Guardar**.

### 10.6 Qué pasa al crear una venta

El sistema, en una sola operación:

1. Calcula el **total**, el **costo total** y la **utilidad** de la venta.
2. **Si es venta de stock:** descuenta las unidades de los lotes (el más antiguo primero, salvo que
   elijas uno), guarda en cada línea el costo real de la mercadería y registra los movimientos de
   salida.
3. **Si es Pedido de Cliente:** crea automáticamente una **compra** asociada a tu cliente,
   en estado **Pendiente**, sin proveedor y sin tarjeta. No mueve inventario.
4. **Aumenta la deuda del cliente** por el total de la venta.
5. Si hubo pago inicial, registra el cobro y **reduce la deuda** por ese monto.

Como resultado: una venta al contado no deja deuda; una venta al crédito sin pagos deja deuda por el
total.

### 10.7 Estados de la venta

```
Pendiente ──(cobro parcial)──▶ Parcial ──(cobro total)──▶ Cobrada
     │                             │                          │
     └────────────(Anular)─────────┴──────────────────────────┴──▶ Anulada
```

| Estado | Qué significa |
|--------|---------------|
| **Pendiente** | Registrada sin ningún cobro |
| **Parcial** | Tiene cobros, pero todavía debe saldo |
| **Cobrada** | Está totalmente pagada |
| **Anulada** | Se canceló por completo |

### 10.8 Cobrar una venta

Acción **Cobrar** disponible en las ventas **Pendiente** y **Parcial**.

| Campo | Obligatorio | Notas |
|-------|-------------|-------|
| **Fecha de pago** | **Sí** (hoy por defecto) | |
| **Método de pago** | **Sí** | `Efectivo`, `Transferencia` u `Otro` |
| **Monto a cobrar** | **Sí** (mayor a 0) | Precargado con el saldo pendiente. **No puede superar el saldo pendiente** |
| Referencia | No | N.º de operación, voucher, etc. |
| Notas | No | |

**Reglas:** no se puede cobrar una venta totalmente pagada ni una venta anulada. Un monto mayor al
saldo pendiente se **rechaza** (no se ajusta automáticamente).

Después del cobro, la venta pasa a **Parcial** o **Cobrada** según corresponda.

### 10.9 Historial de cobros

Acción **Cobros** en la fila de la venta. Muestra el total, lo pagado, el saldo pendiente y la fecha
de vencimiento, y luego cada cobro con su número, método, fecha, referencia y monto.

### 10.10 Anular una venta

Acción **Anular** disponible en cualquier venta que no esté anulada. **El motivo es obligatorio.**

| Efecto | Qué pasa |
|--------|----------|
| Inventario | Si era una venta de stock, la mercadería **vuelve a los lotes** de donde salió |
| Deuda del cliente | Se reduce en el monto que estaba pendiente (lo ya cobrado no se devuelve) |
| Cobros | **No se anulan.** Los cobros registrados permanecen en el historial |
| Pedido de cliente | Si tenía una compra asociada, puedes anularla desde Compras |

Una venta anulada ya no admite cobros.

### 10.11 La lista de ventas

**Indicadores:** *Ventas registradas*, *Monto total*, *Utilidad* (suma de las utilidades),
*Por Cobrar* (ventas pendientes o parciales).

**Columnas:** Fecha de venta · Número · Cliente · Estado · Total · Utilidad (en rojo si es negativa)
· acciones.

**Buscar:** por número o por cliente.
**Filtrar por estado:** Todos · Pendientes · Parciales · Pagadas · Anuladas.

**Acciones por fila:** *Ver detalle*, *Cobros*, *Cobrar* (solo si está pendiente o parcial),
*Envío* (crea un envío ya vinculado a esa venta), *Anular* (si no está anulada).

### 10.12 El detalle de una venta

Un panel lateral con fecha, vencimiento, estado, total, pagado, costo y utilidad. Si la venta nació de
un pedido de cliente, aparece un acceso directo a la **compra** relacionada. Debajo, las
notas, el motivo si fue anulada y el detalle de cada línea: descripción, cantidad × precio unitario,
de qué lote salió (con su fecha, o la indicación `FIFO` si se consumieron varios lotes), el **origen**
de esa mercadería (número de compra, con enlace directo; o `Sin origen` si fue un ingreso
manual) y el importe de la línea.

---

## 11. Tesorería: tarjetas de crédito y ciclos

**Para qué:** controlar con qué tarjeta se pagan las compras, cuánto se debe y cuándo vence cada
ciclo de facturación.

**Dónde:** `Tesorería` (menú lateral). Botón **Nueva tarjeta**.

### 11.1 Tarjetas de crédito

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Banco / Entidad** | **Sí** | Emisor de la tarjeta |
| **Últimos 4 dígitos** | **Sí** (exactamente 4 dígitos) | Identificación de la tarjeta |
| **Límite de crédito** | **Sí** (USD, mayor a 0) | Tope disponible de la tarjeta |
| **Día de corte (1-31)** | **Sí** | Día en que cierra el ciclo de la tarjeta |
| **Día de pago (1-31)** | **Sí** | Día en que hay que pagar a la tarjeta |

Cada tarjeta se muestra como `Banco •••• 1234`.

**El saldo de la tarjeta se actualiza solo:**

| Situación | Efecto en el saldo |
|-----------|--------------------|
| Creas una compra pagando con tarjeta | Se suma el costo total en USD |
| Anulas una compra con tarjeta | Se resta el costo total en USD (nunca baja de cero) |
| Registras un pago a la tarjeta | Se resta el monto pagado (si pagas de más, el saldo queda en cero) |

**Una compra que haga exceder el límite de crédito de la tarjeta se rechaza.** Si necesitas pagar
con efectivo o billetera digital, usa esa forma de pago.

**Acciones por tarjeta:** *Registrar pago* (deshabilitada si el saldo ya está en cero), *Editar* y
*Eliminar*. Al eliminar una tarjeta queda desactivada y sus compras asociadas se conservan.

### 11.2 Registrar el pago de una tarjeta

| Campo | Obligatorio | Notas |
|-------|-------------|-------|
| **Monto del pago** | **Sí** (USD, mayor a 0) | Precargado con el saldo actual de la tarjeta |

El panel superior muestra el **saldo actual del ciclo**.

### 11.3 Ciclos de facturación

Cada tarjeta activa genera una proyección de ciclo, calculada automáticamente a partir de su día de
corte y su día de pago.

| Dato | Qué representa |
|------|----------------|
| **Ciclo** | Período que va del día siguiente al corte anterior hasta el día de corte |
| **Total (USD)** | Suma de las compras cargadas a esa tarjeta dentro del ciclo, **excluyendo** las anuladas |
| **Reintegros (USD)** | Suma de los reintegros generados por compras anuladas de esa tarjeta en el ciclo |
| **Fecha de pago** | El día de pago de la tarjeta, en el mes siguiente al corte |
| **Estado** | `Pendiente` mientras el ciclo está abierto; `Liquidado` una vez pasada la fecha de corte |

La tabla se puede filtrar por estado del ciclo y buscarse por tarjeta.

> El ciclo se calcula con el **día de corte** de la tarjeta. Si el corte es el 31, en meses cortos
> se ajusta al último día del mes.

### 11.4 Tipo de cambio

El sistema mantiene el tipo de cambio **dólares → soles** del día para que los montos en dólares
puedan convertirse a soles.

**De dónde sale el tipo de cambio:**

1. Primero intenta obtenerlo de fuentes públicas del mercado (el tipo de cambio de venta del día).
2. Si no hay conexión, usa el **TC de respaldo** de Ajustes (3,75 por defecto) y marca los
   formularios con la etiqueta **"Modo contingencia"** para avisarte.

**Dónde lo usas:**

| Lugar | Para qué |
|-------|----------|
| Formulario de compra | Se precarga el tipo de cambio; **puedes editarlo** antes de guardar |
| Formulario de costo extra | Se autocompleta al guardar el costo |
| Venta por pedido de cliente | Se usa para valorar la mercadería que aún no se ha comprado |

Guardar la compra **congela** el tipo de cambio: queda registrado y no cambia aunque el
mercado se mueva después.

---

## 12. Envíos

**Para qué:** dar seguimiento a las salidas de mercadería con un código que puedas compartir y una
**clave de seguridad** que verifica la entrega.

**Dónde:** `Envíos` (menú lateral). Botón **Nuevo envío**. También puedes crear un envío desde el menú
de acciones de cualquier venta (`Envío`), que lo abre con esa venta ya asignada.

> Esta área es de **seguimiento**, no de logística: no registra transportista ni número de contenedor.

### Qué se guarda de un envío

| Campo | Obligatorio | Qué representa |
|-------|-------------|----------------|
| **Venta** | **Sí** | La venta que origina el envío. **Fija**: no se puede cambiar después |
| **Clave de seguridad** | **Sí** (4 dígitos) | Código que **muestra el cliente al recibir**. Se genera solo, se puede editar antes de guardar y después queda **fija** |
| **Código** | Automático | `ENV-AÑO-#####` (p. ej. `ENV-2026-00001`). No se puede cambiar ni reutilizar |
| **Descripción** | **Sí** | Descripción corta del envío (p. ej. *"Caja de mercadería"*) |
| **Ubicación de envío** | No | Lugar de salida (p. ej. *"Oficina de Lima"*) |
| **Fecha de envío** | No | Cuándo salió la mercadería |
| **Fecha de entrega** | No | Cuándo la recibió el cliente. No puede ser anterior a la fecha de envío |
| Notas | No | Instrucciones de entrega |

El **cliente se deduce de la venta**, así que siempre coincide con el comprador real.

El **código** se genera solo y se muestra en la lista con un botón para **copiarlo al portapapeles**. La
**clave de seguridad** también tiene su propio botón de copiar: es lo que te pide el repartidor para
confirmar que estás recibiendo la mercadería correcta. Al marcar el envío como **Enviado** o
**Entregado** se autocompleta la fecha correspondiente si la dejaste vacía.

### Estados

```
Pendiente ──▶ Enviado ──▶ Entregado
```

Al abrir un envío en modo edición aparece su estado actual y un botón **"Marcar como …"** con el
siguiente paso. Cada confirmación muestra una notificación con el nuevo estado.

| Estado | Qué significa |
|--------|---------------|
| **Pendiente** | Registrado, todavía no salió |
| **Enviado** | En camino |
| **Entregado** | Recibido por el cliente |

### La lista de envíos

**Columnas:** Fecha · Código (con botón de copiar) · **Clave** (con botón de copiar) · Venta (número de
la venta) · Cliente · **Ubicación** · **Envío** (fecha) · **Entrega** (fecha) · Descripción · Estado ·
acciones.
**Buscar:** por código, por clave de seguridad o por descripción. **Filtrar:** Todos · Pendiente ·
Enviado · Entregado.
**Acciones:** *Editar* y *Eliminar* (con confirmación).

Todo envío tiene **venta y clave de seguridad**: son las dos cosas que lo hacen verificable.

---

## 13. Ajustes

**Dónde:** `Ajustes` (menú lateral). Todos los formularios tienen un botón **Guardar**.

### 13.1 Parámetros de negocio

| Parámetro | Obligatorio | Qué controla | Por defecto |
|-----------|-------------|--------------|-------------|
| **Días para remate** | **Sí** (1 a 365) | Cuántos días desde el ingreso tiene un lote antes de entrar en remate | 25 |
| **Costo de importación USD** | **Sí** (0 a 100) | Factor que se suma al costo en dólares de la compra para calcular el costo real en soles. Ej.: `0.07` = 7% | 0.07 |
| **TC de respaldo** | **Sí** (0.01 a 100) | Tipo de cambio a usar cuando no hay conexión para consultar el del día | 3.75 |
| **Tope de compra** | **Sí** (0 a 1.000.000 USD) | Monto en dólares a partir del cual aparece el aviso al crear una compra | 200 |

> Los *días para remate* también se editan desde `Inventario → Reglas`.

### 13.2 Datos de la empresa

| Campo | Obligatorio | Notas |
|-------|-------------|-------|
| **Razón social** | **Sí** (mínimo 2 caracteres) | |
| Nombre comercial | No | |
| RUC | No | 11 dígitos, empieza con 10 o 20 |
| Correo electrónico | No | Debe tener formato de correo |
| Dirección fiscal | No | |
| Teléfono | No | |
| Web | No | |

Son los mismos datos que pediste en el asistente inicial.

### 13.3 Seguridad

Botón **Administrar autenticación** que abre el panel descrito en [§4](#4-acceso-contraseña-y-seguridad).

---

## 14. Copias de seguridad y nube

### 14.1 Copia de seguridad

Cada copia es un archivo con toda tu información, en una carpeta que tú eliges.

| Elemento | Opciones |
|----------|----------|
| **Carpeta** | Botón **Elegir** que abre el explorador de carpetas del sistema. Por defecto: `~/.vfinancy/backups` |
| **Frecuencia automática** | `Desactivada` · `Al cerrar la aplicación` · `Diaria` · `Semanal` |
| **Crear copia ahora** | Botón para generar una copia inmediata |

- Cada copia se llama `backup_AAAAMMDD_HHMMSS`.
- Si la carpeta configurada no está disponible, el sistema **igual hace la copia** en la carpeta de
  respaldo por defecto y te avisa con el mensaje *"Resguardo de emergencia"* indicando dónde quedó.
- Antes de copiar, el sistema verifica que haya **al menos 64 MB libres** en el destino. Si no hay
  espacio, te avisa en lugar de crear un archivo incompleto.

### 14.2 Sincronización en la nube (opcional)

Puedes conectar la aplicación a un servidor propio para que varios equipos compartan la misma
información. **Es opcional y está apagado por defecto**; mientras no lo actives, todo funciona solo en
tu equipo.

| Campo | Por defecto | Qué es |
|-------|-------------|--------|
| Servidor (host) | vacío | Dirección del servidor |
| Puerto | 5432 | Puerto del servidor |
| Base de datos | vacío | Nombre de la base |
| Usuario | vacío | Usuario de acceso |
| Contraseña | vacío | Contraseña de acceso |
| Modo SSL | `Requerido` | `Desactivado` · `Requerido` · `Verificación completa` |
| Intervalo de sincronización | 30 segundos (mínimo 10) | Cada cuánto se sincroniza |
| Estado | Desactivado | Activado / Desactivado |

**Botones:**

| Botón | Qué hace |
|-------|----------|
| **Probar conexión** | Verifica que el servidor responda, sin guardar nada. Resultado: *"Conexión exitosa"* o el error |
| **Guardar** | Guarda la configuración y aplica el estado elegido |
| **Sincronizar ahora** | Fuerza una sincronización inmediata |

**Importante:**

- Para que la sincronización funcione, el servidor, la base de datos y el usuario deben estar
  completos y el estado debe estar **Activado**.
- La información **siempre está disponible en tu equipo**, incluso sin la nube. La nube es una copia
  adicional.
- Los equipos comparten clientes, proveedores, productos, compras, inventario, ventas, tarjetas y
  envíos. La contraseña y los datos de tu empresa son **solo tuyos** y no se comparten.
- La contraseña y los datos de la empresa **nunca salen de tu equipo**.

---

## 15. El ciclo completo de una importación

Este es el recorrido habitual, de principio a fin.

### Paso 0 — Configurar el negocio (una vez)

1. Completa el asistente inicial con los datos de tu empresa.
2. Opcionalmente define una contraseña.
3. En `Ajustes → Parámetros de negocio`, revisa los días de remate, el factor de importación, el
   TC de respaldo y el tope de compra.

### Paso 1 — Registrar al proveedor

`Compras → Proveedores → Nuevo proveedor`

- **Obligatorio:** Nombre.
- Ej.: *Shenzhen Yiwu Trading Co.*

### Paso 2 — Registrar tus productos (opcional)

`Inventario → Productos → Nuevo producto`

- Si ya no los tienes, **no hace falta**: al crear la compra puedes escribir directamente la
  descripción y el sistema creará el producto.
- **Obligatorio:** Descripción y Unidad de medida.

### Paso 3 — Crear la compra

`Compras → Nueva compra`

**Paso 1 del formulario** — Proveedor (obligatorio, debe estar activo), forma de pago (obligatoria),
tarjeta de crédito (obligatoria si paga con tarjeta), fecha de pedido (obligatoria), tipo de cambio
(obligatorio, se autocompleta), y opcionalmente el cliente si es una compra para un pedido.

**Paso 2 del formulario** — Los productos:

- Elige un producto del catálogo **o** escribe una descripción nueva.
- Cantidad (obligatoria) y costo en dólares (obligatorio).
- Precio de venta en soles (opcional, se autocompleta si elegiste producto).
- Agrega varias líneas con **Agregar ítem**.
- Revisa el **Costo total (USD)**.

**Paso 3 del formulario** — Notas (opcional) y **Guardar**.

**Lo que pasa al guardar:**

- Se genera el número `PO-AÑO-#####` y el estado **Pendiente**.
- Si pagaste con tarjeta, el costo total se **carga a esa tarjeta** (si excede el límite, se rechaza).
- Aparece un aviso si superaste el **tope de compra** (puedes confirmarlo o cancelar).

### Paso 4 — Registrar los costos reales del envío

Hazlo **antes** de marcar la compra como recibida: es el momento en que los extras todavía alcanzan al
costo unitario de los lotes.

Abre el **detalle** de la compra (clic en la fila) y ve a **Costos extras**:

- **Agregar costo extra** → concepto (obligatorio), monto (obligatorio) y tipo de cambio
  (se autocompleta).
- Repite por cada gasto: flete marítimo, IUI, arancel, manejo en aduana, agente de carga,
  certificado, etc.
- Los conceptos ya usados aparecen en el autocompletado, con su monto y moneda, para reutilizarlos.

> Cada extra que agregas actualiza al instante los **costos extras**, el **costo total** y el **costo
> real** de la compra, y eso se reparte entre las líneas para costear los lotes. Por eso conviene
> cargarlos antes de recibirlas.
>
> El **cargo de la tarjeta** no se recalcula: sigue siendo solo el costo base del momento en que
> creaste la compra.

### Paso 5 — Cuando llega la mercadería

En la fila de la compra → **Marcar como recibido**

- **Obligatorio:** Fecha de llegada (no puede ser futura).
- Al confirmar: la compra pasa a **Recibida** y **cada línea se convierte en un lote** en inventario,
  con la fecha de llegada y el costo en soles (costo en soles de la línea).

### Paso 6 — Vender la mercadería

`Ventas → Nueva venta`

**Paso 1** — Cliente (obligatorio), fecha de venta (obligatoria).
**Paso 2** — Tipo **General (stock)**, y las líneas: producto (obligatorio), cantidad (obligatoria) y
precio unitario (obligatorio). Si el producto tiene varios lotes puedes elegir uno o dejar que se use
el más antiguo.
**Paso 3** — Condición de pago:

- **Contado:** el sistema cobra el total automáticamente.
- **Crédito:** defines fecha de vencimiento y, si quieres, un pago inicial.

**Guardar.** El stock sale de los lotes y la utilidad de la venta queda calculada.

Si la venta es de un artículo que **aún no tienes**, elige **Pedido de Cliente** en el paso 2: el
sistema crea la compra por ti.

### Paso 7 — Cobrar

En la fila de la venta → **Cobrar**

- **Obligatorio:** fecha, método de pago y monto (no puede superar el saldo pendiente).
- La venta pasa a **Parcial** o **Cobrada**, y la deuda del cliente baja.

Para ver todo el historial: botón **Cobros**.

### Paso 8 — Entregar y dar seguimiento

`Envíos → Nuevo envío`

- Se genera el **código de 4 dígitos**.
- Elige cliente y venta (opcional), escribe la descripción y las notas.
- Usa **Marcar como Enviado** y luego **Marcar como Entregado**.

### Paso 9 — Cerrar el ciclo en Tesorería

`Tesorería`

- Revisa el **saldo actual** de cada tarjeta.
- Cuando la 보이는 el estado liquidado, registra el **pago** de la tarjeta.
- Revisa la **fecha de pago** de cada ciclo para no pagar tarde.

### Paso 10 — Resumen

`Inicio`

- Ventas cobradas del mes, utilidad del mes, cuántas ventas están cobradas / pendientes / anuladas.
- Gráfico de utilidad de los últimos 6 meses.
- Lotes que ya entraron en remate.

---

## 16. Estados y transiciones

| Entidad | Estados | Transiciones |
|---------|---------|--------------|
| **Compra** | Pendiente → Recibida · Pendiente → Anulada · Recibida → Anulada | Marcar como recibido (solo Pendiente) · Anular (cualquiera que no esté anulada) · Marca *Defectuoso* al anular por mal estado |
| **Venta** | Pendiente → Parcial → Cobrada · cualquiera → Anulada | Registrar cobro (Pendiente o Parcial) · Anular (cualquiera que no esté anulada) |
| **Lote de inventario** | Activo (Normal / REMATE) → Agotado → Anulado | Se marca Agotado solo al vender todo · se Anula con la acción explícita |
| **Envío** | Pendiente → Enviado → Entregado | Botón "Marcar como …" |
| **Ciclo de tarjeta** | Pendiente → Liquidado | Automático al pasar la fecha de corte |
| **Producto** | Activo / Inactivo | Activar / Desactivar |
| **Proveedor** | Activo / Inactivo | No hay un control para cambiarlo. Un proveedor inactivo no se puede usar en compras |

---

## 17. Cálculos que hace el sistema por ti

| Cálculo | Cómo se obtiene |
|---------|-----------------|
| **Costo base de la compra (USD)** | Suma de (cantidad × costo en dólares) de cada línea |
| **Costos extras de la compra (USD)** | Suma de los costos extras. Un monto en USD se suma tal cual; uno en PEN se divide por el tipo de cambio de esa fila |
| **Costo total de la compra (USD)** | **Costo base + Costos extras**. Se recalcula en cada alta, edición o baja de un costo extra |
| **Costo real de la compra (PEN)** | (Costo total USD + factor de importación) × tipo de cambio de la compra |
| **Costo unitario del lote (PEN)** | Costo real de la compra repartido entre sus líneas según la parte de cada una en el costo base, y dividido por su cantidad |
| **Total de la venta** | Suma de (cantidad × precio unitario) de cada línea |
| **Costo total de la venta** | Suma de (cantidad × costo del lote) para ventas de stock. Como el costo del lote ya incluye los costos extras de su compra, la utilidad los refleja. Para pedidos de cliente se usa el costo del producto en dólares convertido al tipo de cambio |
| **Utilidad de la venta** | Total − costo total. Si es negativa, se muestra en rojo |
| **Saldo pendiente de la venta** | Total − pagado. Nunca baja de cero |
| **Deuda del cliente** | Suma de todas sus ventas menos todos sus cobros. Nunca baja de cero |
| **Fecha de venta máxima del lote** | Fecha de ingreso + días para remate (25 por defecto). El día de esa fecha ya cuenta como remate |
| **Días restantes** | Días que faltan para la fecha de venta máxima |
| **Valor del inventario** | Suma de (cantidad × costo unitario) de todos los lotes no anulados |
| **Total de un ciclo de tarjeta** | Suma de las compras con tarjeta dentro del ciclo, excluyendo las anuladas |
| **Reintegro de un ciclo** | Suma de los reintegros de las compras anuladas de esa tarjeta dentro del ciclo |
| **Saldo de la tarjeta** | Costo total de las compras con tarjeta − pagos − anulaciones. Nunca baja de cero |
| **Ventas cobradas del mes (panel)** | Suma de los cobros del mes actual, excluyendo ventas anuladas |
| **Costo base / extras / total de un mes (panel)** | Suma de las compras no anuladas cuya **fecha de compra** cae en ese mes |
| **Precio de venta de un mes (panel)** | Suma de las ventas no anuladas cuya fecha de venta cae en ese mes |
| **Utilidad de un mes (panel)** | Precio de venta − (Costo base + Costos extras) de ese mismo mes. Los costos se convierten a soles con el tipo de cambio de cada compra |
| **Reintegro de una compra anulada** | Se genera cuando el ciclo de la tarjeta ya estaba liquidado al anular la compra |

---

## 18. Validaciones, bloqueos y dependencias

### Datos que siempre debes ingresar

| Entidad | Obligatorio |
|----------|-------------|
| **Compra** | Proveedor activo · al menos un ítem · forma de pago (tarjeta ⇒ tarjeta elegida) · fecha de pedido · tipo de cambio > 0 · por ítem: cantidad y costo > 0 (o producto o descripción) |
| **Venta** | Cliente · fecha · tipo de pago · al menos un producto · por línea: producto, cantidad y precio unitario > 0 · si es a crédito: fecha de vencimiento ≥ fecha de venta |
| **Lote (manual)** | Producto · fecha de ingreso (no futura) · cantidad entera > 0 |
| **Ajuste de lote** | Nueva existencia (entera > 0) · motivo |
| **Envío** | Venta · clave de seguridad de 4 dígitos · descripción |
| **Proveedor** | Nombre (único) |
| **Producto** | Descripción · unidad de medida |
| **Cliente** | Nombre o razón social |
| **Tarjeta de crédito** | Banco · 4 dígitos · límite · día de corte · día de pago |
| **Costo extra** | Concepto · monto > 0 · tipo de cambio (se suma al costo total de la compra) |
| **Anulación** | Motivo obligatorio (en el lote se usa un motivo fijo: *"Anulado por error en el ingreso"*) |
| **Empresa** | Razón social |

### Acciones bloqueadas

| No puedes… | Motivo |
|------------|--------|
| Eliminar un proveedor con compras | *"No se puede eliminar un proveedor con compras asociadas"* |
| Usar un proveedor inactivo en una compra nueva | *"El proveedor está inactivo"* |
| Usar una tarjeta que exceda su límite de crédito | El cargo no cabe en el saldo disponible |
| Vender por stock un producto sin existencias | *"Sin stock disponible"* — la venta no se guarda |
| Elegir un lote que no tiene existencias | *"El lote seleccionado no tiene stock disponible"* |
| Cobrar más que el saldo pendiente | *"El pago supera el saldo pendiente"* |
| Cobrar una venta totalmente pagada o anulada | *"La venta ya está cobrada"* |
| Pagar una tarjeta con saldo cero | El botón está deshabilitado |
| Una venta al contado con cobro parcial | *"Una venta al contado requiere el pago completo"* |
| Registrar recepción con fecha futura | *"La fecha de recepción no puede ser posterior a hoy"* |
| Marcar como recibido dos veces una compra | Solo se puede desde el estado Pendiente |
| Marcar como recibido una compra defectuosa | Las compras marcadas como *Mal estado* quedan anuladas |
| Agregar, editar o eliminar costos extras en una compra anulada | *"No se pueden modificar costos de una compra anulada"* |
| Anular dos veces la misma compra o venta | La compra o venta ya está anulada |
| Anular un lote ya anulado | El lote ya está anulado |
| Usar el mismo documento (DNI/RUC) en dos clientes | *"Ya existe un cliente con ese tipo y número de documento"* |
| Usar el mismo nombre de proveedor dos veces | *"Ya existe un proveedor con ese nombre"* |
| Usar el mismo SKU en dos productos | *"Ya existe un producto con ese SKU"* |
| Usar el mismo número en dos compras | *"Ya existe una compra con ese número"* |
| Desbloquear la aplicación con una contraseña incorrecta 5 veces seguidas | Se bloquea durante 15 minutos |

### Reglas de formato

| Campo | Formato |
|-------|---------|
| DNI | 8 dígitos |
| RUC | 11 dígitos, empezando con 10 o 20 |
| Correo | Formato de correo válido |
| Cantidades | Números enteros, mayores a 0 |
| Montos | 2 decimales, no negativos (salvo donde se indique) |
| Tipo de cambio | Mayor a 0.01 |

---

## 19. Cómo se conectan las áreas

```
        PROVEEDORES ──────────┐
             │                │
             ▼                ▼
   PRODUCTOS ──────▶ COMPRAS ──────┬──▶ COSTOS EXTRAS
             │            │        │      (suman al costo total)
             │            ▼        │
             │      (Marcar como recibido)
             │            │
             │            ▼
             │      INVENTARIO (lotes + movimientos)
             │            │
             │            │ (FIFO)
             │            ▼
   CLIENTES ────────────▶ VENTAS ──────┬──▶ COBROS ──▶ (baja la deuda del cliente)
             │            │            │
             │            │            └──▶ COMPRA (Pedido de Cliente)
             │            │
             │            └──▶ LÍNEA DE VENTA (guarda de qué lote salió y su costo)
             │
             └──▶ DEUDA ACUMULADA

   COMPRAS ──(si se pagan con tarjeta)──▶ TESORERÍA (saldo de la tarjeta + ciclo de facturación)
```

**Historias concretas de conexión:**

| Necesitas… | Áreas que se conectan |
|------------|----------------------|
| Saber por qué un lote está en remate | Compra → fecha de recepción → Inventario → regla de días de remate |
| Saber cuánto gané con una venta | Venta → lote de origen → compra → costo del lote |
| Ver toda la deuda de un cliente | Ventas del cliente − cobros registrados |
| Ver qué compra generó una venta por pedido de cliente | Venta → compra asociada (enlace directo) |
| Ver qué lote salió con una venta concreta | Detalle de venta → lote y fecha de ingreso → compra del lote |
| Saber cuánto tengo que pagar a la tarjeta | Tesorería → compras con esa tarjeta → total y reintegros del ciclo |
| Ver el historial de un cliente | Ventas del cliente (importe y deuda) + datos de contacto |

---

## 20. Límites conocidos

Cosas que la aplicación **no** hace, para que no pierdas tiempo buscándolas:

| Área | Lo que no existe |
|------|------------------|
| **Envíos** | No hay transportista, número de contenedor ni seguimiento automático. Las fechas de envío y de entrega sí se registran a mano |
| **Compras** | No hay recepción parcial ni pesaje por línea. Al marcar como recibido, entra todo el pedido |
| **Compras** | No hay conversión de moneda distinta del dólar; la compra se registra siempre en USD con un tipo de cambio USD→PEN |
| **Compras** | Las compras no se eliminan, solo se anulan |
| **Ventas** | No hay devoluciones ni notas de crédito. La única forma de revertir una venta es anularla (lo cobrado no se devuelve) |
| **Costos extras** | No vuelven a cargar la tarjeta: el cargo se fija al crear la compra y solo cubre el costo base. Si agregas un extra después de haber recibido la mercadería, las ventas ya registradas conservan el costo anterior |
| **Productos** | El costo en dólares se usa para pedidos de cliente y referencia; el inventario se valora con el costo real del lote |
| **Clientes** | No se pueden desactivar clientes desde la interfaz |
| **Clientes** | Se puede eliminar un cliente aunque tenga deuda o ventas registradas |
| **Reportes** | No hay reportes exportables ni impresión de comprobantes. La información se consulta en pantalla |
| **Contabilidad** | No hay libros contables, asientos ni declaración |
| **Multiempresa** | Una sola empresa, un solo usuario, sin roles ni permisos |

---

*vfinancy · ERP de importación · Interfaz en español (Perú)*