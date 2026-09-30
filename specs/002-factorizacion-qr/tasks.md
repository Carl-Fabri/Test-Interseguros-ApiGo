---

description: "Lista de tareas de la funcionalidad Factorización QR"
---

# Tasks: Factorización QR de matrices

**Input**: Design documents from `/specs/002-factorizacion-qr/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ · depende de `001-autenticacion-jwt`

**Tests**: obligatorios por constitución (principio IV); se escriben antes de la implementación de cada historia.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 Definir `QRRequest`, `QRResponse`, `Statistics` y errores 400/502/504 en api/openapi.yaml
- [x] T002 [P] Agregar `NODE_API_URL`, `NODE_API_TIMEOUT_MS` y `MATRIX_MAX_DIMENSION` a internal/config/config.go y .env.example
- [x] T003 [P] Servir el contrato y Scalar (`/openapi.yaml`, `/docs`) embebidos con go:embed en api/api.go

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T004 Crear el tipo Matrix con Clone, Transpose, Zeros e Identity en internal/domain/matrix.go
- [x] T005 Definir el puerto StatisticsClient, NamedMatrix, Statistics y los errores ErrStatistics* en internal/service/ports.go

**Checkpoint**: tipos de dominio y puertos listos.

---

## Phase 3: User Story 1 - Factorizar una matriz rectangular (Priority: P1) 🎯 MVP

**Goal**: A = Q·R con Givens para cualquier forma.

**Independent Test**: el ejemplo clásico devuelve la R conocida y Q ortogonal.

### Tests for User Story 1

- [x] T006 [P] [US1] Pruebas por propiedades (Q·R ≈ A, QᵀQ ≈ I, R triangular, diag ≥ 0, entrada intacta) para cuadrada, alta, ancha, singular, 1×1, fila/columna única y valores 1e150 en internal/domain/qr_test.go
- [x] T007 [P] [US1] Prueba con resultado conocido y matriz diagonal en internal/domain/qr_test.go

### Implementation for User Story 1

- [x] T008 [US1] Implementar QRGivens (hypot, ceros exactos, omitir ceros existentes) en internal/domain/qr.go
- [x] T009 [US1] Normalizar signos (diag(R) ≥ 0) y limpiar -0 en internal/domain/qr.go

**Checkpoint**: el núcleo matemático está verificado de forma aislada.

---

## Phase 4: User Story 2 - Validar la entrada (Priority: P1)

**Goal**: rechazar entradas inválidas con 400 antes de calcular.

**Independent Test**: vacía, no rectangular, no finita o excesiva → 400 sin llamar a api-node.

### Tests for User Story 2

- [x] T010 [P] [US2] Pruebas table-driven de Validate en internal/domain/matrix_test.go
- [x] T011 [P] [US2] Pruebas de integración de entradas inválidas en test/integration/matrix_test.go

### Implementation for User Story 2

- [x] T012 [US2] Implementar Validate con ValidationError en internal/domain/matrix.go
- [x] T013 [US2] Mapear ValidationError a 400 INVALID_MATRIX y errores de bind a INVALID_REQUEST en internal/middleware/error_handler.go
- [x] T014 [P] [US2] Limitar el cuerpo a 1 MiB en internal/server/server.go

---

## Phase 5: User Story 3 - Obtener estadísticas de Q y R desde api-node (Priority: P2)

**Goal**: enviar Q y R a api-node y devolver sus estadísticas.

**Independent Test**: con api-node simulado la respuesta incluye `statistics`; con fallas, 502/504.

### Tests for User Story 3

- [x] T015 [P] [US3] Pruebas del cliente (éxito, error HTTP, JSON inválido, timeout, conexión rechazada) en internal/client/statistics_client_test.go
- [x] T016 [P] [US3] Pruebas del caso de uso con doble del puerto en internal/service/service_test.go
- [x] T017 [P] [US3] Pruebas de integración (camino feliz, 502, 504) en test/integration/matrix_test.go

### Implementation for User Story 3

- [x] T018 [US3] Implementar StatisticsHTTPClient con timeout y Bearer en internal/client/statistics_client.go
- [x] T019 [US3] Implementar MatrixService.Analyze (validar → QR → estadísticas) en internal/service/matrix_service.go
- [x] T020 [P] [US3] Crear los DTOs y el mapeo de respuesta en internal/model/dto.go e internal/handler/matrix_handler.go
- [x] T021 [US3] Mapear ErrStatisticsUnavailable → 502 y ErrStatisticsTimeout → 504 en internal/middleware/error_handler.go
- [x] T022 [US3] Registrar `POST /api/v1/matrix/qr` con RequireJWT en internal/server/server.go

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T023 [P] Documentar algoritmo, errores y ejemplos en README.md, .claude/CLAUDE.md y docs/architecture/ADR-002
- [x] T024 [P] Dockerfile multi-stage distroless y docker-compose.yml con `DOCKER_NODE_API_URL`
- [x] T025 [P] Pila de despliegue (NODE_API_URL, health check del ALB) en deploy/aws/service.yml
- [x] T026 Ejecutar quickstart.md y `go test -coverpkg=./internal/... ./...` (cobertura ≈ 97 %)

## Dependencies & Execution Order

- Setup → Foundational → US1 → US2 → US3 → Polish.
- US3 depende de US1 (necesita Q y R) y de `001-autenticacion-jwt` (propagación del token).
