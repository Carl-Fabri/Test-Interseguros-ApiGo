// Package service contiene los casos de uso de api-go. Orquesta el dominio y define los
// puertos (interfaces) que implementan los adaptadores externos, sin conocer HTTP.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/elfab/retotecnico/api-go/internal/domain"
)

// Errores que los adaptadores de estadísticas deben envolver para que la capa HTTP
// los traduzca a 502/504 sin conocer detalles de transporte.
var (
	ErrStatisticsUnavailable = errors.New("el servicio de estadísticas no está disponible")
	ErrStatisticsTimeout     = errors.New("el servicio de estadísticas no respondió a tiempo")
	ErrInvalidCredentials    = errors.New("usuario o contraseña incorrectos")
)

// NamedMatrix es una matriz identificada por nombre (p. ej. "Q" o "R").
type NamedMatrix struct {
	Name   string
	Values domain.Matrix
}

// MatrixDiagonal indica si una matriz concreta es diagonal.
type MatrixDiagonal struct {
	Name       string
	IsDiagonal bool
}

// Statistics es el resultado que devuelve el servicio de estadísticas (api-node).
type Statistics struct {
	Max         float64
	Min         float64
	Average     float64
	Sum         float64
	Count       int
	AnyDiagonal bool
	Matrices    []MatrixDiagonal
}

// StatisticsClient es el puerto hacia el servicio de estadísticas.
// El contexto transporta la cancelación y el token del usuario a reenviar.
type StatisticsClient interface {
	Compute(ctx context.Context, matrices []NamedMatrix) (Statistics, error)
}

// TokenIssuer es el puerto para emitir tokens de acceso.
type TokenIssuer interface {
	Issue(subject string) (token string, expiresAt time.Time, err error)
}
