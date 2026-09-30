package service

import (
	"context"

	"github.com/elfab/retotecnico/api-go/internal/domain"
)

// Analysis es el resultado completo de analizar una matriz: su QR y las estadísticas de Q y R.
type Analysis struct {
	QR         domain.QRResult
	Statistics Statistics
}

// MatrixService es el caso de uso principal: valida la matriz, calcula su QR y
// solicita a api-node las estadísticas de las matrices resultantes.
type MatrixService struct {
	stats        StatisticsClient
	maxDimension int
}

// NewMatrixService crea el servicio con su cliente de estadísticas y el tamaño máximo permitido.
func NewMatrixService(stats StatisticsClient, maxDimension int) *MatrixService {
	return &MatrixService{stats: stats, maxDimension: maxDimension}
}

// Analyze ejecuta el flujo completo. Devuelve *domain.ValidationError si la matriz es inválida
// o un error envuelto en ErrStatisticsUnavailable / ErrStatisticsTimeout si falla api-node.
func (s *MatrixService) Analyze(ctx context.Context, m domain.Matrix) (Analysis, error) {
	// * Caso de uso principal: validar → factorizar → pedir estadísticas a api-node.
	// ? Si api-node falla, la petición falla (502/504): el contrato promete Q, R y estadísticas juntas (ADR-002 §4).
	if err := m.Validate(s.maxDimension); err != nil {
		return Analysis{}, err
	}

	qr := domain.QRGivens(m)

	stats, err := s.stats.Compute(ctx, []NamedMatrix{
		{Name: "Q", Values: qr.Q},
		{Name: "R", Values: qr.R},
	})
	if err != nil {
		return Analysis{}, err
	}

	return Analysis{QR: qr, Statistics: stats}, nil
}
