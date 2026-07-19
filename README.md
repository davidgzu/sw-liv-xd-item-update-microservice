# Item Update Microservice

Microservicio en Go para procesar actualizaciones de items desde Google Cloud Pub/Sub usando sqlx y MySQL.

## Estructura del Proyecto

```
.
├── main.go                  # Punto de entrada de la aplicación
├── internal/
│   ├── config/             # Configuración de la aplicación
│   ├── handlers/           # HTTP handlers (Echo)
│   ├── middleware/         # Middleware personalizados
│   ├── models/             # Estructuras de datos
│   ├── requestctx/         # Context de request
│   ├── router/             # Configuración de rutas
│   ├── services/           # Lógica de negocio
│   ├── store/              # Capa de acceso a datos (sqlx)
│   └── utils/              # Utilidades
├── docs/                   # Documentación
├── bin/                    # Binarios compilados
└── package/                # Paquetes adicionales
```

## Configuración

### Variables de Entorno

Copia `.env.example` a `.env` y configura:

```bash
# Server
SERVER_PORT=8080
APP_ENV=dev

# Database - Conexión Local (TCP)
DB_USER=root
DB_PASSWORD=password
DB_NAME=xd_database
DB_HOST=127.0.0.1
DB_PORT=3306

# Database - Conexión Cloud SQL (Unix Socket) para Cloud Run
# Descomentar cuando despliegues en Cloud Run:
# DB_UNIX_SOCKET=/cloudsql/project-id:region:instance-name

# Database Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
```

### Schema de Base de Datos

```sql
CREATE TABLE IF NOT EXISTS item_remision (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    sku VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) DEFAULT 0.00,
    stock INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sku (sku),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## Inicio Rápido

```bash
# 1. Instalar dependencias
go mod download

# 2. Configurar .env
cp .env.example .env
# Editar .env con tus valores

# 3. Crear la tabla en MySQL
mysql -u root -p xd_database < docs/schema.sql

# 4. Ejecutar
go run main.go
```

## Endpoints

| Método | Ruta                   | Descripción                |
|--------|------------------------|----------------------------|
| GET    | `/health`              | Health check               |
| POST   | `/api/v1/items/pubsub` | Recibe mensajes de Pub/Sub |

## Desarrollo

```bash
# Ejecutar en modo desarrollo
go run main.go

# Compilar
go build -o bin/server main.go

# Ejecutar tests
go test ./...
```

## Despliegue en Cloud Run

El proyecto está configurado para funcionar tanto localmente (TCP) como en Cloud Run (Unix Socket).

```bash
# Desplegar en Cloud Run
gcloud run deploy item-update-service \
  --source . \
  --platform managed \
  --region us-central1 \
  --set-env-vars "DB_UNIX_SOCKET=/cloudsql/PROJECT:REGION:INSTANCE" \
  --add-cloudsql-instances PROJECT:REGION:INSTANCE
```

La configuración cambia automáticamente:
- **Local**: Usa `DB_HOST` y `DB_PORT` (TCP)
- **Cloud Run**: Usa `DB_UNIX_SOCKET` (Unix Socket)

## Licencia

[Especificar licencia]
