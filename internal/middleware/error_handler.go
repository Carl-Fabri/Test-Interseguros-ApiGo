package middleware

import (
	"errors"
	"log/slog"

	"github.com/elfab/retotecnico/api-go/internal/domain"
	"github.com/elfab/retotecnico/api-go/internal/model"
	"github.com/elfab/retotecnico/api-go/internal/service"
	"github.com/gofiber/fiber/v3"
)

// ErrorHandler traduce cualquier error devuelto por handlers o middlewares al formato
// único de error de la API: {"error": {"code": "...", "message": "..."}}.
// Es el único lugar que conoce la correspondencia entre errores de dominio y códigos HTTP.
func ErrorHandler(c fiber.Ctx, err error) error {
	status, code, message := classify(err)
	if status >= fiber.StatusInternalServerError {
		slog.Error("error procesando petición", "method", c.Method(), "path", c.Path(), "status", status, "error", err)
	}
	return c.Status(status).JSON(model.ErrorResponse{Error: model.ErrorBody{Code: code, Message: message}})
}

func classify(err error) (status int, code, message string) {
	// * Único lugar que traduce errores de dominio/servicio a HTTP: agregar aquí cualquier error nuevo.
	// ? errors.Is/As atraviesa errores envueltos con %w (p. ej. "ErrStatisticsTimeout: detalle").
	var validationErr *domain.ValidationError
	var fiberErr *fiber.Error

	switch {
	case errors.As(err, &validationErr):
		return fiber.StatusBadRequest, "INVALID_MATRIX", validationErr.Message
	case errors.Is(err, service.ErrInvalidCredentials):
		return fiber.StatusUnauthorized, "INVALID_CREDENTIALS", service.ErrInvalidCredentials.Error()
	case errors.Is(err, service.ErrStatisticsTimeout):
		return fiber.StatusGatewayTimeout, "STATISTICS_TIMEOUT", service.ErrStatisticsTimeout.Error()
	case errors.Is(err, service.ErrStatisticsUnavailable):
		return fiber.StatusBadGateway, "STATISTICS_UNAVAILABLE", service.ErrStatisticsUnavailable.Error()
	case errors.As(err, &fiberErr):
		return fiberErr.Code, codeForStatus(fiberErr.Code), fiberErr.Message
	default:
		// ! Nunca exponer el detalle interno al cliente (puede contener datos sensibles); queda en el log.
		return fiber.StatusInternalServerError, "INTERNAL_ERROR", "error interno del servidor"
	}
}

func codeForStatus(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "INVALID_REQUEST"
	case fiber.StatusUnauthorized:
		return "UNAUTHORIZED"
	case fiber.StatusNotFound:
		return "NOT_FOUND"
	case fiber.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case fiber.StatusRequestEntityTooLarge:
		return "PAYLOAD_TOO_LARGE"
	case fiber.StatusUnsupportedMediaType:
		return "UNSUPPORTED_MEDIA_TYPE"
	default:
		if status >= fiber.StatusInternalServerError {
			return "INTERNAL_ERROR"
		}
		return "HTTP_ERROR"
	}
}
