# Spec de Lógica — Portafolio

**Status:** v0.2
**Stack:** Go + templ + HTMX + templ-islands
**Arquitectura:** Hexagonal (scaffold con go-arch MCP)
**Base de datos:** PostgreSQL
**Diseño:** ver `design_spec.md` (Neobrutalismo Tech / Cyber-Editorial + Regla de Revelado de Color al Hover)
**Secciones de referencia:** `clean-portfolio` (estructura), no su estética

---

## 1. Arquitectura

Server-rendered Go SPA (scroll de una página), scaffold con **go-arch hexagonal**.

- **templ** — componentes tipados que compilan a Go (`_templ.go`).
- **HTMX** — interactividad server-driven: envío de formulario, modales, copiado de email.
- **templ-islands** — islands para componentes con comportamiento client-side real.
- **PostgreSQL** — única fuente de datos para todas las secciones y el sidebar.
- **CSS puro con custom properties** — tokens del design spec. Sin Tailwind.
- **Auth**: login único de admin (cuenta del dueño) con sesión por cookie + protección de rutas CRUD.

### Regla de fallback (islands)
Toda island mantiene `hx-post` + `hx-swap` como fallback server-driven cuando el runtime de islands no está presente.

## 2. Estructura de carpetas (go-arch hexagonal)

```
.
├── cmd/
│   └── server/
│       └── main.go              # bootstrap: config, rutas, estáticos, runtime islands
├── internal/
│   ├── config/
│   │   └── config.go            # env vars tipadas (ver §10)
│   ├── core/                    # dominio (puertos)
│   │   ├── models/              # structs de dominio (§5)
│   │   └── ports/               # interfaces de repositorio y servicios
│   ├── adapters/                # adaptadores (infraestructura)
│   │   ├── db/
│   │   │   ├── postgres.go      # conexión pgx + migraciones
│   │   │   └── repos/           # implementaciones de repositorios (CRUD)
│   │   ├── cloudinary/
│   │   │   └── uploader.go      # subida de imágenes a Cloudinary
│   │   ├── mailer/
│   │   │   └── contact.go       # envío del formulario (servicio TBD §8)
│   │   └── auth/
│   │       ├── password.go      # hash/verificación de password
│   │       └── session.go       # sesiones por cookie
│   ├── app/                     # casos de uso (aplicación)
│   │   ├── sections/            # use cases por sección (CRUD)
│   │   └── auth/                # login, logout, sesión
│   └── handlers/                # HTTP handlers
│       ├── page.go              # GET / — render página completa (público)
│       ├── section_api.go       # CRUD JSON de secciones (auth)
│       ├── sidebar_api.go       # CRUD JSON de sidebar/icons (auth)
│       ├── auth.go              # POST /login, POST /logout
│       ├── upload.go            # subida de imagen: compresión local → Cloudinary
│       └── contact.go           # POST /api/contact (público)
├── views/
│   ├── layout/
│   │   └── base.templ           # shell HTML: head, css, fonts, runtime islands, slots
│   ├── components/
│   │   ├── sidebar_left.templ   # sidebar izquierdo: tema + idioma + socials + mail
│   │   ├── sidebar_right.templ  # sidebar derecho: secciones (scroll)
│   │   ├── crud_button.templ    # icono junto a título/card → abre modal CRUD (solo admin)
│   │   ├── crud_modal.templ     # modal genérico CRUD (formulario por sección)
│   │   ├── login_modal.templ    # modal de login (oculto, icono en footer)
│   │   ├── project_modal.templ  # modal detalle de proyecto (§9)
│   │   ├── copy_email.templ     # island: botón email con clipboard (§6)
│   │   └── footer.templ         # pie: CTA, copyright, icono login oculto
│   ├── sections/
│   │   ├── hero.templ
│   │   ├── skills.templ
│   │   ├── projects.templ
│   │   ├── experience.templ
│   │   ├── education.templ
│   │   └── contact.templ
│   └── islands/
│       ├── theme_toggle.templ   # island: light/dark
│       ├── lang_toggle.templ    # island: ES/EN
│       └── contact_form.templ   # island: formulario con errores inline
├── static/
│   ├── css/                     # tokens.css, base.css, components.css
│   ├── js/                      # theme.js (anti-FOUC)
│   └── fonts/
├── migrations/                  # SQL de migraciones (versionadas)
├── islands_gen.go               # generado: templ-islands generate
├── Dockerfile                   # multi-stage (§11)
├── docker-compose.yml           # postgres + app (local/Dokploy)
├── design_spec.md
├── logic_spec.md
└── go.mod
```

