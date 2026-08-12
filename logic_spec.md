# Spec de Lógica — Portafolio

**Status:** Draft v0.1
**Stack:** Go + templ + HTMX + templ-islands
**Diseño:** ver `design_spec.md` (Neobrutalismo Tech / Cyber-Editorial + Regla de Revelado de Color al Hover)
**Secciones de referencia:** `clean-portfolio` (estructura), no su estética

---

## 1. Arquitectura

Server-rendered Go, una sola página (SPA de scroll), sin framework frontend.

- **templ** — componentes tipados que compilan a Go (`_templ.go`).
- **HTMX** — interactividad server-driven: toggle de tema, envío de formulario, copiado de email.
- **templ-islands** — islands para los componentes con comportamiento client-side real (optimistic UI, feedback inmediato).
- **CSS puro con custom properties** — los tokens del design spec como única fuente de verdad. Sin Tailwind.
- Sin build de frontend: `templ generate` produce Go, los estáticos se sirven directo.

### Regla de fallback (islands)
Toda island mantiene `hx-post` + `hx-swap` como fallback server-driven cuando el runtime de islands no está presente. El servidor JSON es la fuente de verdad; el runtime revierte mutaciones en error.

## 2. Estructura de carpetas

```
.
├── cmd/
│   └── server/
│       └── main.go              # bootstrap: config, rutas, estáticos, runtime islands
├── internal/
│   ├── config/
│   │   └── config.go            # env vars tipadas (ver §9)
│   ├── models/
│   │   └── models.go            # structs de datos (ver §4)
│   ├── data/
│   │   └── data.go              # datos seed del portafolio (ver §5)
│   ├── handlers/
│   │   ├── page.go              # GET / — render de la página completa
│   │   └── contact.go           # POST /api/contact — envío del formulario
│   └── services/
│       ├── contact/
│       │   └── contact.go       # adapter de envío de email (TBD §7)
│       └── images/
│           └── images.go        # resolución de URLs de imágenes externas (TBD §8)
├── views/
│   ├── layout/
│   │   └── base.templ           # shell HTML: head, css, fonts, runtime islands, slots
│   ├── components/
│   │   ├── navbar.templ         # nav fija con anclas
│   │   ├── sidebar.templ        # barra izquierda: redes sociales SVG
│   │   ├── sidebar_sign.templ   # texto vertical derecho (firma)
│   │   ├── watermark.templ      # marca de agua gigante "developer"
│   │   ├── vinyl_avatar.templ   # avatar vinilo concéntrico
│   │   ├── skill_card.templ     # tarjeta de habilidad
│   │   ├── project_card.templ   # tarjeta de proyecto
│   │   ├── footer.templ         # pie: contacto, copyright, año dinámico
│   │   └── copy_email.templ     # island: botón email con clipboard (§6)
│   ├── sections/
│   │   ├── hero.templ           # presentación + avatar vinilo + DESCARGAR_CV
│   │   ├── skills.templ         # grid de habilidades
│   │   ├── projects.templ       # showcase + thumbnails
│   │   ├── experience.templ     # timeline
│   │   ├── education.templ      # cards académicas
│   │   └── contact.templ        # formulario de contacto (§7)
│   └── islands/
│       ├── theme_toggle.templ   # island: light/dark toggle
│       └── contact_form.templ   # island: formulario con errores inline
├── static/
│   ├── css/
│   │   ├── tokens.css           # custom properties del design spec
│   │   ├── base.css             # reset + tipografía + layout frame
│   │   └── components.css       # estilos por componente/sección
│   ├── js/
│   │   └── theme.js             # arranque del tema antes de pintar (FOUC)
│   └── fonts/                   # Space Grotesk, JetBrains Mono, Inter (local)
├── islands_gen.go               # generado: templ-islands generate
├── design_spec.md
└── go.mod
```

**Nota:** los estilos se aplican por sección exactamente donde el design spec indica — tokens globales en `tokens.css`, layout del frame en `base.css`, y cada sección/componente tiene su bloque en `components.css`.

## 3. Secciones (orden de render)

