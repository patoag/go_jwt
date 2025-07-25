#!/bin/bash

# Script de prueba para la API REST de Go JWT Backend
# Asegúrate de que el servidor esté ejecutándose en localhost:8080

BASE_URL="http://localhost:8080/api/v1"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="admin123"

echo "🚀 Iniciando pruebas de la API REST Go JWT Backend"
echo "=================================================="

# Función para hacer peticiones con formato bonito
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local auth_header=$4

    echo ""
    echo "📡 $method $url"
    if [ ! -z "$data" ]; then
        echo "📤 Datos: $data"
    fi

    if [ ! -z "$auth_header" ]; then
        if [ ! -z "$data" ]; then
            response=$(curl -s -X $method "$url" -H "Content-Type: application/json" -H "$auth_header" -d "$data")
        else
            response=$(curl -s -X $method "$url" -H "Content-Type: application/json" -H "$auth_header")
        fi
    else
        if [ ! -z "$data" ]; then
            response=$(curl -s -X $method "$url" -H "Content-Type: application/json" -d "$data")
        else
            response=$(curl -s -X $method "$url" -H "Content-Type: application/json")
        fi
    fi

    echo "📥 Respuesta:"
    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    echo ""
}

# 1. Health Check
echo "1️⃣ Health Check"
make_request "GET" "$BASE_URL/health"

# 2. Registrar nuevo usuario
echo "2️⃣ Registrar nuevo usuario"
USER_DATA='{
    "email": "test@ejemplo.com",
    "nombre_de_usuario": "testuser",
    "contraseña": "password123"
}'
register_response=$(curl -s -X POST "$BASE_URL/auth/register" -H "Content-Type: application/json" -d "$USER_DATA")
echo "📥 Respuesta de registro:"
echo "$register_response" | python3 -m json.tool 2>/dev/null || echo "$register_response"

# Extraer token del usuario registrado
USER_TOKEN=$(echo "$register_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('token', ''))" 2>/dev/null)

# 3. Login con usuario recién creado
echo ""
echo "3️⃣ Login con usuario recién creado"
LOGIN_DATA='{
    "email": "test@ejemplo.com",
    "contraseña": "password123"
}'
make_request "POST" "$BASE_URL/auth/login" "$LOGIN_DATA"

# 4. Login como administrador
echo "4️⃣ Login como administrador"
ADMIN_LOGIN_DATA='{
    "email": "'$ADMIN_EMAIL'",
    "contraseña": "'$ADMIN_PASSWORD'"
}'
admin_response=$(curl -s -X POST "$BASE_URL/auth/login" -H "Content-Type: application/json" -d "$ADMIN_LOGIN_DATA")
echo "📥 Respuesta de login admin:"
echo "$admin_response" | python3 -m json.tool 2>/dev/null || echo "$admin_response"

# Extraer token del admin
ADMIN_TOKEN=$(echo "$admin_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('token', ''))" 2>/dev/null)

if [ ! -z "$USER_TOKEN" ]; then
    echo ""
    echo "5️⃣ Obtener perfil de usuario"
    make_request "GET" "$BASE_URL/users/profile" "" "Authorization: Bearer $USER_TOKEN"

    echo "6️⃣ Actualizar perfil de usuario"
    UPDATE_DATA='{
        "nombre_de_usuario": "testuser_updated"
    }'
    make_request "PUT" "$BASE_URL/users/profile" "$UPDATE_DATA" "Authorization: Bearer $USER_TOKEN"

    echo "7️⃣ Renovar token"
    make_request "POST" "$BASE_URL/auth/refresh" "" "Authorization: Bearer $USER_TOKEN"
fi

if [ ! -z "$ADMIN_TOKEN" ]; then
    echo ""
    echo "8️⃣ Listar usuarios (Admin)"
    make_request "GET" "$BASE_URL/admin/users?page=1&page_size=5" "" "Authorization: Bearer $ADMIN_TOKEN"

    # Obtener ID del usuario para la siguiente prueba
    users_response=$(curl -s -X GET "$BASE_URL/admin/users?page=1&page_size=5" -H "Authorization: Bearer $ADMIN_TOKEN")
    USER_ID=$(echo "$users_response" | python3 -c "import sys, json; data=json.load(sys.stdin); users=data.get('users', []); print(users[0]['id'] if users else '')" 2>/dev/null)

    if [ ! -z "$USER_ID" ]; then
        echo "9️⃣ Obtener usuario por ID (Admin)"
        make_request "GET" "$BASE_URL/admin/users/$USER_ID" "" "Authorization: Bearer $ADMIN_TOKEN"
    fi
fi

# 10. Probar acceso sin autorización
echo "🔒 Probar acceso sin autorización"
make_request "GET" "$BASE_URL/users/profile"

# 11. Probar acceso de usuario normal a ruta de admin
if [ ! -z "$USER_TOKEN" ]; then
    echo "🚫 Probar acceso de usuario normal a ruta de admin"
    make_request "GET" "$BASE_URL/admin/users" "" "Authorization: Bearer $USER_TOKEN"
fi

echo ""
echo "✅ Pruebas completadas!"
echo ""
echo "📝 Notas:"
echo "- Si ves errores de 'command not found' para python3, instala Python 3"
echo "- Los tokens JWT expiran en 24 horas"
echo "- El usuario admin por defecto es: $ADMIN_EMAIL / $ADMIN_PASSWORD"
echo "- Para eliminar el usuario de prueba, usa: DELETE /api/v1/users/profile con el token del usuario"
