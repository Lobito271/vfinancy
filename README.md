# vfinancy

Sistema ERP de escritorio para gestión empresarial (compras, ventas, inventario, tesorería, contabilidad y reportes), construido con **Wails v2** (Go + React + TypeScript) sobre **PostgreSQL**.

Reemplaza los procesos manuales del negocio con una plataforma centralizada, multi-rol (RBAC) y con auditoría completa. La UI es 100 % en español (es-PE).

## Stack

| Capa       | Tecnología                                                                 |
|------------|----------------------------------------------------------------------------|
| Desktop    | Wails v2 (Go 1.23)                                                         |
| Backend    | Go — Clean Architecture + DDD, vertical slices (Service + Repository)  |
| DB         | PostgreSQL (pgx v5), migraciones SQL versionadas                           |
| Frontend   | React 18 + TypeScript + Vite 5                                             |
| Estado     | Zustand, TanStack Query, React Hook Form, Zod                              |
| Estilos    | Plain CSS3 en `src/index.css` (sin Tailwind/PostCSS)                       |

## Estado actual

- **Fase 0 – 0.6 (fundaciones):** completas — scaffold Wails, i18n es-PE, design system, layout, tema, permisos, mock data.
- **Fase 1 (arquitectura de BD):** completa — ~65 tablas, 9 módulos, ERD y estrategias.
- **Fase 1.1 (esquema Módulo 1):** completa — 20 pares de migraciones (auth + administración).
- **Fase 1.2 (capa de dominio):** completa — value objects, enums, validación, entidades ricas.
- **Fase 1.3 (capa de repositorios):** completa — interfaces + implementaciones PostgreSQL por feature.
- **Fase 1.4 (capa de servicios):** completa — servicios de negocio por feature.
- **Fase 1.5 (orquestación en servicios):** completa — composición cross-feature dentro del servicio del feature (sin capa de casos de uso / workflows).
- **Fase 2 (bindings + integración UI):** en curso — bindings de auth, perfil, settings, system + frontend conectado por `wailsClient`.

Ver `PROJECT_PLAN.md` para el plan completo de desarrollo.

## Repositorio

```
backend/            # Go backend (vertical slices: Service + Repository)
frontend/           # React + TypeScript UI
backend/migrations/ # SQL versionado (0000_init … 0019_audit_events)
```

## Requisitos

- Go 1.23+
- Node.js 20+ (npm)
- Wails v2 CLI
- PostgreSQL 16+

## Quick start

```bash
# Backend: aplicar migraciones
go run ./backend/cmd/cli migrate

# Frontend (solo Vite, sin backend)
cd frontend && npm install && npm run dev

# Aplicación completa (Wails)
wails dev
wails build
```

Variables de entorno clave: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `LOG_LEVEL`, `LOG_FORMAT`.

## Pruebas

```bash
go test ./backend/...
cd frontend && npm run check && npm run build
```

## Documentación

- `PROJECT_PLAN.md` — plan completo de desarrollo por fases.
- `AGENTS.md` — reglas del repositorio para agentes (arquitectura, convenciones, comandos).
- `DESIGN.md` — sistema de diseño (tokens, tipografía, componentes, accesibilidad).
- `frontend/README.md` — frontend: stack, convenciones, comandos.
