# Data Model: Autenticación JWT

**Feature**: `001-autenticacion-jwt`

No hay persistencia: las entidades existen en memoria o en configuración.

## LoginRequest (DTO de entrada)

| Campo | Tipo | Reglas |
|---|---|---|
| `username` | string | Obligatorio, no vacío (tras recortar espacios) |
| `password` | string | Obligatorio, no vacío |

## LoginResponse (DTO de salida)

| Campo | Tipo | Descripción |
|---|---|---|
| `accessToken` | string | JWT HS256 |
| `tokenType` | string | Siempre `"Bearer"` |
| `expiresIn` | integer | Segundos hasta la expiración |

## AccessToken (servicio)

| Campo | Tipo | Descripción |
|---|---|---|
| `Token` | string | JWT firmado |
| `ExpiresAt` | time | Momento de expiración |

## Claims del JWT

| Claim | Origen | Validación en api-go y api-node |
|---|---|---|
| `sub` | usuario autenticado | — |
| `iss` | `JWT_ISSUER` (`api-go`) | Debe coincidir |
| `aud` | `JWT_AUDIENCE` (`matrix-services`) | Debe coincidir |
| `iat` / `nbf` | momento de emisión | `nbf` ≤ ahora |
| `exp` | `iat + JWT_TTL_MINUTES` | Obligatorio y > ahora |

## Estados del token

```text
emitido ──(uso dentro de la vigencia)──▶ válido ──(pasa exp)──▶ expirado (401)
   └──(firma/alg/iss/aud inválidos)──▶ rechazado (401)
```
