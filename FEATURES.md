# DOCUMENTO DE FUNCIONALIDADES TÉCNICAS

## 1. Arquitectura y Restricciones del Sistema

* **Stack Tecnológico:**
  * **Core / Backend Local:** Desplegado mediante **Wails** (lenguaje Go en segundo plano), garantizando compilación nativa de escritorio, bajo consumo de recursos y comunicación directa con el sistema operativo sin depender de un navegador externo.
  * **Frontend / Interfaz de Usuario:** Construido en **React**, ofreciendo una experiencia visual fluida, reactiva y modular para la interacción del usuario.

* **Persistencia Local-First:**
  * **Base de Datos:** Motor relacional **SQLite** incrustado localmente en el equipo. La aplicación opera de manera totalmente autónoma (*offline-first*), garantizando velocidad de lectura/escritura sin latencia de red ni caída de servicios en la nube.

* **Modelo de Acceso Unipersonal:**
  * **Cuenta Única Local:** Diseñado para la operación estricta de un solo usuario (dueño/administrador de la MYPE).
  * **Sin Jerarquía de Roles ni Permisos:** Se excluye por diseño cualquier arquitectura de control de acceso basada en roles (RBAC), perfiles múltiples o matrices de permisos, manteniendo el sistema liviano y sin complejidad administrativa.

---

## 2. Visión General del Sistema

* **Objetivo Principal:** Proveer un sistema ERP de escritorio autónomo y eficiente que consolide la gestión operativa bimonetaria (compras internacionales en USD y ventas locales en PEN) sin depender de una conexión constante a la nube.

* **Público Objetivo:** Micro y pequeñas empresas (MYPEs) peruanas (personas naturales con negocio con RUC 10 o empresas con RUC 20) administradas por un único usuario, enfocadas en la importación de productos bajo el régimen simplificado de aduanas y su comercialización en el mercado local.

* **Propuesta de Valor:** Una solución *Local-First* que mitiga la fricción contable de operaciones multimoneda, automatiza el cálculo del costo real (incluyendo factor logístico y tipo de cambio) y centraliza el control de inventarios, evitando el estancamiento de capital mediante alertas de obsolescencia (remate).

---

## 3. Módulos y Funcionalidades Técnicas

### [Gestión de Pedidos Bimonetaria y Control Aduanero]

* **Descripción:** Permite el registro de productos en el catálogo comercial, la creación de pedidos generales o de cliente, el registro de sus costos en dólares (USD), la agrupación de pedidos en lotes para control aduanero, la asociación a medios de pago de crédito internacional, el marcado de recepción de mercancía y el procesamiento de anulaciones por productos con fallo con reintegro total.

* **Opciones y Flujos Personalizables:**
  * Selección entre tipo de pedido: "General" (stock) o "Cliente" (a pedido).
  * Obtención del Tipo de Cambio mediante API en tiempo real o uso de valor de respaldo/fallback (configurable).
  * Agrupación flexible de pedidos en Lotes de Importación.
  * Ajuste personalizable del tope aduanero (predeterminado en $220 USD).
  * Ajuste personalizable del costo de importación (factor de ganancia/flete/aduanas, predeterminado en +0.07 USD por compra).

* **Campos Obligatorios:**
  * **Nombre / Descripción del Producto:** Obligatorio técnica y fiscalmente para la identificación del bien en el catálogo comercial y posterior emisión del comprobante de pago (Ley del Comprobante de Pago - SUNAT).
  * **Unidad de Medida:** Obligatorio por SUNAT (asignado automáticamente como "Unidad").
  * **Costo del Producto (USD):** Obligatorio para el registro exacto de la transacción de importación y el cálculo posterior del costo real y crédito de pasivos.
  * **Tipo de Pedido:** Obligatorio técnicamente para definir el comportamiento del flujo de venta (stock vs. a pedido).
  * **Tarjeta de Crédito de Pago:** Obligatoria para la asociación y trazabilidad en el módulo de Tesorería/Pasivos.
  * **Tipo de Cambio (USD/PEN):** Obligatorio por la SUNAT (Art. 5 del Reglamento del IGV / Código Tributario) para la conversión de transacciones en moneda extranjera a moneda nacional con fines tributarios.

