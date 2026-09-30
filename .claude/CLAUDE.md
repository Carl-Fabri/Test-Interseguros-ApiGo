# api-go — API de factorización QR (Go + Fiber v3)

> **Servicio autónomo.** No forma parte de un monolito: tiene su propio código, configuración,
> imagen Docker, pruebas y despliegue. Funciona por sí solo, sin el resto del repositorio.

## Responsabilidad
Recibir una matriz rectangular m×n, calcular su **factorización QR completa por rotaciones de Givens**
(Q m×m ortogonal, R m×n triangular superior con diagonal ≥ 0), enviar Q y R a api-node por HTTP
reenviando el JWT del usuario, y devolver Q, R y las estadísticas.

## Límites del servicio
- api-node es una **dependencia externa**: se consume solo por HTTP en `NODE_API_URL`, según el contrato que publica
  en su `/openapi.yaml`. Nunca leer ni importar código de `../api-node`.
- Manejar su indisponibilidad: timeout (`NODE_API_TIMEOUT_MS`) → 504 `STATISTICS_TIMEOUT`; caída o error → 502 `STATISTICS_UNAVAILABLE`.
- Toda la configuración entra por variables de entorno (`internal/config`), sin valores fijos en el código.

## Endpoints (contrato: `api/openapi.yaml`)
| Método | Ruta                  | Auth | Descripción                                   |
|--------|-----------------------|------|-----------------------------------------------|
| GET    | `/health`             | No   | Estado del servicio                           |
| GET    | `/docs`               | No   | Referencia interactiva (Scalar)               |
| GET    | `/openapi.yaml`       | No   | Contrato OpenAPI 3.1                          |
| POST   | `/api/v1/auth/login`  | No   | Emite un JWT HS256 para el usuario de demo    |
| POST   | `/api/v1/matrix/qr`   | JWT  | QR de la matriz + estadísticas de api-node    |

## Arquitectura (Clean Architecture; las dependencias apuntan hacia el dominio)
```
cmd/api/             Punto de entrada: config, cliente de api-node, arranque y apagado ordenado
api/                 openapi.yaml + docs.html (Scalar), embebidos en el binario con go:embed
internal/domain/     Núcleo puro: Matrix, Validate, QRGivens. No importa nada del servicio
internal/service/    Casos de uso (MatrixService, AuthService) y puertos (ports.go): StatisticsClient, TokenIssuer
internal/client/     Adaptador saliente: StatisticsHTTPClient (implementa StatisticsClient)
internal/auth/       JWTManager (emite/valida HS256) y transporte del token en context.Context
internal/model/      DTOs HTTP de entrada/salida y formato de error
internal/handler/    Handlers: DTO → caso de uso → DTO. Sin lógica de negocio
internal/middleware/ RequireJWT y ErrorHandler (único mapeo error → HTTP)
internal/server/     Raíz de composición: dependencias, middlewares (recover, requestid, logger, CORS, helmet) y rutas
test/integration/    App completa + cliente real contra un api-node simulado con httptest
deploy/aws/          service.yml (CloudFormation ECS Fargate) + deploy.sh (build → ECR → stack)
```

## Convenciones
- Fiber v3: handlers `func(c fiber.Ctx) error`, body con `c.Bind().Body(&req)`, errores con `fiber.NewError`.
- `domain/` y `service/` no importan Fiber. Un error nuevo se mapea a HTTP solo en `middleware/error_handler.go`.
- El token se propaga con `auth.WithToken(ctx)` / `auth.TokenFromContext(ctx)`; el servicio no lo conoce.
- GoDoc en todo lo exportado. Pruebas table-driven, flotantes con tolerancia.

## Comentarios (Better Comments)
En los métodos clave: `// *` punto clave · `// !` advertencia o seguridad · `// ?` decisión de diseño · `// TODO:` pendiente.
La documentación de la API pública sigue siendo api-go.

## SDD (Spec Kit)
Cada funcionalidad se especifica antes de implementarse: `.specify/memory/constitution.md` (principios) y
`specs/NNN-funcionalidad/` (spec, plan, research, data-model, contracts, quickstart, tasks y checklists).
Comandos: `/speckit-specify` → `/speckit-clarify` → `/speckit-plan` → `/speckit-tasks` → `/speckit-implement`.