| # | Sección | Componente | Contenido |
|---|---------|-----------|-----------|
| 0 | Fondo | `watermark` + canvas partículas | Marca de agua gigante `developer`, partículas sutiles (opcional) |
| 1 | Barra lateral | `sidebar` + `sidebar_sign` | Iconos sociales SVG a la izquierda; firma vertical derecha |
| 2 | Navbar | `navbar` | Isotipo + anclas (`#hero #skills #projects #experience #education #contact`) + toggle tema + `[ES\|EN]` |
| 3 | Hero | `hero` + `vinyl_avatar` | Saludo, nombre, subtítulo con prefijo terminal, `DESCARGAR_CV` |
| 4 | Skills | `skills` | Grid de tarjetas con icono oficial, nombre, nivel |
| 5 | Projects | `projects` | Showcase principal + columna de thumbnails |
| 6 | Experience | `experience` | Timeline vertical con periodo, cargo, empresa, descripción |
| 7 | Education | `education` | Cards: título, institución, fecha, descripción |
| 8 | Contact | `contact` + `contact_form` | Formulario de envío + email oculto con clipboard |
| 9 | Footer | `footer` | CTA, copyright con año dinámico, links |

## 4. Data model (models.go)

```go
type Profile struct {
    Name        string   // "Facundo Salvucci"
    Role        string   // "Desarrollador Full Stack"
    Headline    string   // línea corta bajo el nombre
    Summary     string   // bio para la sección about/contact
    Email       string   // NUNCA renderizado en texto plano (§6)
    ResumeURL   string   // URL del CV
    AvatarURL   string   // foto de perfil (servicio externo, §8)
}

type Social struct {
    Name string // "GitHub", "LinkedIn"
    URL  string
    Icon string // key del SVG inline
}

type Skill struct {
    Name    string   // "Go"
    Level   string   // "Avanzado" / "Intermedio"
    IconURL string   // icono oficial (CDN devicon o local)
    IconKey string   // alternativa: key del SVG local
}

type Project struct {
    ID          string
    Title       string
    Description string
    Category    string   // "Full-Stack", "AI", "Dashboard"
    Tags        []string
    ImageURL    string   // servicio externo (§8)
    Link        string   // demo live (opcional)
    RepoLink    string   // GitHub (opcional)
    StatusLabel string   // "Production", "In Progress"
}

type Experience struct {
    Period      string // "2023 — Presente"
    Position    string
    Company     string
    Description string // markdown-lite → HTML
}

type Education struct {
    Title       string // "Tecnicatura Universitaria..."
    School      string
    Date        string
    Description string
}

type ContactRequest struct {  // payload del formulario (§7)
    Name    string `json:"name"`
    Email   string `json:"email"`
    Message string `json:"message"`
}
```

## 5. Fuente de datos (data.go)

- **v0.1: datos estáticos en Go** (structs seed) — sin base de datos ni CMS.
- Los datos reales del portafolio se cargan en `data.go` a partir de la información de tu CV/JSON de respaldo (`clean-portfolio/public/assets/facundo-salvucci_*.json`): perfil, experiencia, educación, proyectos, skills.
- El CV (PDF) puede seguir en `static/` o servirse desde el mismo servicio de imágenes/almacenamiento.

## 6. Email oculto + clipboard (REQUISITO)

**El email jamás se muestra en texto plano en pantalla ni en atributo `href="mailto:"` visible.**

- Único elemento visible: un **botón/icono** (`copy_email.templ`) con icono de sobre.
- Al hacer click: `navigator.clipboard.writeText(email)` y feedback visual ("¡Copiado!").
- Implementación como **island de mutación** (re-render local) con fallback HTMX:
  - `// @island CopyEmail endpoint=/api/email... method=POST` — no, el email NO viaja al servidor para copiarse; viaja el valor embebido.
  - El valor del email se embebe en el HTML como `data-email` en el botón (no visible), y el runtime island / un pequeño script lo copia.
  - Fallback sin JS: `hx-post="/api/email-copy"` que devuelve el email en un fragmento solo tras confirmación (o simplemente no-op si no hay JS — el botón queda sin efecto).
- Feedback: check verde + tooltip "Copiado al portapapeles" durante ~2s (respetando `prefers-reduced-motion`).
- Protección anti-scraping (opcional): el email se puede ofuscar en el HTML (codificación base64 o split) y decodificar en el cliente.

## 7. Formulario de contacto (SERVICIO TBD)