* **Campos Opcionales y Personalizables:**
  * **Lote de Importación:** Opcional al crear el pedido individual, necesario únicamente al agrupar para el control del tope aduanero ($220 USD).
  * **Límite / Tope Aduanero Personalizado:** Parámetro de configuración del sistema editable por el usuario.
  * **Tipo de Cambio de Respaldo (Fallback):** Campo de configuración (predeterminado en 3.75 Soles).

---

### [Recepción de Mercadería y Alertas de Remate]

* **Descripción:** Controla el ingreso de productos al inventario al marcar una orden de compra como "Recibida", enrutando e incrementando automáticamente el stock en el almacén principal. Monitorea la antigüedad de los productos a partir de su fecha de ingreso y genera etiquetas visuales e indicadores de liquidación cuando se alcanza o supera el umbral de permanencia.

* **Opciones y Flujos Personalizables:**
  * Realización de ajustes directos de stock manuales en casos excepcionales (cantidad entera > 0).
  * Filtrado rápido en la vista de Stock mediante el botón "Ver productos en Remate".
  * Bloqueo en formulario de fechas de ingreso posteriores a la fecha actual del sistema (mitigación de desfase de reloj).
  * Umbral de días máximos de permanencia configurable por el usuario (predeterminado en 25 días calendario).

* **Campos Obligatorios:**
  * **Orden de Compra / Pedido Asociado:** Obligatorio para garantizar la trazabilidad del ingreso de mercancía al inventario (Kardex físico/valorizado).
  * **Almacén de Destino:** Preseleccionado automáticamente como "Almacén Principal".
  * **Fecha de Ingreso / Recepción:** Obligatoria técnicamente y para control fiscal para establecer la antigüedad de la mercancía e iniciar el cómputo de la regla de permanencia.
  * **Cantidad Ingresada:** Obligatoria para la actualización matemática del stock disponible en Kardex.

* **Campos Opcionales y Personalizables:**
  * **Días para Remate / Límite de Permanencia:** Campo personalizable en el módulo de Configuración General.

---

### [Ventas Locales (PEN)]

* **Descripción:** Gestiona el proceso de comercialización de productos a nivel local en soles (PEN), registrando ventas al contado o al crédito con control de fechas de emisión/vencimiento e historial de cobros.

* **Opciones y Flujos Personalizables:**
  * Selección de clientes existentes o creación de nuevos clientes durante el flujo de venta.
  * Modalidad de venta: Contado o Crédito con fecha de vencimiento configurada.

* **Campos Obligatorios:**
  * **Datos de Identificación del Cliente (DNI / RUC / Nombre o Razón Social):** Obligatorio según normativa de la SUNAT (Resolución de Superintendencia N.º 007-99/SUNAT) para la posterior emisión de Comprobantes de Pago Electrónicos (Boleta de Venta o Factura Electrónica).
  * **Moneda de Venta:** Obligatoria fiscalmente (definida exclusivamente en Soles - PEN).
  * **Monto / Precio de Venta (PEN):** Obligatorio para determinar la base imponible y el valor final de la transacción.
  * **Fecha de Emisión:** Obligatoria fiscalmente para el devengado del impuesto (IGV/Renta).
  * **Fecha de Vencimiento:** Obligatoria para ventas al crédito (no puede ser anterior a la fecha de emisión).

* **Campos Opcionales y Personalizables:**
  * **Notas / Observaciones de Venta:** Campo libre para especificar detalles del pedido o del despacho.

---

### [Proyección Automática de Pasivos Internacionales]

* **Descripción:** Administra y proyecta las obligaciones financieras contratadas en dólares (USD) a través de tarjetas de crédito internacionales, agrupando consumos dentro del ciclo de facturación activo, mostrando proyecciones de cobro y permitiendo la liquidación mediante el registro de pago.

