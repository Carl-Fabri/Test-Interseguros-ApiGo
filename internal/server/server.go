// Package server es la raíz de composición de api-go: construye las dependencias,
// registra middlewares y rutas. Se separa de cmd/api para poder probar la app completa.
package server

import (
	"io"
	"os"

	"github.com/elfab/retotecnico/api-go/api"
	"github.com/elfab/retotecnico/api-go/internal/auth"
	"github.com/elfab/retotecnico/api-go/internal/config"
	"github.com/elfab/retotecnico/api-go/internal/handler"
	"github.com/elfab/retotecnico/api-go/internal/middleware"
	"github.com/elfab/retotecnico/api-go/internal/model"
	"github.com/elfab/retotecnico/api-go/internal/service"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// Options permite ajustar la app sin cambiar la configuración (útil en pruebas).
type Options struct {
	LogOutput io.Writer // destino de los logs de acceso; nil usa os.Stdout
}

// New crea la aplicación con todas sus rutas. stats es el adaptador hacia api-node:
// en producción el cliente HTTP real, en pruebas un doble.
func New(cfg config.Config, stats service.StatisticsClient, opts Options) *fiber.App {
	if opts.LogOutput == nil {
		opts.LogOutput = os.Stdout
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.JWTTTL, nil)
	authHandler := handler.NewAuthHandler(service.NewAuthService(cfg.AuthUsername, cfg.AuthPassword, jwtManager))
	matrixHandler := handler.NewMatrixHandler(service.NewMatrixService(stats, cfg.MatrixMaxDimension))

	app := fiber.New(fiber.Config{
		AppName:      "api-go",
		ErrorHandler: middleware.ErrorHandler,
		BodyLimit:    1 << 20, // 1 MiB: suficiente para matrices de 100x100
	})

	// * Middlewares transversales (el orden importa):
	// *   recover (no tumbar el proceso) → requestid (trazabilidad) → logger → CORS → helmet.
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Stream: opts.LogOutput,
		Format: "[${time}] ${requestid} ${status} - ${latency} ${method} ${path} ${error}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowHeaders: []string{fiber.HeaderAuthorization, fiber.HeaderContentType},
		AllowMethods: []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodOptions},
	}))
	app.Use(helmet.New(helmet.Config{
		// La página de Scalar carga su script desde un CDN, así que se excluye de helmet.
		Next: func(c fiber.Ctx) bool { return c.Path() == "/docs" },
	}))

	// Rutas públicas.
	// * GET / y GET /health responden 200: los balanceadores (p. ej. ECS Express Mode) hacen el health check
	// * en "/" por defecto; así el servicio queda sano con cualquier configuración.
	health := func(c fiber.Ctx) error {
		return c.JSON(model.HealthResponse{Status: "ok", Service: "api-go"})
	}
	app.Get("/", health)
	app.Get("/health", health)
	app.Get("/openapi.yaml", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml")
		return c.Send(api.OpenAPISpec)
	})
	app.Get("/docs", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.Send(api.DocsHTML)
	})

	v1 := app.Group("/api/v1")
	v1.Post("/auth/login", authHandler.Login)

	// ! Rutas protegidas: todo endpoint de negocio nuevo debe registrarse con RequireJWT.
	v1.Post("/matrix/qr", middleware.RequireJWT(jwtManager), matrixHandler.QR)

	// Cualquier otra ruta: 404 con el formato de error estándar.
	app.Use(func(fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "recurso no encontrado")
	})

	return app
}
