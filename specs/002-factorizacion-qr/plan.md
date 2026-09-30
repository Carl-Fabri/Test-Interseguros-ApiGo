# Implementation Plan: Factorización QR de matrices

**Branch**: `002-factorizacion-qr` | **Date**: 2026-09-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-factorizacion-qr/spec.md`

## Summary

`POST /api/v1/matrix/qr` valida la matriz en el dominio, calcula la QR completa con **rotaciones de Givens**
(`math.Hypot`, ceros exactos y signos normalizados) y, mediante el puerto `StatisticsClient`, envía Q y R a api-node
por HTTP con timeout y el JWT del usuario. El handler solo traduce DTOs; los errores se mapean a HTTP en un único
middleware (400/502/504).

## Technical Context

**Language/Version**: Go 1.27

**Primary Dependencies**: Fiber v3 (HTTP), biblioteca estándar (`math`, `net/http`, `encoding/json`)

**Storage**: N/A

**Testing**: `go test` table-driven y por propiedades; integración con `httptest` simulando api-node

**Target Platform**: contenedor Linux distroless (Docker / AWS ECS Fargate)

**Project Type**: web-service (API REST)

**Performance Goals**: 100×100 en < 50 ms; costo O(m·n·min(m,n)) para R y O(m²·min(m,n)) para Q

**Constraints**: entrada ≤ 100×100 (configurable), cuerpo ≤ 1 MiB, timeout a api-node de 5 s (configurable)

**Scale/Scope**: una petición = una matriz; sin estado

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Cumplimiento |
|---|---|
| I. Servicio autónomo | ✅ api-node se consume solo por HTTP (`NODE_API_URL`) según su contrato |
| II. Contrato primero | ✅ `QRRequest`, `QRResponse` y errores 400/401/502/504 en `api/openapi.yaml` |
| III. Dominio puro | ✅ `domain.QRGivens` y `Validate` sin dependencias; `MatrixService` usa el puerto `StatisticsClient` |
| IV. Pruebas obligatorias | ✅ propiedades de la QR, resultado conocido, servicio con dobles y app completa con api-node simulado |
| V. Seguridad por defecto | ✅ JWT obligatorio, límites de tamaño y de cuerpo, errores internos ocultos |
| VI. Simplicidad | ✅ sin bibliotecas de álgebra externas: el algoritmo cabe en ~40 líneas auditables |

**Resultado**: aprobado sin violaciones (re-verificado tras la fase 1).

## Project Structure

### Documentation (this feature)

```text
specs/002-factorizacion-qr/
├── plan.md · research.md · data-model.md · quickstart.md
├── contracts/
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── domain/          matrix.go (Matrix, Validate, Clone, Transpose, Identity) · qr.go (QRGivens) · *_test.go
├── service/         matrix_service.go (Analyze) · ports.go (StatisticsClient, ErrStatistics*) · service_test.go
├── client/          statistics_client.go (HTTP a api-node con timeout y Bearer) · statistics_client_test.go
├── handler/         matrix_handler.go (POST /api/v1/matrix/qr)
├── model/           dto.go (QRRequest, QRResponse, StatisticsDTO)
├── middleware/      error_handler.go (INVALID_MATRIX, STATISTICS_UNAVAILABLE, STATISTICS_TIMEOUT)
└── server/          server.go (ruta protegida)
test/integration/    matrix_test.go (camino feliz, entradas inválidas, fallas de api-node)
api/                 openapi.yaml · docs.html (Scalar)
```

**Structure Decision**: servicio único con Clean Architecture. La matemática vive en `internal/domain` sin
dependencias para poder probarla por propiedades de forma aislada.

## Complexity Tracking

Sin violaciones de la constitución.
