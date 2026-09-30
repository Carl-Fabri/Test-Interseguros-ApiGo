# Quickstart: Autenticación JWT

Validación manual de la funcionalidad (requiere `.env` con `JWT_SECRET`, `AUTH_USERNAME` y `AUTH_PASSWORD`).

```bash
docker compose up --build -d          # o: go run ./cmd/api

# 1. Credenciales válidas → 200 con token
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}'

# 2. Credenciales inválidas → 401 INVALID_CREDENTIALS
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' -d '{"username":"admin","password":"mal"}'

# 3. Ruta protegida sin token → 401 UNAUTHORIZED
curl -s -X POST http://localhost:8080/api/v1/matrix/qr -H 'Content-Type: application/json' -d '{"matrix":[[1]]}'
```

Pruebas automatizadas de esta funcionalidad:
```bash
go test ./internal/auth/... ./internal/service/... ./internal/config/...
go test ./test/integration/ -run 'TestLogin|TestProtectedRoute'
```
