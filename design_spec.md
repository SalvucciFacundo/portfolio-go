# Especificación de Diseño (Design Spec): Portafolio Neobrutalista Tech / Editorial

Documento de especificación técnica de diseño y guía de estilos (*Design System Spec*) para crear un portafolio web de estética **Neobrutalista Cyber-Editorial Monocromática**.

---

## 1. Concepto y Estilo Visual

- **Nombre del Estilo**: Neobrutalismo Tech / Cyber-Editorial Minimalista.
- **Esencia**: Inspirado en consolas de comandos, revistas de diseño suizo contemporáneo y portafolios de agencias creativas digitales.
- **Principios**:
  - Monocromía rigurosa y alto contraste tipográfico.
  - Asimetría controlada con elementos en marcas de agua (*watermarks*).
  - Enmarcado rígido con barras laterales y cabeceras fijas (*fixed viewport frame*).

---

## 2. Paleta de Colores y Tokens (Color System)

### Modo Claro (Light Theme - Default)
| Token | Valor Hex/RGBA | Uso |
| :--- | :--- | :--- |
| `--bg-main` | `#F4F4F2` / `#EFEFEF` | Fondo principal (blanco estudio / gris claro) |
| `--text-primary` | `#1A1A1A` | Títulos, nombres y textos principales |
| `--text-secondary` | `#666666` | Subtítulos, labels y metadatos |
| `--text-muted` | `#999999` | Textos de niveles y detalles sutiles |
| `--watermark-color` | `rgba(0, 0, 0, 0.04)` | Textos gigantes de fondo desvanecidos |
| `--card-bg` | `#EAEAEA` | Fondo de tarjetas de habilidades y contenedores |
| `--card-border` | `#E0E0E0` | Bordes sutiles de tarjetas y separadores |
| `--vinyl-outer` | `#D4D4D0` | Anillo exterior del avatar |
| `--vinyl-mid` | `#666666` | Anillo intermedio |
| `--vinyl-inner` | `#1A1A1A` | Anillo interno |
| `--btn-bg` | `#2D2D2D` | Botón principal `DESCARGAR_CV` |
| `--btn-text` | `#FFFFFF` | Texto de botones oscuros |

### Modo Oscuro (Dark Theme - Opcional)
| Token | Valor Hex/RGBA | Uso |
| :--- | :--- | :--- |
| `--bg-main` | `#0E0E10` | Fondo principal estilo cyberpunk oscuro |
| `--text-primary` | `#F4F4F2` | Títulos principales |
| `--text-secondary` | `#A0A0A0` | Subtítulos y labels |
| `--watermark-color` | `rgba(255, 255, 255, 0.03)` | Marcas de agua en oscuro |
| `--card-bg` | `#18181B` | Tarjetas de habilidades |
| `--card-border` | `#27272A` | Bordes en modo oscuro |

---

## 3. Sistema Tipográfico (Typography System)

### Fuentes Recomendadas (Google Fonts)
1. **Titulares y Marcas de Agua**: `Space Grotesk` o `Rajdhani` (Sans-serif geométrica tech).
2. **Consola y Etiquetas (Data/Labels)**: `JetBrains Mono` o `Fira Code` (Monospace industrial).
3. **Lectura / Cuerpo de Texto**: `Inter` (Sans-serif limpia).

### Escala y Reglas Tipográficas
- **Título Hero (Nombre)**:
  - `font-family`: `'Space Grotesk', sans-serif`
  - `font-size`: `3.5rem` a `5rem`
  - `font-weight`: `700` (Bold)
  - `letter-spacing`: `-0.02em`
- **Subtítulos con Prefijo Terminal**:
  - `font-family`: `'JetBrains Mono', monospace`
  - `font-size`: `1rem`
  - Formato: `■ Desarrollador Full Stack` o `> Full Stack Developer`
- **Matriz de Conceptos (Letter-Spacing Extremo)**:
  - `font-family`: `'JetBrains Mono', monospace`
  - `font-size`: `0.75rem`
  - `letter-spacing`: `0.8em` a `1em`
  - `text-transform`: `lowercase`
  - Ejemplo: `b a c k e n d`, `f r o n t e n d`, `i n t e r f a z`
- **Texto Vertical Lateral (Right Sidebar)**:
  - `font-size`: `4rem` a `6rem`
  - `font-weight`: `900`
  - `writing-mode`: `vertical-rl`
  - `opacity`: `0.15` a `0.25`
  - `text-transform`: `uppercase`

---

## 4. Disposición y Maquetación (Layout & Grid Spec)

```
+-------------------------------------------------------------------------+
| [Isotipo]   (☽)                ( ✕ Menu )                [ES|EN]  (🌐)  |
+---+-----------------------------------------------------------------+---+
|   |  [WATERMARK GIGANTE: "developer"]                               | J |
| S |                                                                 | U |
| O |   Hola a todos, soy                                             | N |
| C |   # Carlos Ramos                     /-------------\            | I |
| I |   ■ Desarrollador Full Stack        /  ( Anillo )   \           | O |
| A |                                    |  (  Avatar  )  |           | R |
| L |   [ DESCARGAR_CV ]                  \  ( Vinilo )   /           | E |
| E |                                      \-------------/            | N |
| S |                                                                 | C |
|   |   b a c k e n d  /  f r o n t e n d  /  i n t e r f a z         | O |
+---+-----------------------------------------------------------------+---+
|                [ Footer / Audio-Player Controls / Time ]                |
+-------------------------------------------------------------------------+
```

