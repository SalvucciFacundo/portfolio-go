# ==========================================================================
# portfolio-go — imagen multi-stage
# Build: golang alpine — los _templ.go ya están generados y commiteados
#        (templ generate se corre en dev, no en el build de Docker)
# Runtime: alpine slim + libwebp-tools (cwebp) para conversión local de imágenes
# ==========================================================================

# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilar binario estático (los _templ.go ya están en el repo)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Stage 2: Runtime
FROM alpine:3.20

# libwebp-tools — provee el binario cwebp requerido por internal/adapters/imageproc
# (conversión WebP local). ca-certificates para HTTPS a Cloudinary/Resend.
RUN apk add --no-cache libwebp-tools ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/static ./static

# Usuario no-root (buena práctica)
RUN addgroup -S app && adduser -S app -G app
USER app

# Puerto que espera Dokploy (Container Port configurado — el server escucha
# en el puerto de SERVER_PORT del environment, default 8080)
EXPOSE 3000

CMD ["./main"]
