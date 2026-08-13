# API Contract — Portfolio Go

**Status:** Draft v0.1
**Base URL:** `/api/v1`
**Auth:** cookie `admin_session` (HttpOnly) para endpoints admin. Login público.
**Respuestas:** JSON `application/json`. Errores: `{"error": "mensaje"}` con status HTTP.
**IDs:** numéricos (BIGSERIAL). El front admin manda el ID en las mutaciones, no claves naturales.

---

## 1. Modelos (JSON shape)

### Profile
```json
{
  "name": "Facundo Salvucci",
  "role_es": "Desarrollador Full Stack",
  "role_en": "Full Stack Developer",
  "headline_es": "...",
  "headline_en": "...",
  "summary_es": "...",
  "summary_en": "...",
  "email": "fds1288@gmail.com",
  "avatar_url": "/static/uploads/avatar.webp",
  "resume_url": "/static/uploads/cv.pdf",
  "socials": [ { "id": 1, "position": 0, "name": "GitHub", "url": "https://...", "icon_key": "github" } ],
  "skills": [ { "id": 1, "position": 0, "name": "Go", "icon_url": "/static/uploads/go.webp", "is_tool": false } ]
}
```

### Skill
```json
{ "id": 1, "position": 0, "name": "Go", "icon_url": "/static/uploads/go.webp", "is_tool": false }
```

### Project
```json
{
  "id": 1,
  "position": 0,
  "title_es": "GAIA",
  "title_en": "GAIA",
  "description_es": "...",
  "description_en": "...",
  "tech_description_es": "...",
  "tech_description_en": "...",
  "category": "AI",
  "tags": ["Go", "templ", "HTMX"],
  "link": "https://...",
  "repo_link": "https://github.com/...",
  "cover_url": "/static/uploads/project-cover.webp",
  "screenshots": ["/static/uploads/project-1.webp", "/static/uploads/project-2.webp"]
}
```

### Experience
```json
{ "id": 1, "position": 0, "company": "Dubbz", "position_es": "...", "position_en": "...", "period_es": "2023 - 2025", "period_en": "2023 - 2025", "description_es": "...", "description_en": "..." }
```

### Education
```json
{ "id": 1, "position": 0, "title_es": "...", "title_en": "...", "school": "UTN", "date": "2019 - 2021", "is_course": false, "description_es": "...", "description_en": "..." }
```

### Auth
```json
// login request
{ "password": "..." }
// login response
{ "ok": true }
```

---

## 2. Endpoints públicos (sin auth)

| Método | Path | Descripción | Respuesta |
|---|---|---|---|
| GET | `/api/v1/profile` | Perfil completo (profile + socials + skills embebidos) | `Profile` |
| GET | `/api/v1/skills` | Lista de skills | `[]Skill` |
| GET | `/api/v1/projects` | Lista de proyectos (sin screenshots) | `[]Project` |
| GET | `/api/v1/projects/{id}` | Proyecto detalle (con screenshots) | `Project` |
| GET | `/api/v1/experience` | Lista de experiencias | `[]Experience` |
| GET | `/api/v1/education` | Lista de educación | `[]Education` |
| POST | `/api/v1/contact` | Enviar mensaje de contacto | `200 {"ok":true}` |

## 3. Endpoints admin (requieren cookie `admin_session`)

| Método | Path | Descripción | Payload |
|---|---|---|---|
| POST | `/api/v1/auth/login` | Login con password | `{"password":"..."}` |
| POST | `/api/v1/auth/logout` | Logout | — |
| PUT | `/api/v1/profile` | Actualizar perfil (campos bilingües) | `Profile` (parcial) |
| POST | `/api/v1/profile/avatar` | Subir avatar (multipart `avatar`) | file |
| POST | `/api/v1/profile/cv` | Subir CV (multipart `cv`) | file |
| POST | `/api/v1/skills` | Crear skill | `Skill` sin id |
| PUT | `/api/v1/skills/{id}` | Actualizar skill | `Skill` |
| DELETE | `/api/v1/skills/{id}` | Eliminar skill | — |
| POST | `/api/v1/projects` | Crear proyecto | `Project` sin id |
| PUT | `/api/v1/projects/{id}` | Actualizar proyecto | `Project` |
| DELETE | `/api/v1/projects/{id}` | Eliminar proyecto | — |
| POST | `/api/v1/projects/{id}/images` | Subir screenshots (multipart `screenshots[]`) | files |
| DELETE | `/api/v1/projects/{id}/images/{imageId}` | Eliminar screenshot | — |
| POST | `/api/v1/experience` | Crear experiencia | `Experience` sin id |
| PUT | `/api/v1/experience/{id}` | Actualizar experiencia | `Experience` |
| DELETE | `/api/v1/experience/{id}` | Eliminar experiencia | — |
| POST | `/api/v1/education` | Crear educación | `Education` sin id |
| PUT | `/api/v1/education/{id}` | Actualizar educación | `Education` |
| DELETE | `/api/v1/education/{id}` | Eliminar educación | — |
| PUT | `/api/v1/socials` | Actualizar socials (email + github + linkedin) | `Profile.socials` |

**Nota de migración desde el front HTMX actual:** los forms admin hoy mandan campos con `old_name`/`old_title`/`old_company` como claves naturales. Con REST JSON esto cambia a `id` en el body. El frontend admin se adapta (fetch + JSON), el diseño visual queda idéntico.

---

## 4. Orden y posiciones

- Todas las listas ordenables (`skills`, `projects`, `experience`, `education`, `socials`) tienen campo `position` (int).
- `GET` devuelve las listas **ordenadas por position ASC**.
- Para reordenar, el admin actualiza `position` vía PUT del item (o endpoint de reorder a futuro).

## 5. Uploads

- Los archivos se suben como `multipart/form-data` y se guardan en `static/uploads/` (por ahora local; el adapter Cloudinary se pluguea después sin cambiar el contrato).
- Nombres generados: `<entidad>-<unix>.webp` / `.pdf`.
- La URL resultante es `/static/uploads/<archivo>` y se guarda en el campo correspondiente.

## 6. Errores

| Status | Uso |
|---|---|
| 400 | payload inválido / campos requeridos faltantes |
| 401 | sin sesión admin o password incorrecto |
| 404 | entidad no encontrada |
| 409 | conflicto (ej. nombre duplicado) |
| 500 | error interno |

Formato: `{"error": "descripción"}`.

## 7. Idiomas

- Los campos bilingües (`_es`/`_en`) se envían SIEMPRE ambos en las mutaciones admin (el front los edita en tabs ES/EN).
- El render público elige según `lang` (cookie > Accept-Language > es).
- Tema: dark por defecto (localStorage ausente → dark).