* **Opciones y Flujos Personalizables:**
  * Creación dinámica de nuevas tarjetas de crédito mediante el formulario "Nueva Tarjeta".
  * Edición y eliminación de tarjetas existentes en el catálogo de Tesorería.
  * Registro y liquidación total del ciclo de deuda activa mediante el botón "Registrar Pago" (reinicio del saldo a $0.00).

* **Campos Obligatorios:**
  * **Nombre del Banco / Entidad Financiera:** Obligatorio para la identificación de la cuenta o instrumento financiero.
  * **Últimos 4 dígitos de la Tarjeta:** Obligatorio para la trazabilidad y diferenciación de los instrumentos de pago.
  * **Día de Corte (del 1 al 31):** Obligatorio para determinar la agrupación de operaciones por ciclo de facturación.
  * **Día de Pago:** Obligatorio técnicamente para proyectar la fecha límite de cancelación de la deuda.
  * **Límite de Crédito (USD):** Obligatorio para el control del disponible de la tarjeta.
  * **Moneda de Deuda:** Obligatoria (exclusivamente en USD) para la correcta proyección del pasivo internacional.

* **Campos Opcionales y Personalizables:**
  * Ninguno. Todos los parámetros son requeridos para estructurar el cálculo de pasivos.

---

### [Dashboard Mensual Financiero]

* **Descripción:** Módulo de analítica visual que consolida las ventas efectivamente cobradas en el mes, aplica los algoritmos de costos reales en moneda nacional a cada producto vendido y muestra la ganancia neta total consolidada mediante un gráfico dinámico.

* **Opciones y Flujos Personalizables:**
  * Consolidación automática de ganancias derivadas tanto de "Pedidos Generales" como de "Pedidos de Clientes".
  * Reescritura dinámica de indicadores en función de los parámetros de costo configurados.

* **Campos Obligatorios:**
  * **Ventas Cobradas del Mes:** Obligatorio como insumo principal del cálculo financiero.
  * **Costo del Producto (USD):** Obligatorio para la aplicación de la fórmula financiera.
  * **Tipo de Cambio Aplicado:** Obligatorio para la conversión monetaria de la fórmula.
  * **Factor de Costo de Importación (USD):** Obligatorio para el cálculo del Costo Real en PEN.
  * **Precio de Venta en Soles (PEN):** Obligatorio para la determinación de la utilidad bruta y neta.

* **Campos Opcionales y Personalizables:**
  * **Factor de Ajuste de Costo (+0.07 u otro):** Configurable globalmente desde la pantalla de reglas de negocio.

---

### [Módulo de Configuración General]

* **Descripción:** Interfaz centralizada para la administración de las reglas de negocio globales de la aplicación *local-first*, permitiendo modificar los parámetros operativos del sistema sin requerir intervención en el código fuente.

* **Opciones y Flujos Personalizables:**
  * Edición de los Datos de la Empresa del perfil (Razón Social, RUC, Correo y Dirección Fiscal).
  * Modificación libre de los límites de permanencia para alertas de inventario.
  * Ajuste del factor aditivo de importación/flete.
  * Actualización del valor de tipo de cambio de respaldo tributario/comercial.

* **Campos Obligatorios:**
  * **Días para Remate:** Obligatorio. Define el parámetro numérico de días límite para la activación de alertas visuales en inventario.
  * **Costo de Importación (Factor en USD):** Obligatorio. Numérico decimal que representa el costo adicional por flete o aduana sumado a la tasa de cambio en las compras (predeterminado en +0.07).
  * **Tipo de Cambio de Respaldo (Fallback):** Obligatorio. Valor numérico en Soles por Dólar ante indisponibilidad de la API externa (predeterminado en 3.75).

* **Campos Opcionales y Personalizables:**
  * No aplican campos opcionales; todos los elementos expuestos constituyen parámetros requeridos para la lógica del sistema.

---

### [Autenticación y Control de Acceso Local]

* **Descripción:** Garantiza el acceso seguro al sistema de escritorio mediante un esquema de cuenta única local, permitiendo la protección opcional de los datos del negocio ante terceros no autorizados con acceso al equipo físico.

* **Opciones y Flujos Personalizables:**
  * Activación o desactivación opcional de la solicitud de clave al iniciar la aplicación.
  * Definición, cambio o remoción de la contraseña local de acceso.

