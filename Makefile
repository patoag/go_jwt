# Makefile para Go JWT Backend

.PHONY: help build up down logs test clean restart

# Variables
DOCKER_COMPOSE = docker-compose
APP_NAME = go-jwt-backend

# Ayuda por defecto
help:
	@echo "🚀 Go JWT Backend - Comandos disponibles:"
	@echo ""
	@echo "  build     - Construir las imágenes Docker"
	@echo "  up        - Iniciar todos los servicios"
	@echo "  down      - Detener todos los servicios"
	@echo "  restart   - Reiniciar todos los servicios"
	@echo "  logs      - Ver logs de la aplicación"
	@echo "  logs-db   - Ver logs de PostgreSQL"
	@echo "  test      - Ejecutar pruebas de la API"
	@echo "  clean     - Limpiar contenedores y volúmenes"
	@echo "  shell     - Acceder al shell del contenedor de la app"
	@echo "  db-shell  - Acceder al shell de PostgreSQL"
	@echo ""

# Construir imágenes
build:
	@echo "🔨 Construyendo imágenes Docker..."
	$(DOCKER_COMPOSE) build

# Iniciar servicios
up:
	@echo "🚀 Iniciando servicios..."
	$(DOCKER_COMPOSE) up -d --build
	@echo "✅ Servicios iniciados!"
	@echo "📡 API disponible en: http://localhost:8080"
	@echo "🏥 Health check: http://localhost:8080/api/v1/health"

# Iniciar servicios en primer plano
up-fg:
	@echo "🚀 Iniciando servicios en primer plano..."
	$(DOCKER_COMPOSE) up --build

# Detener servicios
down:
	@echo "🛑 Deteniendo servicios..."
	$(DOCKER_COMPOSE) down

# Reiniciar servicios
restart:
	@echo "🔄 Reiniciando servicios..."
	$(DOCKER_COMPOSE) restart

# Ver logs de la aplicación
logs:
	@echo "📋 Logs de la aplicación:"
	$(DOCKER_COMPOSE) logs -f app

# Ver logs de PostgreSQL
logs-db:
	@echo "📋 Logs de PostgreSQL:"
	$(DOCKER_COMPOSE) logs -f postgres

# Ejecutar pruebas
test:
	@echo "🧪 Ejecutando pruebas de la API..."
	@chmod +x test-api.sh
	@./test-api.sh

# Limpiar todo (¡CUIDADO! Elimina datos)
clean:
	@echo "🧹 Limpiando contenedores y volúmenes..."
	$(DOCKER_COMPOSE) down -v --remove-orphans
	@docker system prune -f

# Acceder al shell del contenedor de la app
shell:
	@echo "🐚 Accediendo al shell del contenedor..."
	$(DOCKER_COMPOSE) exec app sh

# Acceder al shell de PostgreSQL
db-shell:
	@echo "🐚 Accediendo a PostgreSQL..."
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -d go_jwt_db

# Ver estado de los servicios
status:
	@echo "📊 Estado de los servicios:"
	$(DOCKER_COMPOSE) ps

# Verificar que la API esté funcionando
health:
	@echo "🏥 Verificando health check..."
	@curl -s http://localhost:8080/api/v1/health | python3 -m json.tool 2>/dev/null || echo "❌ API no disponible"

# Desarrollo: reiniciar solo la app
dev-restart:
	@echo "🔄 Reiniciando solo la aplicación..."
	$(DOCKER_COMPOSE) restart app

# Ver logs en tiempo real
dev-logs:
	@echo "📋 Logs en tiempo real:"
	$(DOCKER_COMPOSE) logs -f

# Backup de la base de datos
backup:
	@echo "💾 Creando backup de la base de datos..."
	@mkdir -p backups
	$(DOCKER_COMPOSE) exec postgres pg_dump -U postgres go_jwt_db > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup creado en backups/"

# Restaurar backup (usar con: make restore BACKUP=archivo.sql)
restore:
	@echo "📥 Restaurando backup: $(BACKUP)"
	@if [ -z "$(BACKUP)" ]; then echo "❌ Especifica el archivo: make restore BACKUP=archivo.sql"; exit 1; fi
	$(DOCKER_COMPOSE) exec -T postgres psql -U postgres go_jwt_db < $(BACKUP)

# Información del proyecto
info:
	@echo "ℹ️  Información del proyecto:"
	@echo "  Nombre: $(APP_NAME)"
	@echo "  Puerto API: 8080"
	@echo "  Puerto DB: 5432"
	@echo "  Base de datos: go_jwt_db"
	@echo "  Usuario admin: admin@example.com / admin123"
