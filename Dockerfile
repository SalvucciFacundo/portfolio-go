# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 = binario estático que corre en alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

# Puerto que espera Dokploy (Container Port configurado)
EXPOSE 3000

CMD ["./main"]