## 3. Secciones y layout (orden de render)

### Layout de frame (fijo)
- **Sidebar izquierdo** (`sidebar_left`): columna fija con — toggle tema (claro/oscuro), toggle idioma (ES/EN), iconos GitHub, LinkedIn, Mail (clipboard). Íconos SVG monocromáticos con reveal de color al hover.
- **Sidebar derecho** (`sidebar_right`): columna fija con las secciones de la página; al tocarlas hace **scroll suave** a la sección (`scroll-behavior: smooth` + active state).

### Secciones (scroll principal)
| # | Sección | Contenido |
|---|---------|-----------|
| 1 | Hero | Saludo, nombre, subtítulo terminal, `DESCARGAR_CV`, avatar vinilo |
| 2 | Skills | Grid de tarjetas: icono oficial, nombre, nivel |
| 3 | Projects | Cards: título + tecnología (a definir). Click / "Ver más" → **modal detalle** (§9) |
| 4 | Experience | Timeline: periodo, cargo, empresa, descripción |
| 5 | Education | Cards: título, institución, fecha, descripción |
| 6 | Contact | Formulario de envío + email oculto con clipboard |
| 7 | Footer | CTA, copyright con año dinámico, **icono de login oculto** |

## 4. Base de datos (PostgreSQL)

### Modelo relacional

```sql
-- Perfil / hero
profile (
  id SERIAL PRIMARY KEY,
  name TEXT, role TEXT, headline TEXT, summary TEXT,
  email TEXT NOT NULL,          -- nunca renderizado en pantalla
  resume_url TEXT, avatar_url TEXT
)

-- Íconos del sidebar izquierdo
sidebar_icons (
  id SERIAL PRIMARY KEY,
  position INTEGER,             -- orden
  kind TEXT,                    -- 'theme' | 'lang' | 'social' | 'mail'
  name TEXT,                    -- 'GitHub', 'LinkedIn'
  url TEXT,                     -- para kind='social'
  icon_key TEXT,                -- key del SVG inline
  enabled BOOLEAN DEFAULT true
)

-- Secciones del sidebar derecho
sidebar_sections (
  id SERIAL PRIMARY KEY,
  position INTEGER,
  label TEXT,                   -- 'Skills', 'Proyectos', ...
  section_ref TEXT,             -- 'skills', 'projects', ...
  enabled BOOLEAN DEFAULT true
)

skills (
  id SERIAL PRIMARY KEY,
  position INTEGER,
  name TEXT, level TEXT,
  icon_url TEXT,                -- Cloudinary o CDN
  color TEXT, bg_color TEXT
)

projects (
  id SERIAL PRIMARY KEY,
  position INTEGER,
  title TEXT, description TEXT,
  category TEXT, status_label TEXT,
  tags TEXT[],                  -- tecnologías
  link TEXT, repo_link TEXT,
  cover_url TEXT                -- imagen principal (Cloudinary)
)

project_images (
  id SERIAL PRIMARY KEY,
  project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
  position INTEGER,
  url TEXT                      -- imagen del proyecto (Cloudinary)
)

experience (
  id SERIAL PRIMARY KEY,
  position INTEGER,
  period TEXT, position_name TEXT,
  company TEXT, description TEXT
)

education (
  id SERIAL PRIMARY KEY,
  position INTEGER,
  title TEXT, school TEXT,
  date TEXT, description TEXT
)

admin (
  id SERIAL PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,   -- cuenta única (la del dueño)
  password_hash TEXT NOT NULL
)

sessions (
  id TEXT PRIMARY KEY,             -- cookie token (random)
  admin_id INTEGER REFERENCES admin(id),
  expires_at TIMESTAMPTZ NOT NULL
)
```

