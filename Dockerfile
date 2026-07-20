# Build stage
FROM golang:1.25-alpine AS builder

# Instalar dependencias necesarias para compilación
RUN apk add --no-cache git ca-certificates tzdata

# Crear directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar la aplicación
# CGO_ENABLED=0 para crear un binario estático
# -ldflags "-w -s" para reducir tamaño del binario
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server \
    main.go

# Runtime stage
FROM alpine:latest

# Instalar ca-certificates para conexiones HTTPS y tzdata para zonas horarias
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root para seguridad
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Crear directorio de trabajo
WORKDIR /app

# Copiar el binario compilado desde el build stage
COPY --from=builder /app/server /app/server

# Copiar archivo de credenciales de Firebase (si existe)
# En producción, esto debería manejarse con secrets de Cloud Run
COPY --chown=appuser:appuser crp-qas-log-firebase.json /app/crp-qas-log-firebase.json

# Cambiar al usuario no-root
USER appuser

# Exponer puerto (Cloud Run usa la variable PORT)
# Por defecto 8080, pero Cloud Run lo sobrescribirá
EXPOSE 8080

# Variables de entorno por defecto
ENV SERVER_PORT=8080 \
    APP_ENV=production \
    GOOGLE_APPLICATION_CREDENTIALS=/app/crp-qas-log-firebase.json

# Health check (opcional, útil para debugging)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${SERVER_PORT}/health || exit 1

# Comando para iniciar la aplicación
ENTRYPOINT ["/app/server"]
