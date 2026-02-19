# Makefile para Go JWT Backend

.PHONY: help build up down logs test test-unit test-api test-coverage clean restart swagger

# Variables
DOCKER_COMPOSE = docker-compose
APP_NAME = go-jwt-backend

help:
	@echo "Go JWT Backend - Comandos disponibles:"
	@echo ""
	@echo "  build          - Construir las imagenes Docker"
	@echo "  up             - Iniciar todos los servicios"
	@echo "  up-fg          - Iniciar servicios en primer plano"
	@echo "  down           - Detener todos los servicios"
	@echo "  restart        - Reiniciar todos los servicios"
	@echo "  logs           - Ver logs de la aplicacion"
	@echo "  logs-db        - Ver logs de PostgreSQL"
	@echo "  logs-redis     - Ver logs de Redis"
	@echo "  test           - Ejecutar todos los tests (unit + API)"
	@echo "  test-unit      - Ejecutar tests unitarios"
	@echo "  test-api       - Ejecutar tests de integracion (requiere servicios)"
	@echo "  test-coverage  - Tests con reporte de cobertura"
	@echo "  swagger        - Generar documentacion Swagger"
	@echo "  clean          - Limpiar contenedores y volumenes"
	@echo "  shell          - Acceder al shell del contenedor"
	@echo "  db-shell       - Acceder a PostgreSQL"
	@echo "  health         - Health check del servidor"
	@echo ""

build:
	@echo "Construyendo imagenes Docker..."
	$(DOCKER_COMPOSE) build

up:
	@echo "Iniciando servicios..."
	$(DOCKER_COMPOSE) up -d --build
	@echo "Servicios iniciados!"
	@echo "API disponible en: http://localhost:8080"
	@echo "Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "Health check: http://localhost:8080/api/v1/health"

up-fg:
	@echo "Iniciando servicios en primer plano..."
	$(DOCKER_COMPOSE) up --build

down:
	@echo "Deteniendo servicios..."
	$(DOCKER_COMPOSE) down

restart:
	@echo "Reiniciando servicios..."
	$(DOCKER_COMPOSE) restart

logs:
	$(DOCKER_COMPOSE) logs -f app

logs-db:
	$(DOCKER_COMPOSE) logs -f postgres

logs-redis:
	$(DOCKER_COMPOSE) logs -f redis

test: test-unit test-api

test-unit:
	@echo "Ejecutando tests unitarios..."
	docker run --rm -v $(PWD):/app -w /app golang:1.23-alpine sh -c "apk add --no-cache git && go test ./utils/... ./services/... ./middleware/... -v -count=1"

test-api:
	@echo "Ejecutando tests de integracion..."
	@chmod +x test-api.sh
	@./test-api.sh

test-coverage:
	@echo "Ejecutando tests con cobertura..."
	docker run --rm -v $(PWD):/app -w /app golang:1.23-alpine sh -c "apk add --no-cache git && go test ./utils/... ./services/... ./middleware/... -coverprofile=coverage.out -v && go tool cover -func=coverage.out"

swagger:
	@echo "Generando documentacion Swagger..."
	swag init

clean:
	@echo "Limpiando contenedores y volumenes..."
	$(DOCKER_COMPOSE) down -v --remove-orphans
	@docker system prune -f

shell:
	$(DOCKER_COMPOSE) exec app sh

db-shell:
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -d go_jwt_db

status:
	$(DOCKER_COMPOSE) ps

health:
	@curl -s http://localhost:8080/api/v1/health | python3 -m json.tool 2>/dev/null || echo "API no disponible"

dev-restart:
	$(DOCKER_COMPOSE) restart app

backup:
	@mkdir -p backups
	$(DOCKER_COMPOSE) exec postgres pg_dump -U postgres go_jwt_db > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup creado en backups/"

restore:
	@if [ -z "$(BACKUP)" ]; then echo "Especifica el archivo: make restore BACKUP=archivo.sql"; exit 1; fi
	$(DOCKER_COMPOSE) exec -T postgres psql -U postgres go_jwt_db < $(BACKUP)

info:
	@echo "Informacion del proyecto:"
	@echo "  Nombre: $(APP_NAME)"
	@echo "  Puerto API: 8080"
	@echo "  Puerto DB: 5432"
	@echo "  Puerto Redis: 6379"
	@echo "  Base de datos: go_jwt_db"
	@echo "  Usuario admin: admin@example.com / admin123"
	@echo "  Swagger: http://localhost:8080/swagger/index.html"
