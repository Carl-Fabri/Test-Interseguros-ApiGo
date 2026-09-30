# Implementation Plan: Autenticación JWT

**Branch**: `001-autenticacion-jwt` | **Date**: 2026-09-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-autenticacion-jwt/spec.md`

## Summary

api-go emite tokens JWT HS256 en `POST /api/v1/auth/login` para un usuario de demostración configurado por entorno,
protege sus endpoints de negocio con un middleware que valida firma, algoritmo, emisor, audiencia y expiración, y
propaga el token del usuario a api-node a través de `context.Context`. Se usa `golang-jwt/jwt/v5` (estándar de facto
en Go) y un caso de uso `AuthService` independiente de HTTP.

## Technical Context

**Language/Version**: Go 1.27

**Primary Dependencies**: Fiber v3, `github.com/golang-jwt/jwt/v5`

**Storage**: N/A (credenciales en variables de entorno)

**Testing**: `go test` (table-driven), `httptest`, `app.Test` de Fiber

**Target Platform**: contenedor Linux (distroless) en Docker local y AWS ECS Fargate

**Project Type**: web-service (API REST)

**Performance Goals**: login < 50 ms; validación del token despreciable frente al cálculo de la QR

**Constraints**: secreto ≥ 32 caracteres; sin estado de sesión en el servidor (stateless)

**Scale/Scope**: un usuario de demostración; tokens de 60 minutos

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Cumplimiento |
|---|---|
| I. Servicio autónomo | ✅ api-node solo recibe el token por HTTP; no hay código compartido |
| II. Contrato primero | ✅ `/auth/login`, `bearerAuth` y errores definidos en `api/openapi.yaml` antes de implementar |
| III. Dominio puro | ✅ `AuthService` y el puerto `TokenIssuer` no importan Fiber; el JWT vive en el adaptador `internal/auth` |
| IV. Pruebas obligatorias | ✅ unitarias (JWTManager, AuthService, config) e integración (login y rutas protegidas) |
| V. Seguridad por defecto | ✅ algoritmo fijo, `exp` obligatorio, tiempo constante, arranque bloqueado con secreto débil |
| VI. Simplicidad | ✅ una dependencia estándar; sin base de datos ni refresh tokens |

**Resultado**: aprobado sin violaciones (re-verificado tras la fase 1).

## Project Structure

### Documentation (this feature)

```text
specs/001-autenticacion-jwt/
├── plan.md              # Este archivo
├── research.md          # Fase 0: decisiones y alternativas
├── data-model.md        # Fase 1: entidades y claims
├── quickstart.md        # Fase 1: validación manual
├── contracts/           # Fase 1: contrato de los endpoints
├── checklists/          # Calidad de la especificación
└── tasks.md             # Fase 2: tareas
```

### Source Code (repository root)

```text
internal/
├── auth/            jwt.go (JWTManager: Issue/Verify, token en context) · jwt_test.go
├── service/         auth_service.go (Login en tiempo constante) · ports.go (TokenIssuer, ErrInvalidCredentials)
├── handler/         auth_handler.go (POST /api/v1/auth/login)
├── middleware/      jwt.go (RequireJWT) · error_handler.go (UNAUTHORIZED / INVALID_CREDENTIALS)
├── model/           dto.go (LoginRequest, LoginResponse)
├── config/          config.go (JWT_*, AUTH_* obligatorios)
└── client/          statistics_client.go (reenvía el Bearer a api-node)
test/integration/    auth_test.go (login y rutas protegidas)
api/openapi.yaml     contrato
```

**Structure Decision**: servicio único con Clean Architecture (ver constitución, principio III). La emisión y
validación del JWT es un adaptador (`internal/auth`) detrás del puerto `TokenIssuer`, de modo que cambiar a un
proveedor de identidad externo no toca los casos de uso.

## Complexity Tracking

Sin violaciones de la constitución.