## Definición de terminado (OBLIGATORIA en cada cambio)
Ningún cambio está terminado hasta cumplir **los tres puntos**. Si alguno no aplica, decirlo explícitamente en el resumen.
1. **Pruebas**: agregar o actualizar pruebas unitarias de toda lógica nueva o modificada, y de integración si cambia
   un endpoint. Un bug corregido lleva una prueba que lo reproduce. `go test ./...` en verde.
2. **Documentación**: GoDoc de lo exportado, este `CLAUDE.md` (estructura, comandos y variables), `README.md` y,
   si cambia la API, el contrato en `api/openapi.yaml`.
3. **Docker**: revisar y actualizar si hace falta `Dockerfile`, `docker-compose.yml`, `.env.example` y `.dockerignore`
   cuando cambien dependencias, variables de entorno, puertos, archivos que la imagen necesita o pasos de build.
   Validar con `docker compose build`.
   Si cambia una variable de entorno, un puerto o el health check, actualizar también `deploy/aws/service.yml`
   (y validar con `cfn-lint`). Guía: `docs/deploy/aws.md`.

El CI (`.github/workflows/ci.yml`) bloquea el cambio si fallan el formato, `go vet`, las pruebas unitarias,
las de integración o el build de la imagen.

## Comandos (desde `api-go/`)
- Ejecutar: `go run ./cmd/api` (carga `.env` si existe; las variables del entorno tienen prioridad)
- Pruebas unitarias: `go test ./internal/... ./cmd/...` · integración: `go test ./test/integration/...`
- Cobertura total: `go test -coverpkg=./internal/... -coverprofile=c.out ./... && go tool cover -func=c.out`
- Formato/vet: `go fmt ./... && go vet ./...`
- Docker aislado: `docker compose up --build`
- Despliegue en AWS: `./deploy/aws/deploy.sh` (crea la plataforma compartida si falta)
- **Windows con Smart App Control**: si bloquea un `*.test.exe`, correr las pruebas en contenedor:
  `docker run --rm -v "${PWD}:/src" -w /src golang:1.27-alpine go test ./...`

## Variables de entorno (ver `.env.example`)
| Variable               | Default                 | Descripción                                          |
|------------------------|-------------------------|------------------------------------------------------|
| `PORT`                 | 8080                    | Puerto HTTP                                          |
| `NODE_API_URL`         | http://localhost:3000   | URL base de api-node                                 |
| `NODE_API_TIMEOUT_MS`  | 5000                    | Timeout de llamadas a api-node                       |
| `JWT_SECRET`           | — (obligatoria, ≥ 32)   | Secreto HS256, **igual al de api-node**              |
| `JWT_ISSUER`           | api-go                  | Claim `iss`, igual al esperado por api-node          |
| `JWT_AUDIENCE`         | matrix-services         | Claim `aud`, igual al esperado por api-node          |
| `JWT_TTL_MINUTES`      | 60                      | Vigencia del token                                   |
| `AUTH_USERNAME`        | — (obligatoria)         | Usuario de demo del login                            |
| `AUTH_PASSWORD`        | — (obligatoria)         | Contraseña de demo del login                         |
| `CORS_ALLOWED_ORIGINS` | http://localhost:4200   | Orígenes permitidos, separados por coma              |
| `MATRIX_MAX_DIMENSION` | 100                     | Máximo de filas y columnas por matriz                |
| `DOCKER_NODE_API_URL`  | http://host.docker.internal:3000 | Solo compose: reemplaza `NODE_API_URL` dentro del contenedor |

## Agentes, skills y memoria
- **Subagentes** (`.claude/agents/`, con memoria persistente en `.claude/agent-memory/<nombre>/`):
  - `go-api-developer`: implementa tareas o specs. Delegarle el código de producción y sus pruebas.
  - `go-api-reviewer`: revisa sin editar (correctitud de la QR, JWT, resiliencia, pruebas). Usarlo tras cada funcionalidad.
- **Skills**: `fiber-v3-endpoint` (propia), `golang-*` de samber y `speckit-*` (Spec Kit, flujo SDD).
  Versiones de terceros fijadas en `skills-lock.json`.
- **Reglas por ruta**: `.claude/rules/testing.md` se carga solo al tocar pruebas.
- **Dónde guardar cada cosa**: regla permanente → este archivo o `.claude/rules/`; decisión o contrato →
  `docs/architecture/` o `api/openapi.yaml`; contexto o preferencia que no se deduce del código → memoria. Al cerrar: `/retro`.
