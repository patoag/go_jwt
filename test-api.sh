#!/bin/bash

# Script de prueba para la API REST de Go JWT Backend
# Asegúrate de que el servidor esté ejecutándose en localhost:8080

BASE_URL="http://localhost:8080/api/v1"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="admin123"
TIMESTAMP=$(date +%s)
TEST_EMAIL="test${TIMESTAMP}@ejemplo.com"
TEST_USERNAME="testuser${TIMESTAMP}"
PASS=0
FAIL=0

echo "Iniciando pruebas de la API REST Go JWT Backend"
echo "=================================================="

# Función para hacer peticiones con formato bonito
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local auth_header=$4

    echo ""
    echo ">> $method $url"

    if [ ! -z "$auth_header" ]; then
        if [ ! -z "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X $method "$url" -H "Content-Type: application/json" -H "$auth_header" -d "$data")
        else
            response=$(curl -s -w "\n%{http_code}" -X $method "$url" -H "Content-Type: application/json" -H "$auth_header")
        fi
    else
        if [ ! -z "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X $method "$url" -H "Content-Type: application/json" -d "$data")
        else
            response=$(curl -s -w "\n%{http_code}" -X $method "$url" -H "Content-Type: application/json")
        fi
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    echo "   Status: $http_code"
    echo "   Body: $(echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body")"
    echo "$body"
}

assert_status() {
    local expected=$1
    local actual=$2
    local test_name=$3

    if [ "$actual" = "$expected" ]; then
        echo "   [PASS] $test_name (status $actual)"
        PASS=$((PASS + 1))
    else
        echo "   [FAIL] $test_name (esperado $expected, obtenido $actual)"
        FAIL=$((FAIL + 1))
    fi
}

# 1. Health Check
echo ""
echo "=== 1. Health Check ==="
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
http_code=$(echo "$response" | tail -n1)
assert_status "200" "$http_code" "Health check"

# 2. Registrar nuevo usuario
echo ""
echo "=== 2. Registrar nuevo usuario ==="
USER_DATA='{"email": "'$TEST_EMAIL'", "nombre_de_usuario": "'$TEST_USERNAME'", "contraseña": "password123"}'
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" -H "Content-Type: application/json" -d "$USER_DATA")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
assert_status "201" "$http_code" "Registrar usuario"
USER_TOKEN=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('token', ''))" 2>/dev/null)

# 3. Login con usuario recién creado
echo ""
echo "=== 3. Login ==="
LOGIN_DATA='{"email": "'$TEST_EMAIL'", "contraseña": "password123"}'
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" -H "Content-Type: application/json" -d "$LOGIN_DATA")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
assert_status "200" "$http_code" "Login usuario"
USER_TOKEN=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('token', ''))" 2>/dev/null)

# 4. Login como administrador
echo ""
echo "=== 4. Login admin ==="
ADMIN_LOGIN_DATA='{"email": "'$ADMIN_EMAIL'", "contraseña": "'$ADMIN_PASSWORD'"}'
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" -H "Content-Type: application/json" -d "$ADMIN_LOGIN_DATA")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
assert_status "200" "$http_code" "Login admin"
ADMIN_TOKEN=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('token', ''))" 2>/dev/null)