### UI
- Campos: **Nombre**, **Email**, **Mensaje** — con validación HTML5 + errores inline.
- Island re-render con `trigger=submit` + spans `data-error-for` por campo.
- Estados: enviando (spinner), éxito (confirmación), error (mensaje inline).
- Fallback HTMX: `hx-post="/api/contact"` + `hx-swap`.

### Servicio de envío — TBD (investigar antes de implementar)
| Candidato | Ventaja | Nota |
|---|---|---|
| **Resend** | API moderna, SDK oficial Go, gratis hasta 3k emails/mes | Fácil, sin SMTP propio |
| **Formspree** | Cero backend propio, endpoint estático | No se integra al server Go |
| **SMTP directo** (net/smtp + Gmail u otro) | Sin dependencia de terceros | Hay que manejar rate limits |
| **SendGrid** | Maduro, buen SDK Go | Overkill para pocos mensajes |

- El handler `POST /api/contact` valida y delega en `services/contact` — un adapter que cambia según el servicio elegido.
- Rate limiting básico por IP (un mensaje cada X segundos) para evitar spam.

## 8. Imágenes — servicio externo (TBD)

| Candidato | Ventaja | Nota |
|---|---|---|
| **Cloudinary** | Transformaciones on-the-fly (resize, formatos webp/avif) | Gratis generoso; la opción más "set and forget" |
| **Cloudflare R2** | Egress gratis, S3-compatible | Requiere pipeline propio de transformación |
| **Supabase Storage** | Gratis, integración simple, bucket público | Buen balance para pocas imágenes |
| **S3 + CloudFront** | Estándar, potente | Más configuración |

- **v0.1:** el modelo guarda `ImageURL` completo (URL absoluta del servicio externo). `services/images` solo normaliza/resuelve URLs y sirve fallbacks si una imagen falta.
- Formato objetivo: **WebP/AVIF**, `loading="lazy"` en thumbnails, `grayscale` en reposo → color al hover (design spec §6).
- Avatar (foto de perfil) se carga eager, no lazy.

## 9. Configuración (env vars)

```
# Servidor
PORT=8080

# Email (servicio TBD, §7)
CONTACT_EMAIL=fds1288@gmail.com          # destinatario real (nunca renderizado)
CONTACT_SERVICE=resend                   # resend | formspree | smtp
RESEND_API_KEY=...                       # si aplica
SMTP_HOST=... SMTP_USER=... SMTP_PASS=...

# Imágenes (servicio TBD, §8)
IMAGE_BASE_URL=https://res.cloudinary.com/...   # si aplica
```

- `config.go` lee y valida env vars; valores ausentes → fail fast en dev con mensaje claro.
- Nunca hardcodear secrets; `.env` local fuera de git (`.gitignore`).

## 10. Tema (light/dark)

- Light = default (design spec §2).
- Dark = clase en `<html>`/`<body>` + toggle island.
- **FOUC prevention:** script inline mínimo en `<head>` que lee `localStorage('theme')` y `prefers-color-scheme` antes del primer paint.
- Persistencia en `localStorage`, toggle con `aria-pressed` y `aria-label` accesibles.

## 11. Accesibilidad y rendimiento

- `prefers-reduced-motion`: desactiva rotación del vinilo, transiciones de color (mantiene el cambio sin animación).
- `:focus-visible` en todos los hovers (design spec §6.4).
- Navegación por teclado funcional (navbar, sidebar, formulario).
- Fuentes locales con `font-display: swap`; preload solo de la fuente de mayor impacto.
- Meta tags OG/Twitter; `lang="es"` en `<html>` (o `[ES|EN]` como estándar del spec — ver §12).

## 12. Pendientes / decisiones abiertas

1. **Servicio de email** para el formulario (§7) — investigar y decidir.
2. **Servicio de imágenes** (§8) — investigar y decidir.
3. **Datos reales**: completar `data.go` con tu perfil, proyectos, experiencia, educación reales.
4. **Idioma de la UI**: el spec de diseño usa `[ES|EN]` — ¿traducción completa (i18n) o solo indicador visual? (v0.1 sugiere solo español, el selector queda visual).
5. **Partículas de fondo**: ¿canvas como clean-portfolio o solo watermark? (el spec de diseño no las menciona; definir).
6. **CV**: ¿PDF local en `static/` o en el servicio de almacenamiento externo?