### Acceso
- Conexión `pgxpool`, migraciones SQL versionadas en `migrations/`.
- Todos los repositorios implementan puertos (`ports`): `ProfileRepo`, `SidebarIconRepo`, `SectionRepo`, `SkillRepo`, `ProjectRepo`, `ExperienceRepo`, `EducationRepo` — cada uno con `List/Get/Create/Update/Delete`.

## 5. Data model de dominio (core/models)

Mismo shape que las tablas: `Profile`, `SidebarIcon`, `SidebarSection`, `Skill`, `Project`, `ProjectImage`, `Experience`, `Education`, `Admin`, `Session`. JSON tags para la API.

## 6. Email oculto + clipboard (REQUISITO)

**El email jamás se muestra en texto plano en pantalla ni en `href="mailto:"` visible.**

- Único elemento visible: botón/icono de sobre en el sidebar izquierdo.
- Click → `navigator.clipboard.writeText(email)` + feedback "Copiado" (~2s), respetando `prefers-reduced-motion`.
- El valor viaja en `data-email` (no visible); decodificado del HTML.
- Anti-scraping opcional: ofuscación base64/split en el HTML, decodificación en cliente.
- Accesibilidad: `:focus-visible` + `aria-label` descriptivo.
- El email real sale de `profile.email` en DB, nunca hardcodeado.

## 7. CRUD + Admin (REQUISITO)

### Patrón CRUD (secciones + sidebar)
- **Icono junto al título de cada sección/card** (`crud_button`) — visible **solo cuando hay sesión de admin**.
- Click → abre **modal CRUD** (`crud_modal`) con formulario por entidad: crear / editar / eliminar.
- Operaciones vía `hx-post`/`hx-put`/`hx-delete` contra `section_api.go` (JSON, auth required).
- Después de mutar: re-render de la sección (server-side) vía HTMX.

### Login (oculto)
- **Icono de login solo en el footer** (sutil, sin destacar) → abre `login_modal`.
- Formulario: usuario + password → `POST /login` → cookie de sesión (HttpOnly, Secure, SameSite=Lax, expiración ~24h).
- **Cuenta única**: el admin se crea en DB (migración seed o comando). Sin registro público.
- Password con `bcrypt`/`argon2`; nunca en texto plano.
- Middleware de auth protege todos los endpoints CRUD y de subida (`/api/section/*`, `/api/sidebar/*`, `/api/upload`). 401 → modal de login.

## 8. Formulario de contacto (servicio TBD)

- Campos: **Nombre**, **Email**, **Mensaje** — validación HTML5 + errores inline (island `contact_form`).
- Estados: enviando / éxito / error. Fallback HTMX: `hx-post="/api/contact"`.
- Rate limiting básico por IP.
- Servicio de envío **gratuito** (investigar): candidatos Resend (SDK Go, 3k/mes gratis), SMTP directo (Gmail u otro), Formspree.
- El handler delega en `adapters/mailer` — un adapter intercambiable según servicio.

## 9. Proyectos — modal detalle (REQUISITO)

- Cards: **título + tecnología** (formato final a definir).
- Click en la card o botón "Ver más" → `project_modal` con:
  - Galería de imágenes del proyecto (de `project_images`)
  - Tecnologías (tags)
  - Descripción
  - Texto informativo adicional
  - Links: **demo** (`link`) y **repositorio** (`repo_link`)
- El modal se abre con HTMX (`hx-get` al detalle) o island re-render; accesible (focus trap, ESC cierra).

## 10. Cloudinary — imágenes (REQUISITO)

**Política: compresión y redimensionado SIEMPRE en local; Cloudinary solo almacena y entrega.**

### Pipeline de subida (admin)
1. Admin sube imagen en el modal CRUD (file input).
2. **En el servidor Go** (handler `upload.go`): redimensiona + comprime a **WebP** localmente (ej. `github.com/chai2010/webp` / `golang.org/x/image` o `imgproxy`-style). Sin gastar créditos de transformación.
3. Sube el WebP ya optimizado a Cloudinary (upload API, sin transformations).
4. Guarda la URL devuelta en DB (`icon_url`, `cover_url`, `url`, `avatar_url`).
5. Entrega: `<img src="...">` directo desde Cloudinary con `fetch_format=auto` (entrega gratis) + `loading="lazy"` (excepto avatar) + `grayscale` en reposo → color al hover.

