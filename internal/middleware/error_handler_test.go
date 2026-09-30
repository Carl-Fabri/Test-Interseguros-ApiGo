package middleware

import (
	"errors"
	"fmt"
	"testing"

	"github.com/elfab/retotecnico/api-go/internal/domain"
	"github.com/elfab/retotecnico/api-go/internal/service"
	"github.com/gofiber/fiber/v3"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"validación de dominio", &domain.ValidationError{Message: "x"}, 400, "INVALID_MATRIX"},
		{"credenciales inválidas", service.ErrInvalidCredentials, 401, "INVALID_CREDENTIALS"},
		{"timeout envuelto", fmt.Errorf("%w: detalle", service.ErrStatisticsTimeout), 504, "STATISTICS_TIMEOUT"},
		{"api-node caído envuelto", fmt.Errorf("%w: detalle", service.ErrStatisticsUnavailable), 502, "STATISTICS_UNAVAILABLE"},
		{"error HTTP 401", fiber.NewError(401, "x"), 401, "UNAUTHORIZED"},
		{"error HTTP 413", fiber.NewError(413, "x"), 413, "PAYLOAD_TOO_LARGE"},
		{"error HTTP no mapeado", fiber.NewError(418, "x"), 418, "HTTP_ERROR"},
		{"error desconocido no filtra detalles", errors.New("sql: secreto"), 500, "INTERNAL_ERROR"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, code, msg := classify(tc.err)
			if status != tc.wantStatus || code != tc.wantCode {
				t.Fatalf("classify = (%d, %s), want (%d, %s)", status, code, tc.wantStatus, tc.wantCode)
			}
			if status == 500 && msg != "error interno del servidor" {
				t.Errorf("un 500 no debe exponer el error interno: %q", msg)
			}
		})
	}
}
