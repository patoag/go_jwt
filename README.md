# Go JWT Backend

API REST de autenticación y gestión de usuarios construida con Go, diseñada como boilerplate profesional para backends con JWT.

## Stack

- **Go 1.23** + **Gin** + **GORM**
- **PostgreSQL** - Base de datos principal
- **Redis** - Token blacklist + Rate limiting
- **Docker Compose** - Orquestación de servicios
- **Swagger UI** - Documentación interactiva

## Quick Start

```bash
# Clonar e iniciar
git clone <repository-url>
cd go-jwt

# Levantar todos los servicios
make up

# Verificar
curl http://localhost:8080/api/v1/health
```

La API estará disponible en `http://localhost:8080` y Swagger UI en `http://localhost:8080/swagger/index.html`.

**Admin por defecto:** `admin@example.com` / `admin123`

## Arquitectura

```
Request → Router → Middleware (CORS, RateLimit, Auth) → Handler → Service → Repository → DB/Redis
```

```
go-jwt/
├── main.go                    # Entry point, router, graceful shutdown
├── config/
│   ├── config.go              # Configuración centralizada (env vars)
│   ├── database.go            # Conexión PostgreSQL + migraciones
│   └── redis.go               # Conexión Redis
├── models/
│   └── user.go                # Modelo User + DTOs
├── repositories/
│   ├── user_repository.go     # Persistencia de usuarios (GORM)
│   └── redis_blacklist.go     # Token blacklist (Redis)
├── services/
│   ├── interfaces.go          # Interfaces para DI
│   ├── errors.go              # Errores de dominio
│   ├── auth_service.go        # Lógica de autenticación
│   └── user_service.go        # Lógica de usuarios
├── handlers/
│   ├── auth.go                # Handlers de auth (register, login, logout, etc.)
│   └── user.go                # Handlers de usuarios (profile, CRUD)
├── middleware/
│   └── auth.go                # Auth, Admin, CORS, RateLimit, Logger
├── utils/
│   ├── auth.go                # JWT, bcrypt
│   ├── validation.go          # Validadores
│   └── response.go            # Helpers de respuesta
├── docs/                      # Swagger (generado)
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── test-api.sh
└── .env.example
```

## Endpoints

| Método | Ruta | Auth | Descripción |
|--------|------|------|-------------|
| `POST` | `/api/v1/auth/register` | No | Registrar usuario |
| `POST` | `/api/v1/auth/login` | No | Iniciar sesión |
| `POST` | `/api/v1/auth/refresh` | Bearer | Renovar token |
| `POST` | `/api/v1/auth/logout` | Bearer | Cerrar sesión (blacklist token) |
| `POST` | `/api/v1/auth/change-password` | Bearer | Cambiar contraseña |
| `GET` | `/api/v1/users/profile` | Bearer | Obtener perfil |
| `PUT` | `/api/v1/users/profile` | Bearer | Actualizar perfil |
| `DELETE` | `/api/v1/users/profile` | Bearer | Eliminar cuenta |
| `GET` | `/api/v1/admin/users` | Admin | Listar usuarios (paginado) |
| `GET` | `/api/v1/admin/users/:id` | Admin | Obtener usuario por ID |
| `GET` | `/api/v1/health` | No | Health check (DB + Redis) |
| `GET` | `/swagger/*` | No | Swagger UI |

## Comandos

```bash
make up             # Iniciar servicios (Docker)
make down           # Detener servicios
make logs           # Ver logs de la app
make test           # Ejecutar todos los tests
make test-unit      # Tests unitarios
make test-api       # Tests de integración
make test-coverage  # Tests con cobertura
make health         # Health check
make shell          # Shell del contenedor
make db-shell       # Shell de PostgreSQL
make swagger        # Regenerar docs Swagger
```

## Variables de Entorno

