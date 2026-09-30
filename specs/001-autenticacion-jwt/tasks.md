---

description: "Lista de tareas de la funcionalidad Autenticación JWT"
---

# Tasks: Autenticación JWT

**Input**: Design documents from `/specs/001-autenticacion-jwt/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: obligatorios por constitución (principio IV); se escriben antes de la implementación de cada historia.

**Organization**: tareas agrupadas por historia de usuario.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: se puede ejecutar en paralelo (archivos distintos, sin dependencias)
- **[Story]**: historia de usuario (US1, US2, US3)

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 Agregar la dependencia `github.com/golang-jwt/jwt/v5` en go.mod
- [x] T002 [P] Definir `/api/v1/auth/login`, `bearerAuth` y errores de autenticación en api/openapi.yaml
- [x] T003 [P] Agregar `JWT_SECRET`, `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_TTL_MINUTES`, `AUTH_USERNAME`, `AUTH_PASSWORD` a .env.example

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T004 Validar variables obligatorias y secreto ≥ 32 caracteres en internal/config/config.go
- [x] T005 [P] Pruebas table-driven de configuración en internal/config/config_test.go
- [x] T006 Implementar el formato único de error y el mapeo 401 en internal/middleware/error_handler.go

**Checkpoint**: configuración y errores listos.

---

## Phase 3: User Story 1 - Obtener un token de acceso (Priority: P1) 🎯 MVP

**Goal**: emitir un JWT HS256 para credenciales válidas.

**Independent Test**: `POST /api/v1/auth/login` devuelve 200 con token o 401 con credenciales incorrectas.

### Tests for User Story 1

- [x] T007 [P] [US1] Pruebas de emisión y verificación (expirado, otro secreto, otra audiencia, `alg: none`, mal formado) en internal/auth/jwt_test.go
- [x] T008 [P] [US1] Pruebas de AuthService (correctas, usuario/contraseña incorrectos, vacías) en internal/service/service_test.go
- [x] T009 [P] [US1] Pruebas de integración del login (200, 401, 400) en test/integration/auth_test.go

### Implementation for User Story 1

- [x] T010 [P] [US1] Implementar JWTManager (Issue/Verify con HS256 fijo, iss, aud, exp obligatorio) en internal/auth/jwt.go
- [x] T011 [US1] Definir el puerto TokenIssuer y ErrInvalidCredentials en internal/service/ports.go
- [x] T012 [US1] Implementar AuthService.Login con comparación en tiempo constante en internal/service/auth_service.go
- [x] T013 [P] [US1] Crear LoginRequest y LoginResponse en internal/model/dto.go
- [x] T014 [US1] Implementar AuthHandler.Login en internal/handler/auth_handler.go
- [x] T015 [US1] Registrar la ruta pública en internal/server/server.go

**Checkpoint**: el login funciona de forma independiente.

---

## Phase 4: User Story 2 - Proteger los endpoints de negocio (Priority: P1)

**Goal**: solo clientes con token vigente acceden a las rutas de negocio.

**Independent Test**: rutas protegidas responden 401 sin token o con token inválido y no llaman a api-node.

### Tests for User Story 2

- [x] T016 [P] [US2] Pruebas de ruta protegida sin token y con token inválido en test/integration/auth_test.go
- [x] T017 [P] [US2] Pruebas del mapeo de errores 401 en internal/middleware/error_handler_test.go

### Implementation for User Story 2

- [x] T018 [US2] Implementar el middleware RequireJWT (esquema Bearer, verificación y subject en Locals) en internal/middleware/jwt.go
- [x] T019 [US2] Aplicar RequireJWT a `/api/v1/matrix/qr` en internal/server/server.go

**Checkpoint**: todas las rutas de negocio exigen JWT.

---

## Phase 5: User Story 3 - Propagar la identidad a api-node (Priority: P2)

**Goal**: api-go reenvía el token del usuario a api-node.

**Independent Test**: el api-node simulado recibe `Authorization: Bearer <token>`.

### Tests for User Story 3

- [x] T020 [P] [US3] Prueba de contexto con token en internal/auth/jwt_test.go (TestTokenContext)
- [x] T021 [P] [US3] Prueba de reenvío del header en internal/client/statistics_client_test.go y test/integration/matrix_test.go

### Implementation for User Story 3

- [x] T022 [US3] Guardar el token en context.Context desde RequireJWT (auth.WithToken)
- [x] T023 [US3] Leer el token del contexto y enviarlo en internal/client/statistics_client.go

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T024 [P] Documentar autenticación, variables y errores en README.md y .claude/CLAUDE.md
- [x] T025 [P] Inyectar JWT_SECRET y AUTH_PASSWORD desde Secrets Manager en deploy/aws/service.yml
- [x] T026 Ejecutar quickstart.md y la suite completa (`go test ./...`)

## Dependencies & Execution Order

- Setup → Foundational → US1 → US2 → US3 → Polish.
- US2 depende de JWTManager (US1); US3 depende de RequireJWT (US2).
