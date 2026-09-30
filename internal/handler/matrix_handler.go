package handler

import (
	"context"

	"github.com/elfab/retotecnico/api-go/internal/domain"
	"github.com/elfab/retotecnico/api-go/internal/model"
	"github.com/elfab/retotecnico/api-go/internal/service"
	"github.com/gofiber/fiber/v3"
)

// MatrixAnalyzer es el caso de uso de análisis de matrices que necesita el handler.
type MatrixAnalyzer interface {
	Analyze(ctx context.Context, m domain.Matrix) (service.Analysis, error)
}

// MatrixHandler expone la factorización QR por HTTP.
type MatrixHandler struct {
	analyzer MatrixAnalyzer
}

// NewMatrixHandler crea el handler de matrices.
func NewMatrixHandler(analyzer MatrixAnalyzer) *MatrixHandler {
	return &MatrixHandler{analyzer: analyzer}
}

// QR maneja POST /api/v1/matrix/qr: factoriza la matriz y devuelve Q, R y sus estadísticas.
func (h *MatrixHandler) QR(c fiber.Ctx) error {
	var req model.QRRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "el cuerpo debe ser un JSON con \"matrix\": un array de arrays de números")
	}

	analysis, err := h.analyzer.Analyze(c.Context(), domain.Matrix(req.Matrix))
	if err != nil {
		return err
	}

	return c.JSON(toQRResponse(analysis))
}

func toQRResponse(a service.Analysis) model.QRResponse {
	matrices := make([]model.MatrixDiagonalDTO, len(a.Statistics.Matrices))
	for i, m := range a.Statistics.Matrices {
		matrices[i] = model.MatrixDiagonalDTO{Name: m.Name, IsDiagonal: m.IsDiagonal}
	}
	s := a.Statistics
	return model.QRResponse{
		Q: a.QR.Q,
		R: a.QR.R,
		Statistics: model.StatisticsDTO{
			Max: s.Max, Min: s.Min, Average: s.Average, Sum: s.Sum,
			Count: s.Count, AnyDiagonal: s.AnyDiagonal, Matrices: matrices,
		},
	}
}
