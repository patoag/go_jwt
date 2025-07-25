# Backend RESTful en Go con JWT

Un backend robusto y seguro desarrollado en Go que proporciona gestión completa de usuarios con autenticación JWT, autorización basada en roles y operaciones CRUD completas.

## 🚀 Características

### Funcionalidades Principales
- **Gestión Completa de Usuarios (CRUD)**
  - Registro de usuarios con validación
  - Inicio de sesión con JWT
  - Obtener y actualizar perfil de usuario
  - Eliminación de cuenta
  - Operaciones de administrador (listar usuarios, obtener por ID)

### Seguridad
- **Autenticación JWT** con expiración configurable
- **Hashing de contraseñas** con bcrypt
- **Autorización basada en roles** (user/admin)
- **Validación robusta** de entrada de datos
- **Middleware de seguridad** CORS y rate limiting
- **Manejo consistente de errores** con códigos HTTP apropiados

### Tecnologías
- **Go 1.21** - Lenguaje de programación
- **Gin Gonic** - Framework web de alto rendimiento
- **GORM** - ORM para Go
- **PostgreSQL** - Base de datos
- **JWT** - Autenticación basada en tokens
- **Docker & Docker Compose** - Contenedorización

## 📋 Requisitos

- Docker y Docker Compose instalados
- Puerto 8080 y 5432 disponibles

## 🛠️ Instalación y Configuración

### 1. Clonar el repositorio
```bash
git clone <repository-url>
cd go-jwt-backend
```

### 2. Configurar variables de entorno
```bash
cp .env.example .env
# Editar .env con tus configuraciones si es necesario
```

### 3. Iniciar con Docker Compose
```bash
# Construir e iniciar todos los servicios
docker-compose up --build

# O en modo detached (segundo plano)
docker-compose up -d --build
```

### 4. Verificar que el servidor esté funcionando
```bash
curl http://localhost:8080/api/v1/health
```

## 📚 Documentación de la API

### Base URL
```
http://localhost:8080/api/v1
```

### Autenticación
La API utiliza JWT Bearer tokens. Incluye el token en el header Authorization:
```
Authorization: Bearer <tu-jwt-token>
```

## 🔐 Endpoints de Autenticación

### 1. Registrar Usuario
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@ejemplo.com",
    "nombre_de_usuario": "miusuario",
    "contraseña": "micontraseña123"
  }'
```

**Respuesta exitosa (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "usuario@ejemplo.com",
    "nombre_de_usuario": "miusuario",
    "rol": "user",
    "fecha_creacion": "2024-01-15T10:30:00Z",
    "fecha_actualizacion": "2024-01-15T10:30:00Z"
  }
}
```

### 2. Iniciar Sesión
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@ejemplo.com",
    "contraseña": "micontraseña123"
  }'
```

### 3. Renovar Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Authorization: Bearer <tu-token-actual>"
```

## 👤 Endpoints de Usuario

### 1. Obtener Perfil
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <tu-token>"
```

### 2. Actualizar Perfil
```bash
curl -X PUT http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <tu-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nuevo@ejemplo.com",
    "nombre_de_usuario": "nuevousuario"
  }'
```

### 3. Eliminar Cuenta
```bash
curl -X DELETE http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <tu-token>"
```

## 👑 Endpoints de Administrador

### 1. Listar Usuarios (con paginación)
```bash
curl -X GET "http://localhost:8080/api/v1/admin/users?page=1&page_size=10" \
  -H "Authorization: Bearer <token-admin>"
```

**Respuesta:**
```json
{
  "users": [...],
  "total": 25,
  "page": 1,
  "page_size": 10,
  "total_pages": 3
}
```

### 2. Obtener Usuario por ID
```bash
curl -X GET http://localhost:8080/api/v1/admin/users/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer <token-admin>"
```

## 🗄️ Modelo de Datos

### Usuario
```json
{
  "id": "UUID",
  "email": "string (único, requerido)",
  "nombre_de_usuario": "string (único, requerido, 3-50 chars)",
  "rol": "string (user|admin, default: user)",
  "fecha_creacion": "timestamp",
  "fecha_actualizacion": "timestamp"
}
```

## ⚠️ Códigos de Error

- **400 Bad Request** - Datos inválidos o errores de validación
- **401 Unauthorized** - Token faltante, inválido o expirado
- **403 Forbidden** - Permisos insuficientes
- **404 Not Found** - Recurso no encontrado
- **409 Conflict** - Email o nombre de usuario ya existe
- **500 Internal Server Error** - Error interno del servidor

## 🔧 Desarrollo

### Estructura del Proyecto
```
go-jwt-backend/
├── config/          # Configuración de base de datos
├── handlers/        # Controladores HTTP
├── middleware/      # Middleware de autenticación y autorización
├── models/          # Modelos de datos
├── utils/           # Utilidades (JWT, validación, respuestas)
├── main.go          # Punto de entrada de la aplicación
├── docker-compose.yml
├── Dockerfile
└── README.md
```

### Usuario Administrador por Defecto
Al iniciar la aplicación, se crea automáticamente un usuario administrador:
- **Email:** admin@example.com
- **Contraseña:** admin123
- **Rol:** admin

### Comandos Útiles

```bash
# Ver logs de la aplicación
docker-compose logs -f app

# Ver logs de PostgreSQL
docker-compose logs -f postgres

# Reiniciar solo la aplicación
docker-compose restart app

# Detener todos los servicios
docker-compose down

# Detener y eliminar volúmenes (¡cuidado! elimina datos de BD)
docker-compose down -v
```

## 🧪 Testing

Para probar la API, puedes usar el script de ejemplo incluido:

```bash
# Hacer el script ejecutable
chmod +x test-api.sh

# Ejecutar tests básicos
./test-api.sh
```

## 🚀 Despliegue en Producción

### Variables de Entorno Importantes
```bash
# Cambiar a modo producción
GIN_MODE=release

# Usar un JWT secret más seguro
JWT_SECRET=tu-super-secreto-jwt-muy-largo-y-aleatorio

# Configurar base de datos de producción
DB_HOST=tu-host-de-produccion
DB_PASSWORD=contraseña-segura
```

### Consideraciones de Seguridad
1. Cambiar el JWT_SECRET por uno más seguro
2. Usar HTTPS en producción
3. Configurar rate limiting apropiado
4. Implementar logging y monitoreo
5. Usar variables de entorno para configuración sensible

## 📄 Licencia

Este proyecto está bajo la Licencia MIT.
