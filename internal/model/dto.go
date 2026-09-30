// Package model define los DTOs de entrada y salida de la API HTTP de api-go.
// Son el contrato público (ver api/openapi.yaml) y se mantienen separados de los tipos de dominio.
package model

// LoginRequest es el cuerpo de POST /api/v1/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse es la respuesta de un login exitoso.
type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"` // siempre "Bearer"
	ExpiresIn   int64  `json:"expiresIn"` // segundos hasta la expiración
}

// QRRequest es el cuerpo de POST /api/v1/matrix/qr.
type QRRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

// QRResponse es el resultado de la factorización QR junto con las estadísticas de api-node.
type QRResponse struct {
	Q          [][]float64   `json:"q"`
	R          [][]float64   `json:"r"`
	Statistics StatisticsDTO `json:"statistics"`
}

// StatisticsDTO son las estadísticas calculadas por api-node sobre Q y R.
type StatisticsDTO struct {
	Max         float64             `json:"max"`
	Min         float64             `json:"min"`
	Average     float64             `json:"average"`
	Sum         float64             `json:"sum"`
	Count       int                 `json:"count"`
	AnyDiagonal bool                `json:"anyDiagonal"`
	Matrices    []MatrixDiagonalDTO `json:"matrices"`
}

// MatrixDiagonalDTO indica si una matriz concreta es diagonal.
type MatrixDiagonalDTO struct {
	Name       string `json:"name"`
	IsDiagonal bool   `json:"isDiagonal"`
}

// HealthResponse es la respuesta de GET /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// ErrorResponse es el formato único de error de la API.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody describe un error con un código estable y un mensaje legible.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