if [ ! -z "$USER_TOKEN" ]; then
    # 5. Obtener perfil
    echo ""
    echo "=== 5. Obtener perfil ==="
    response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/users/profile" -H "Authorization: Bearer $USER_TOKEN")
    http_code=$(echo "$response" | tail -n1)
    assert_status "200" "$http_code" "Obtener perfil"

    # 6. Actualizar perfil
    echo ""
    echo "=== 6. Actualizar perfil ==="
    UPDATE_DATA='{"nombre_de_usuario": "'${TEST_USERNAME}_upd'"}'
    response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/users/profile" -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" -d "$UPDATE_DATA")
    http_code=$(echo "$response" | tail -n1)
    assert_status "200" "$http_code" "Actualizar perfil"

    # 7. Renovar token
    echo ""
    echo "=== 7. Renovar token ==="
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/refresh" -H "Authorization: Bearer $USER_TOKEN")
    http_code=$(echo "$response" | tail -n1)
    assert_status "200" "$http_code" "Renovar token"

    # 8. Cambiar contraseña
    echo ""
    echo "=== 8. Cambiar contraseña ==="
    CHANGE_PWD_DATA='{"contraseña_actual": "password123", "nueva_contraseña": "newpassword456"}'
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/change-password" -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" -d "$CHANGE_PWD_DATA")
    http_code=$(echo "$response" | tail -n1)
    assert_status "200" "$http_code" "Cambiar contraseña"

    # 9. Login con nueva contraseña
    echo ""
    echo "=== 9. Login con nueva contraseña ==="
    NEW_LOGIN_DATA='{"email": "'$TEST_EMAIL'", "contraseña": "newpassword456"}'
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" -H "Content-Type: application/json" -d "$NEW_LOGIN_DATA")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    assert_status "200" "$http_code" "Login nueva contraseña"
    NEW_TOKEN=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('token', ''))" 2>/dev/null)

    # 10. Logout
    echo ""
    echo "=== 10. Logout ==="
    if [ ! -z "$NEW_TOKEN" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/logout" -H "Authorization: Bearer $NEW_TOKEN")
        http_code=$(echo "$response" | tail -n1)
        assert_status "200" "$http_code" "Logout"

        # 11. Verificar que token ya no funciona después del logout
        echo ""
        echo "=== 11. Token revocado post-logout ==="
        response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/users/profile" -H "Authorization: Bearer $NEW_TOKEN")
        http_code=$(echo "$response" | tail -n1)
        assert_status "401" "$http_code" "Token revocado"
    fi
fi

if [ ! -z "$ADMIN_TOKEN" ]; then
    # 12. Listar usuarios (Admin)
    echo ""
    echo "=== 12. Listar usuarios (Admin) ==="
    response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/admin/users?page=1&page_size=5" -H "Authorization: Bearer $ADMIN_TOKEN")
    http_code=$(echo "$response" | tail -n1)
    assert_status "200" "$http_code" "Listar usuarios"

    # 13. Obtener usuario por ID (Admin)
    echo ""
    echo "=== 13. Obtener usuario por ID (Admin) ==="
    users_response=$(curl -s -X GET "$BASE_URL/admin/users?page=1&page_size=5" -H "Authorization: Bearer $ADMIN_TOKEN")
    USER_ID=$(echo "$users_response" | python3 -c "import sys, json; data=json.load(sys.stdin); users=data.get('users', []); print(users[0]['id'] if users else '')" 2>/dev/null)
    if [ ! -z "$USER_ID" ]; then
        response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/admin/users/$USER_ID" -H "Authorization: Bearer $ADMIN_TOKEN")
        http_code=$(echo "$response" | tail -n1)
        assert_status "200" "$http_code" "Obtener usuario por ID"
    fi
fi

# 14. Acceso sin autorización
echo ""
echo "=== 14. Acceso sin autorización ==="
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/users/profile")
http_code=$(echo "$response" | tail -n1)
assert_status "401" "$http_code" "Sin autorizacion"

# 15. Usuario normal intenta acceder a ruta admin
if [ ! -z "$USER_TOKEN" ]; then
    echo ""
    echo "=== 15. Usuario normal a ruta admin ==="
    response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/admin/users" -H "Authorization: Bearer $USER_TOKEN")
    http_code=$(echo "$response" | tail -n1)
    assert_status "403" "$http_code" "Acceso denegado a admin"
fi

# Resumen
echo ""
echo "=================================================="
echo "Resultados: $PASS pasaron, $FAIL fallaron"
echo "=================================================="

if [ $FAIL -gt 0 ]; then
    exit 1
fi