* **Campos Obligatorios:**
  * **Estado de Protección por Contraseña:** Obligatorio para definir si el sistema debe solicitar autenticación al iniciar (Habilitado / Deshabilitado).

* **Campos Opcionales y Personalizables:**
  * **Contraseña Local:** Opcional. Si el estado de protección está deshabilitado, la aplicación omitirá la pantalla de login e ingresará directamente al panel principal.
  * **Datos de Empresa del Perfil:** Capturados en el asistente de primera ejecución y editables posteriormente en Configuración: Razón Social (obligatoria), RUC (opcional, 11 dígitos con prefijo 10/20), Correo Electrónico (opcional) y Dirección Fiscal (opcional), usados como identidad corporativa en la documentación del negocio.

---

### [Gestión de Respaldos y Copias de Seguridad (Backups)]

* **Descripción:** Ejecuta copias de seguridad de la base de datos local SQLite hacia una ubicación elegida por el usuario para prevenir la pérdida de información por fallos en el disco o del sistema operativo.

* **Opciones y Flujos Personalizables:**
  * Generación de copias de seguridad manuales bajo demanda mediante el botón "Crear Backup".
  * Programación de copias de seguridad automáticas al cerrar la aplicación o según una rutina periódica.
  * Exploración y selección de la carpeta de destino utilizando un selector interactivo nativo del sistema operativo (*Directory Picker*).

* **Campos Obligatorios:**
  * **Ruta / Directorio de Destino:** Obligatorio. Ubicación del disco local o unidad externa seleccionada mediante el diálogo del sistema operativo para almacenar el archivo de respaldo.
  * **Nombre / Estampa del Archivo:** Generado de forma automática con formato `backup_YYYYMMDD_HHMMSS.db`.

* **Campos Opcionales y Personalizables:**
  * **Frecuencia de Backup Automático:** Opcional (Opciones: Desactivado, Al cerrar el sistema, Diario, Semanal).

---

### [Sincronización Opcional en la Nube (Cloud Sync)]

* **Descripción:** Permite replicar o enviar una copia contingente de los datos desde la base de datos SQLite local hacia una base de datos PostgreSQL hospedada en la nube, garantizando resguardo remoto y disponibilidad opcional.

* **Opciones y Flujos Personalizables:**
  * Conmutador global para activar o desactivar el módulo de sincronización (*Opt-in*).
  * Sincronización manual iniciada por el usuario (botón "Sincronizar ahora") o automática en segundo plano cuando el equipo disponga de conexión a internet.

* **Campos Obligatorios:**
  * **Estado del Servicio de Sincronización:** Obligatorio (Activado / Desactivado).
  * **Host / Servidor Postgres:** Obligatorio si la sincronización está activa (dirección IP o dominio del servidor remoto).
  * **Puerto de Conexión:** Obligatorio si la sincronización está activa (predeterminado: `5432`).
  * **Nombre de la Base de Datos Remota:** Obligatorio si la sincronización está activa.
  * **Usuario de Base de Datos:** Obligatorio si la sincronización está activa.
  * **Contraseña de Conexión:** Obligatorio si la sincronización está activa.

* **Campos Opcionales y Personalizables:**
  * **Intervalo de Sincronización Automática:** Opcional (ej. Cada N horas o únicamente bajo demanda).

---

### [Preferencias de Interfaz y Tema Visual]

* **Descripción:** Permite al usuario personalizar la apariencia visual de la aplicación para adaptar la experiencia de lectura e interfaz de trabajo según su entorno de iluminación.

* **Opciones y Flujos Personalizables:**
  * Conmutación instantánea entre temas de color.
  * Detección y adopción automática del tema del sistema operativo.

* **Campos Obligatorios:**
  * **Tema Activo:** Obligatorio para determinar los estilos gráficos de React aplicados en la interfaz (predeterminado: "Tema del Sistema").

* **Campos Opcionales y Personalizables:**
  * **Modo de Apariencia:** Opcional entre "Claro", "Oscuro" o "Sincronizado con el Sistema Operativo".