### Límites por tipo de imagen (a definir)
- Avatar: ~400px, quality 80
- Cover de proyecto: ~1200px wide, quality 80
- Iconos skills: ~64-128px, quality 80
- Screenshots de proyecto: ~1200px wide

## 11. Configuración (env vars)

```
# Servidor
PORT=8080
APP_ENV=production|development
SESSION_SECRET=...                 # firma de cookies

# PostgreSQL (Dokploy service)
DATABASE_URL=postgres://user:pass@postgres:5432/portafolio

# Admin seed
ADMIN_USERNAME=facundo
ADMIN_PASSWORD_HASH=...            # se genera con comando, no en texto plano

# Cloudinary
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...

# Email (servicio TBD, §8)
CONTACT_EMAIL=fds1288@gmail.com    # destinatario real (nunca renderizado)
CONTACT_SERVICE=resend             # resend | smtp | formspree
RESEND_API_KEY=...                 # si aplica
SMTP_HOST=... SMTP_USER=... SMTP_PASS=...
```

- Secrets nunca hardcodeados; `.env` local fuera de git.
- `config.go` valida en startup: fail fast con mensaje claro.

## 12. Despliegue — Dockerfile + Dokploy (REQUISITO)

### Dockerfile (multi-stage)
```dockerfile
# 1. Build stage
FROM golang:1.24 AS build
WORKDIR /src
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN templ generate
RUN go build -o /out/server ./cmd/server

# 2. Runtime stage
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/server /server
COPY --from=build /src/static /static
COPY --from=build /src/migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/server"]
```

### docker-compose.yml
```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: portafolio
      POSTGRES_USER: portafolio
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U portafolio"], interval: 5s }

  app:
    build: .
    env_file: .env
    depends_on:
      postgres:
        condition: service_healthy
    ports: ["8080:8080"]

volumes:
  pgdata:
```

- En Dokploy: proyecto → servicio Docker/compose, dominio con HTTPS automático, `.env` en la instancia, volumen persistente para Postgres (backup).
- El build stage corre `templ generate` (nunca commitear `_templ.go` obsoleto).

## 13. Tema (light/dark) e idioma (ES/EN)

- Tema: light = default (design spec §2), dark con clase en `<html>`, toggle island, persistencia `localStorage` + `prefers-color-scheme`, anti-FOUC con script inline en `<head>`, `aria-pressed`/`aria-label`.
- Idioma: toggle ES/EN en sidebar izquierdo. **Alcance a definir**: textos fijos de la UI, datos del admin (títulos de secciones, labels de skills) o ambos. v0.1 sugiere al menos UI + labels de sidebar.
- Preferencia de idioma persistida (`localStorage` o cookie).

## 14. Accesibilidad y rendimiento

- `prefers-reduced-motion`: desactiva rotación del vinilo y transiciones de color.
- `:focus-visible` en todos los hovers (design spec §6.4).
- Modales accesibles: focus trap, ESC para cerrar, `aria-modal`.
- Fuentes locales con `font-display: swap`; preload solo la fuente crítica.
- Meta tags OG/Twitter; `lang` dinámico según idioma.

## 15. Pendientes / decisiones abiertas

1. **Servicio de email** del formulario (§8) — investigar candidatos gratuitos.
2. **Formato exacto de la card de proyecto** (título + tecnología, resto a definir).
3. **Alcance del toggle ES/EN** (§13).
4. **Partículas de fondo**: ¿canvas o solo watermark? (design spec no las menciona).
5. **CV**: ¿se mantiene `DESCARGAR_CV` en hero y dónde vive el PDF (local vs Cloudinary)?
6. **Seed de datos**: migración inicial con los datos reales del portafolio (desde el JSON del clean-portfolio).
7. **Medidas de compresión exactas** de Cloudinary (§10).
