// Package client contiene los adaptadores salientes de api-go (llamadas HTTP a otros servicios).
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/elfab/retotecnico/api-go/internal/auth"
	"github.com/elfab/retotecnico/api-go/internal/domain"
	"github.com/elfab/retotecnico/api-go/internal/service"
)

// statisticsPath es la ruta del endpoint de estadísticas de api-node (ver su openapi.yaml).
const statisticsPath = "/api/v1/statistics"

// StatisticsHTTPClient implementa service.StatisticsClient llamando a api-node por HTTP.
type StatisticsHTTPClient struct {
	baseURL string
	http    *http.Client
}

// NewStatisticsHTTPClient crea el cliente con la URL base de api-node y un timeout total por petición.
func NewStatisticsHTTPClient(baseURL string, timeout time.Duration) *StatisticsHTTPClient {
	return &StatisticsHTTPClient{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

// Contrato de api-node (DTOs de transporte, privados a este adaptador).
type (
	statisticsRequest struct {
		Matrices []namedMatrixDTO `json:"matrices"`
	}
	namedMatrixDTO struct {
		Name   string        `json:"name"`
		Values domain.Matrix `json:"values"`
	}
	statisticsResponse struct {
		Max         float64 `json:"max"`
		Min         float64 `json:"min"`
		Average     float64 `json:"average"`
		Sum         float64 `json:"sum"`
		Count       int     `json:"count"`
		AnyDiagonal bool    `json:"anyDiagonal"`
		Matrices    []struct {
			Name       string `json:"name"`
			IsDiagonal bool   `json:"isDiagonal"`
		} `json:"matrices"`
	}
)

// Compute envía las matrices a api-node reenviando el JWT del usuario (tomado del contexto).
// Los fallos se envuelven en service.ErrStatisticsTimeout o service.ErrStatisticsUnavailable.
func (c *StatisticsHTTPClient) Compute(ctx context.Context, matrices []service.NamedMatrix) (service.Statistics, error) {
	payload := statisticsRequest{Matrices: make([]namedMatrixDTO, len(matrices))}
	for i, m := range matrices {
		payload.Matrices[i] = namedMatrixDTO{Name: m.Name, Values: m.Values}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return service.Statistics{}, fmt.Errorf("serializar petición: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+statisticsPath, bytes.NewReader(body))
	if err != nil {
		return service.Statistics{}, fmt.Errorf("%w: %v", service.ErrStatisticsUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// * Propagación del JWT del usuario: api-node valida el mismo token con el mismo secreto.
	if token, ok := auth.TokenFromContext(ctx); ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// * El timeout del http.Client (NODE_API_TIMEOUT_MS) acota la espera: api-node lento → 504, caído → 502.
	resp, err := c.http.Do(req)
	if err != nil {
		if isTimeout(err) {
			return service.Statistics{}, fmt.Errorf("%w: %v", service.ErrStatisticsTimeout, err)
		}
		return service.Statistics{}, fmt.Errorf("%w: %v", service.ErrStatisticsUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return service.Statistics{}, fmt.Errorf("%w: api-node respondió %d: %s",
			service.ErrStatisticsUnavailable, resp.StatusCode, bytes.TrimSpace(detail))
	}

	var out statisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return service.Statistics{}, fmt.Errorf("%w: respuesta inválida: %v", service.ErrStatisticsUnavailable, err)
	}

	stats := service.Statistics{
		Max: out.Max, Min: out.Min, Average: out.Average, Sum: out.Sum,
		Count: out.Count, AnyDiagonal: out.AnyDiagonal,
		Matrices: make([]service.MatrixDiagonal, len(out.Matrices)),
	}
	for i, m := range out.Matrices {
		stats.Matrices[i] = service.MatrixDiagonal{Name: m.Name, IsDiagonal: m.IsDiagonal}
	}
	return stats, nil
}

func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
