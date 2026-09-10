## 4. Casos de Borde (Edge Cases) y Excepciones de Negocio

Como parte de la estrategia de aseguramiento de la calidad (QA) y estabilidad del MVP local-first, se identifican los siguientes escenarios atípicos, límites de bordes y excepciones operativas por módulo:

### 4.1. Importaciones, Tipo de Cambio y Control Aduanero

* **Falla de API de Tipo de Cambio / Conexión Intermitente:**
* *Escenario:* El usuario intenta registrar un pedido en USD sin conexión a internet o cuando la API SBS/SUNAT no responde.


* *Comportamiento esperado:* El sistema debe aplicar de forma transparente el valor de **Tipo de Cambio de Respaldo (*Fallback*)** configurado (predeterminado 3.75 PEN) indicando visualmente una alerta de "Modo Contingencia / Sin Conexión".




* **Superación del Límite Aduanero Simplificado ($220 USD):**
* *Escenario:* La suma de los costos USD de los pedidos asignados a un mismo Lote de Importación excede el parámetro de $220 USD ($220.01 USD o más).


* *Comportamiento esperado:* El sistema no debe bloquear la transacción (para no paralizar la operación), pero emitirá una advertencia crítica indicando que el lote supera el tope aduanero y requerirá la confirmación explícita del usuario para continuar.


* **Anulación de Pedido con Pasivo Liquidado:**
* *Escenario:* Se solicita la anulación por defecto/fallo de un producto cuyo ciclo de tarjeta de crédito en USD ya fue cancelado y registrado como pagado.


* *Comportamiento esperado:* La aplicación debe procesar el reintegro como un nota de crédito/saldo a favor en la cuenta del proveedor o ingreso extraordinario, sin alterar el saldo histórico de ciclos de tarjeta ya liquidados.



### 4.2. Control de Inventario y Alertas de Remate

* **Desfase del Reloj del Sistema Operativo (*Clock Skew*):**
* *Escenario:* El usuario altera manualmente la fecha/hora de la computadora local (ej. cambia de 2026 a 2025 o viceversa).
* *Comportamiento esperado:* Todas las fechas del sistema deben almacenarse e interpretarse en formato ISO-8601 / UTC. El cálculo de días de permanencia para remate debe basarse en la diferencia estricta de días calendario transcurridos; la validación del formulario de recepción bloquea e informa cuando la fecha de ingreso es posterior a la fecha actual del sistema.


* **Reconfiguración Dinámica del Límite de Permanencia:**
* *Escenario:* El usuario reduce el parámetro "Días para Remate" en Configuración General (por ejemplo, de 25 a 10 días) teniendo inventario antiguo registrado.


* *Comportamiento esperado:* El motor de inventario debe recalcular inmediatamente las etiquetas visuales de "Remate" en tiempo real sobre la vista del catálogo sin necesidad de reiniciar la aplicación.


* **Ajuste Manual de Stock a Valores Cero o Negativos:**
* *Escenario:* El usuario intenta realizar un ajuste manual de inventario ingresando un número menor o igual a cero.
* *Comportamiento esperado:* La interfaz de React validará a nivel de formulario que la cantidad ingresada sea un entero positivo ($>0$). No se permitirán stocks negativos en el Kardex.



### 4.3. Ventas Locales (PEN) y Facturación

* **Inconsistencia de Fechas en Ventas al Crédito:**
* *Escenario:* El usuario ingresa una fecha de vencimiento anterior a la fecha de emisión del comprobante.


* *Comportamiento esperado:* Validación estricta en el formulario. La interfaz debe bloquear el envío del registro e informar que la fecha de vencimiento debe ser igual o posterior a la emisión, garantizando el cumplimiento normativo SUNAT.




* **Validación de Estructura de Documento de Identidad (DNI/RUC):**
* *Escenario:* Registro de cliente con un DNI que no tiene 8 dígitos o un RUC que no tiene 11 dígitos o que no inicia con los prefijos legales 10 o 20.


* *Comportamiento esperado:* Control mediante expresiones regulares (Regex) localmente antes de persistir en SQLite, impidiendo la creación de clientes con identificadores fiscales inválidos según SUNAT.




* **Intento de Venta sin Stock Disponible:**
* *Escenario:* Se intenta registrar una venta inmediata ("Pedido General") de un producto con stock 0.
* *Comportamiento esperado:* El sistema alertará sobre la falta de stock en el almacén principal. Si la venta se registra como "Pedido de Cliente" (a pedido), el sistema permitirá la transacción registrando el requerimiento de importación.





### 4.4. Pasivos Internacionales y Ciclos de Tarjeta

* **Días de Corte en Meses de Menor Duración (Día 29, 30 o 31):**
* *Escenario:* Se configura una tarjeta con fecha de corte los días 31, pero el mes activo es Febrero (28/29 días) o un mes de 30 días.


