# ==========================================================================
# portfolio-go — imagen multi-stage
# Build: golang alpine + templ generate
# Runtime: alpine slim + webp (cwebp) para la conversión local de imágenes
# ==========================================================================

# Stage 1: Build
FROM golang:1.25-alpine AS builder

# templ CLI para generar _templ.go en el build (no depender de archivos commiteados)
RUN go install github.com/a-h/templ/cmd/templ@v0.3.906

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generar templates y compilar binario estático
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Stage 2: Runtime
FROM alpine:3.20

# cwebp — requerido por internal/adapters/imageproc (conversión WebP local)
# ca-certificates — para HTTPS a Cloudinary/Resend
RUN apk add --no-cache webp ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/static ./static

# Usuario no-root (buena práctica)
RUN addgroup -S app && adduser -S app -G app
USER app

# Puerto que espera Dokploy (Container Port configurado)
EXPOSE 8080

CMD ["./main"]