| Variable | Default | Descripción |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Host de PostgreSQL |
| `DB_PORT` | `5432` | Puerto de PostgreSQL |
| `DB_USER` | `postgres` | Usuario de PostgreSQL |
| `DB_PASSWORD` | `postgres123` | Password de PostgreSQL |
| `DB_NAME` | `go_jwt_db` | Nombre de la base de datos |
| `REDIS_HOST` | `localhost` | Host de Redis |
| `REDIS_PORT` | `6379` | Puerto de Redis |
| `REDIS_PASSWORD` | `` | Password de Redis |
| `REDIS_DB` | `0` | Base de datos de Redis |
| `JWT_SECRET` | `...desarrollo` | Secret para firmar JWT |
| `JWT_EXPIRATION_HOURS` | `24` | Horas de expiración del JWT |
| `PORT` | `8080` | Puerto del servidor |
| `GIN_MODE` | `debug` | Modo de Gin (`debug`/`release`) |
| `CORS_ALLOWED_ORIGINS` | `*` | Orígenes CORS (separados por coma) |
| `RATE_LIMIT_REQUESTS` | `100` | Requests por ventana |
| `RATE_LIMIT_WINDOW` | `60` | Ventana en segundos |

## Testing

```bash
# Tests unitarios (utils, services, middleware)
make test-unit

# Tests de integración via curl (requiere servicios levantados)
make test-api

# Cobertura
make test-coverage
```

## Seguridad

- Passwords hasheados con **bcrypt**
- JWT firmado con **HS256**, expiración configurable
- **Token blacklist** en Redis para logout real
- **Rate limiting** con Redis sliding window por IP (fail-open)
- **CORS** configurable por orígenes
- IDs son **UUID v4**
- Roles: `user` (default) y `admin`

## Despliegue

### Requisitos previos

- **Docker** y **Docker Compose** instalados
- Acceso a un servidor con puertos 8080, 5432 y 6379 disponibles (o configurar los que se deseen)

### Despliegue con Docker Compose

```bash
# 1. Clonar el repositorio
git clone <repository-url>
cd go-jwt

# 2. Crear archivo .env basado en el ejemplo
cp .env.example .env

# 3. Editar .env con valores de producción (ver sección siguiente)
nano .env

# 4. Levantar todos los servicios
make up

# 5. Verificar que todo está corriendo
make health
```

### Variables críticas para producción

Estas variables **deben** cambiarse respecto a los valores por defecto:

```bash
# Seguridad
GIN_MODE=release
JWT_SECRET=<secret-aleatorio-de-al-menos-32-caracteres>
DB_PASSWORD=<password-seguro-de-postgresql>

# CORS - restringir a los dominios permitidos
CORS_ALLOWED_ORIGINS=https://tudominio.com,https://app.tudominio.com

# Rate limiting - ajustar según tráfico esperado
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_WINDOW=60
```

### Despliegue del binario (sin Docker)

```bash
# 1. Compilar el binario
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# 2. Copiar al servidor el binario y la carpeta docs/
scp server docs/ usuario@servidor:/opt/go-jwt/

# 3. En el servidor, configurar variables de entorno y ejecutar
export DB_HOST=<host-postgresql>
export REDIS_HOST=<host-redis>
export JWT_SECRET=<secret-seguro>
export GIN_MODE=release
./server
```

### Notas de producción

- El servidor implementa **graceful shutdown** (señales SIGINT/SIGTERM), lo que permite reiniciar sin perder requests en curso
- Redis opera en modo **fail-open**: si Redis no está disponible, el rate limiting y la blacklist de tokens se desactivan, pero la API sigue funcionando
- El health check (`GET /api/v1/health`) retorna `503` si la base de datos o Redis están caídos, útil para load balancers
- Se recomienda colocar un reverse proxy (nginx, Caddy) delante para TLS y servir en puerto 443
- El admin por defecto (`admin@example.com` / `admin123`) se crea automáticamente al iniciar; cambiar la contraseña inmediatamente en producción

## Licencia

MIT
