# Quickstart: Factorización QR de matrices

Requiere api-go y api-node levantados con el mismo `JWT_SECRET`.

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}' | jq -r .accessToken)

# 1. Ejemplo clásico → R = [[14,21,-14],[0,175,-70],[0,0,35]]
curl -s -X POST http://localhost:8080/api/v1/matrix/qr -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"matrix":[[12,-51,4],[6,167,-68],[-4,24,-41]]}'

# 2. Matriz alta 3×2 → Q 3×3, R 3×2
curl -s -X POST http://localhost:8080/api/v1/matrix/qr -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"matrix":[[1,2],[3,4],[5,6]]}'

# 3. Matriz no rectangular → 400 INVALID_MATRIX
curl -s -X POST http://localhost:8080/api/v1/matrix/qr -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"matrix":[[1,2],[3]]}'

# 4. Con api-node detenido → 502 STATISTICS_UNAVAILABLE
```

Pruebas automatizadas:
```bash
go test ./internal/domain/... ./internal/service/... ./internal/client/...
go test ./test/integration/ -run 'TestQR'
```
