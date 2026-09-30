---
name: fiber-v3-endpoint
description: Convenciones de api-go para crear o modificar endpoints HTTP con Fiber v3 (handler, DTO, service, middleware, errores y pruebas con app.Test). Usar al agregar rutas, handlers, middlewares o el cliente HTTP hacia api-node.
---

# Endpoints con Fiber v3 en api-go

## Flujo por capas (en este orden)
1. **Contrato**: confirmar el endpoint en `api/openapi.yaml`. Si no existe, detenerse y pedir que se defina.
2. **Dominio** en `internal/domain/`: matemática pura y reglas de validación (`*domain.ValidationError`). Sin imports del servicio.
3. **Service** en `internal/service/`: casos de uso que orquestan el dominio y definen **puertos** (interfaces) en `ports.go`
   para lo externo (p. ej. `StatisticsClient`). **Sin importar Fiber**. Los errores de puertos son sentinelas (`ErrStatistics*`).
4. **Adaptadores** en `internal/client/` (salientes, implementan puertos) y `internal/auth/` (JWT).
5. **DTO** en `internal/model/`: structs de request y response con tags `json` (contrato público).
6. **Handler** en `internal/handler/`: parsea el DTO, llama al caso de uso y mapea el resultado. Sin lógica de dominio.
7. **Registro de ruta** en `internal/server/server.go` (raíz de composición), con prefijo `/api/v1` y `middleware.RequireJWT` si es protegida.
8. **Errores**: si aparece un error nuevo, mapearlo a HTTP **solo** en `internal/middleware/error_handler.go`.
9. **Pruebas**: unitarias de dominio y service (table-driven, con dobles de los puertos) e integración con `app.Test` en `test/integration/`.

## Fiber v3: diferencias con v2 que hay que respetar
- Import `github.com/gofiber/fiber/v3`. Los handlers son `func(c fiber.Ctx) error` (**interfaz, no puntero**).
- Parseo del body: `c.Bind().Body(&req)` (no existe `BodyParser`).
- Errores HTTP: `return fiber.NewError(fiber.StatusBadRequest, "mensaje")`.
- Manejo centralizado de errores con `fiber.Config{ErrorHandler: ...}` y un formato JSON único:
  `{"error": {"code": "INVALID_MATRIX", "message": "..."}}`.
- Pruebas: `resp, err := app.Test(req)` con `httptest.NewRequest`, construyendo la app con `server.New(cfg)`.

## Esqueleto de handler
```go
// QRHandler expone la factorización QR por HTTP.
type QRHandler struct {
	svc service.QRService
}

// Factorize maneja POST /api/v1/qr.
func (h *QRHandler) Factorize(c fiber.Ctx) error {
	var req model.QRRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "body inválido")
	}
	result, err := h.svc.Factorize(req.Matrix)
	if err != nil {
		return err // el ErrorHandler lo traduce a JSON
	}
	return c.JSON(result)
}
```

## Cliente HTTP hacia api-node (`internal/client/`)
- `net/http` con `http.Client{Timeout: cfg.NodeAPITimeout}` (sale de `NODE_API_TIMEOUT_MS`).
- La URL base viene de `cfg.NodeAPIURL` (`NODE_API_URL`). Nunca hardcodear hosts.
- Si api-node no responde, devolver `502 Bad Gateway`; si se agota el tiempo, `504 Gateway Timeout`.
- Definir el cliente como interfaz para poder simularlo en las pruebas del handler.

## Checklist antes de terminar
- [ ] `go fmt ./... && go vet ./...` limpios
- [ ] Pruebas unitarias nuevas o actualizadas; `go test ./...` en verde
- [ ] GoDoc en todo lo exportado; `.claude/CLAUDE.md` y `README.md` al día
- [ ] El contrato en `api/openapi.yaml` coincide con lo implementado
- [ ] Docker revisado: `Dockerfile`, `docker-compose.yml`, `.env.example` si cambiaron dependencias, variables o puertos