* *Comportamiento esperado:* El algoritmo de proyección en Go ajustará automáticamente la fecha de corte al último día calendario del mes en curso para evitar saltos indeseados de ciclo.


* **Anulación de Venta con Reversión:**
* *Escenario:* El usuario anula una venta ya registrada (contado o crédito).
* *Comportamiento esperado:* La aplicación devolverá el stock reservado y revertirá la deuda del cliente dentro de una misma transacción (sin escrituras parciales), dejando la venta marcada como cancelada en el historial y los indicadores del Dashboard recalculados sobre el estado vigente.


* **Eliminación de Tarjeta con Historial de Operaciones:**
* *Escenario:* El usuario intenta eliminar una tarjeta de crédito del catálogo de Tesorería que posee compras o compras proyectadas asociadas.


* *Comportamiento esperado:* Tras confirmación explícita del usuario (diálogo de advertencia), se aplicará un borrado lógico (*soft delete*) marcando la tarjeta como inactiva para evitar nuevos registros, preservando la integridad referencial en SQLite y las proyecciones financieras históricas del catálogo.



### 4.5. Operaciones de Sistema (Auth, Backup y Sync)

* **Pérdida de Contraseña Local en Cuenta Única:**
* *Escenario:* El usuario olvida la contraseña de acceso configurada para la aplicación.
* *Comportamiento esperado:* Dado que es una arquitectura monousuario *offline-first* sin servidor central de autenticación, el sistema debe permitir la generación de un "Token/Clave de Recuperación" descargable al momento de activar la contraseña por primera vez.


* **Directorio de Backup Inaccesible o Desconectado:**
* *Escenario:* La ruta configurada para respaldos automáticos (ej. una memoria USB extraíble) no está disponible al momento de ejecutar la copia de seguridad.
* *Comportamiento esperado:* El sistema capturará la excepción de I/O, notificará al usuario la imposibilidad de guardar en la ruta configurada y generará una copia de emergencia en el directorio predeterminado interno del sistema (`AppData` o `User Home`).


* **Conflictos de Sincronización en la Nube (Offline vs. Cloud Sync):**
* *Escenario:* Se registran transacciones offline en SQLite y al reconectar a Postgres en la nube existe una divergencia de estados o fallo temporal en la conexión.
* *Comportamiento esperado:* La sincronización se ejecutará en segundo plano mediante transacciones atómicas. En caso de corte de red, el motor reintentará en el siguiente intervalo sin bloquear el uso de la interfaz en React. La base de datos SQLite local se mantiene siempre como la fuente principal de verdad (*Single Source of Truth*).



---

## 5. Matriz de Riesgos Técnicos y Estrategia de Mitigación

A continuación se consolidan los riesgos técnicos y de usabilidad identificados para la arquitectura **Wails + React + SQLite**, junto con su evaluación de impacto y plan de contingencia:

| Categoría | Riesgo Técnico / Usabilidad | Impacto | Probabilidad | Estrategia de Mitigación (QA & Arquitectura) |
| --- | --- | --- | --- | --- |
| **Persistencia** | **Corrupción del archivo SQLite** por corte intempestivo de energía o apagado forzado del sistema. | **Alto** | Media | Habilitar el modo WAL (*Write-Ahead Logging*) en SQLite, aplicar transacciones ACID estrictas y programar respaldos rotativos automáticos. |
| **Desempeño** | **Bloqueo de la interfaz en React (UI congelada)** durante operaciones de procesamiento pesado (ej. sincronización cloud o respaldo de BD). | **Medio** | Media | Manejar todas las llamadas entre Wails (Go) y React de forma asíncrona mediante Goroutines y promesas JS, ejecutando tareas pesadas en segundo plano. |
| **Seguridad** | **Bypass de autenticación local** mediante manipulación directa del archivo de base de datos SQLite por usuarios del equipo. | **Medio** | Baja | Ofrecer soporte para encriptación de la base de datos local usando extensiones como SQLCipher o encriptación de clave hash con bcrypt. |
| **Integración** | **Desalineación en cálculo de Costo Real (PEN)** por cambios abruptos en el Tipo de Cambio o redondeos en decimales. | **Alto** | Media | Utilizar tipos de datos matemáticos de precisión fija (`decimal` / `bigint` en centavos) en Go/SQLite para evitar errores de coma flotante IEEE 754. |
| **Almacenamiento** | **Falta de espacio en disco** al seleccionar una ruta de respaldos locales. | **Medio** | Media | Ejecutar una validación previa de disponibilidad de almacenamiento (*Disk Space Check*) vía Go antes de escribir la copia de seguridad. |
| **Sincronización** | **Falla o Timeout al conectar con la BD Postgres remota**, ralentizando el flujo de trabajo del usuario. | **Medio** | Alta | Aislación total del módulo Cloud Sync. Las fallas de red no deben interrumpir el uso local. La app reintentará la conexión exponencialmente (*Exponential Backoff*). |