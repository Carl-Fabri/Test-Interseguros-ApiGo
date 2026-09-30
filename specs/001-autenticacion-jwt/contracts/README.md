# Contratos: Autenticación JWT

Fuente de verdad: [`api/openapi.yaml`](../../../api/openapi.yaml) (servido en `GET /openapi.yaml` y visible en `/docs`).

| Método | Ruta | Seguridad | Respuestas |
|---|---|---|---|
| POST | `/api/v1/auth/login` | Pública | 200 `LoginResponse` · 400 `INVALID_REQUEST` · 401 `INVALID_CREDENTIALS` |
| * | Rutas de negocio (`/api/v1/matrix/*`) | `bearerAuth` (JWT) | 401 `UNAUTHORIZED` si el token falta o es inválido |

```yaml
components:
  securitySchemes:
    bearerAuth: { type: http, scheme: bearer, bearerFormat: JWT }
```

Contrato hacia api-node (consumidor): todas las llamadas a `POST {NODE_API_URL}/api/v1/statistics` llevan el header
`Authorization: Bearer <token del usuario>`.
