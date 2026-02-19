# Usar imagen oficial de Go como base
FROM golang:1.23-alpine AS builder

# Instalar dependencias del sistema
RUN apk add --no-cache git

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod ./

# Descargar dependencias y generar go.sum
RUN go mod download && go mod tidy

# Copiar código fuente
COPY . .

# Obtener todas las dependencias y compilar la aplicación
RUN go get ./... && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Etapa final - imagen mínima
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates

# Crear directorio de trabajo
WORKDIR /root/

# Copiar el binario desde la etapa de construcción
COPY --from=builder /app/main .

# Copiar docs de Swagger
COPY --from=builder /app/docs ./docs

# Exponer puerto
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]