### Componentes Clave

#### A. Barra de Redes Sociales (Left Sidebar)
- Ubicación: Fija en el margen izquierdo (`position: fixed; left: 0; top: 0; bottom: 0; width: 60px`).
- Alineación: Flexbox columna, iconos alineados al fondo inferior.
- Iconos: SVG monocromáticos (18px x 18px) con efecto `hover` de opacidad (`0.5` -> `1.0`).

#### B. Texto de Firma Vertical (Right Sidebar)
- Ubicación: Fija en el margen derecho (`position: fixed; right: 0; top: 0; bottom: 0`).
- Estilo: Texto en mayúsculas sin espacio entre letras, recortado por el borde de la pantalla.

#### C. Elemento Focal: Avatar Vinilo Concéntrico
- Construcción: Tres capas de divs circulares anidados (`border-radius: 50%`).
- Anillo Exterior: Diámetro `380px` - `450px`, color gris suave.
- Anillo Medio: Diámetro `280px` - `320px`, color gris medio.
- Anillo Interior / Avatar: Diámetro `180px` - `220px`, imagen en escala de grises (`filter: grayscale(100%)`).
- Animación: Rotación sutil continua (`transform: rotate(360deg)` en 60 segundos).

#### D. Tarjetas de Habilidades (Skills Grid)
- Layout: Grid responsivo o Flexbox piramidal centrado (`gap: 16px`).
- Dimensiones: Ancho `130px`, Alto `150px`.
- Bordes: `border-radius: 18px`, borde de 1px sólido `#E0E0E0`.
- Estructura Interna:
  - Icono SVG oficial a color o monocromático en el centro (`width: 40px`).
  - Nombre de la tecnología (`font-size: 0.85rem`, bold).
  - Etiqueta de nivel (`Avanzado` / `Intermedio`) en tipografía monoespaciada pequeña.
- Hover Effect: Elevación vertical `translateY(-6px)` con sombra difusa suave.

#### E. Botón `DESCARGAR_CV`
- Estilo: Rectángulo con esquinas chaflanadas/biseladas o borde plano tipo industrial.
- Background: `#2D2D2D` con texto `#FFFFFF`.
- Fuente: Monospace en mayúsculas (`font-size: 0.8rem`, `letter-spacing: 0.1em`).
- Padding: `12px 24px`.

---

## 5. Resumen de Reglas de Implementación

1. **Sin Colores Estridentes**: Mantener la paleta estricta en tonos neutros (blanco, grises, negro). La única excepción son los iconos oficiales de las tecnologías en la sección de habilidades.
2. **Espaciado Generoso**: Usar padding y márgenes amplios para dar sensación de galería de arte u hoja editorial.
3. **Marcas de Agua No Invasivas**: Los textos de fondo deben tener un `z-index` inferior y una opacidad menor a `0.05` para que no afecten la legibilidad.

---

## 6. Regla de Revelado de Color al Hover (Firma de Interacción)

**Principio**: *La página vive en monocromo; el color es la recompensa de la interacción.*

- **Estado de reposo**: Fotos, iconos (incluidos los de habilidades) y thumbnails se renderizan en escala de grises (`filter: grayscale(100%)`) o con opacidad reducida.
- **Estado hover/focus**: El elemento recupera su color real (foto a color, icono oficial de la tecnología, acento de enlace) mediante transición suave.

### Reglas Técnicas

1. **Transición**: `transition: filter 300ms ease, opacity 300ms ease, transform 300ms ease` — sin easing elástico ni efectos de rebote.
2. **Alcance**: Aplica a foto de avatar, iconos de la barra social, iconos de skills, thumbnails de proyectos y enlaces de la navegación.
3. **Reducción de movimiento**: Respetar `prefers-reduced-motion` — desactivar el cambio de filtro suave y aplicar el cambio de color sin animación.
4. **Accesibilidad (teclado)**: El revelado de color también debe dispararse con `:focus-visible`, no solo con `:hover`, para que sea usable sin mouse.
5. **Contraste**: El color revelado nunca debe degradar el contraste del texto; en enlaces, el cambio de color se acompaña de subrayado o cambio de peso.

### Tratamiento por Elemento

| Elemento | Reposo | Hover / Focus |
| :--- | :--- | :--- |
| Avatar vinilo | Imagen en grayscale (anillo exterior gris suave) | Imagen a color, anillos mantienen rotación |
| Iconos sidebar social | Gris `#666` / `#999` | Color de marca del servicio (GitHub, LinkedIn, etc.) |
| Iconos skills grid | Grayscale | Icono oficial a color |
| Thumbnails de proyectos | Grayscale | Foto a color, `scale(1.02)` sutil |
| Enlaces de navegación | `--text-secondary` | Color de acento + subrayado 1px |

**Ejemplo de implementación**:
```css
.grayscale-reveal {
  filter: grayscale(100%);
  transition: filter 300ms ease;
}
.grayscale-reveal:hover,
.grayscale-reveal:focus-visible {
  filter: grayscale(0%);
}
```
