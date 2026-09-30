# Contratos: Factorización QR

## Endpoint expuesto por api-go

Fuente de verdad: [`api/openapi.yaml`](../../../api/openapi.yaml) (en vivo: `GET /openapi.yaml`, referencia en `/docs`).

| Método | Ruta | Seguridad | Respuestas |
|---|---|---|---|
| POST | `/api/v1/matrix/qr` | `bearerAuth` | 200 `QRResponse` · 400 `INVALID_REQUEST`/`INVALID_MATRIX` · 401 `UNAUTHORIZED` · 502 `STATISTICS_UNAVAILABLE` · 504 `STATISTICS_TIMEOUT` |

## Endpoint consumido de api-node

Definido por api-node en su propio `openapi/openapi.yaml` (en vivo: `{NODE_API_URL}/openapi.yaml`).

```http
POST {NODE_API_URL}/api/v1/statistics
Authorization: Bearer <token del usuario>
Content-Type: application/json

{ "matrices": [ { "name": "Q", "values": [[...]] }, { "name": "R", "values": [[...]] } ] }
```

Respuesta esperada (200): `{ max, min, average, sum, count, anyDiagonal, matrices: [{ name, isDiagonal }] }`.
Cualquier otro estado se trata como `STATISTICS_UNAVAILABLE` (502).
