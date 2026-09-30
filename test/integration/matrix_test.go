package integration

import (
	"math"
	"net/http"
	"testing"
	"time"
)

func TestQRHappyPath(t *testing.T) {
	env := newTestEnv(t, fakeNodeOK)
	token := env.login(t)

	resp, body := env.do(t, http.MethodPost, "/api/v1/matrix/qr", token,
		map[string]any{"matrix": [][]float64{{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %v", resp.StatusCode, body)
	}

	r := body["r"].([]any)
	if got := r[0].([]any)[0].(float64); math.Abs(got-14) > 1e-9 {
		t.Errorf("R[0][0] = %g, want 14", got)
	}
	if q := body["q"].([]any); len(q) != 3 {
		t.Errorf("Q debe ser 3x3, tiene %d filas", len(q))
	}
	if stats := body["statistics"].(map[string]any); stats["max"].(float64) != 35 {
		t.Errorf("las estadísticas de api-node deben incluirse en la respuesta: %v", stats)
	}

	// Comunicación con api-node: reenvía el mismo JWT y envía Q y R.
	if env.node.authz != "Bearer "+token {
		t.Errorf("api-node recibió Authorization = %q, se esperaba el token del usuario", env.node.authz)
	}
	matrices := env.node.body["matrices"].([]any)
	if len(matrices) != 2 || matrices[0].(map[string]any)["name"] != "Q" || matrices[1].(map[string]any)["name"] != "R" {
		t.Errorf("api-node debe recibir Q y R, recibió %v", env.node.body)
	}
}

func TestQRInvalidInput(t *testing.T) {
	env := newTestEnv(t, fakeNodeOK)
	token := env.login(t)

	tests := []struct {
		name     string
		payload  any
		wantCode string
	}{
		{"matriz no rectangular", map[string]any{"matrix": [][]float64{{1, 2}, {3}}}, "INVALID_MATRIX"},
		{"matriz vacía", map[string]any{"matrix": [][]float64{}}, "INVALID_MATRIX"},
		{"sin campo matrix", map[string]any{}, "INVALID_MATRIX"},
		{"excede el tamaño máximo", map[string]any{"matrix": [][]float64{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}}}, "INVALID_MATRIX"},
		{"valores no numéricos", `{"matrix":[["a","b"]]}`, "INVALID_REQUEST"},
		{"JSON mal formado", `{"matrix":`, "INVALID_REQUEST"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := env.do(t, http.MethodPost, "/api/v1/matrix/qr", token, tc.payload)
			if resp.StatusCode != http.StatusBadRequest || errorCode(body) != tc.wantCode {
				t.Fatalf("status = %d, body = %v; want 400 %s", resp.StatusCode, body, tc.wantCode)
			}
		})
	}
	if env.node.callCount != 0 {
		t.Error("una entrada inválida no debe llegar a api-node")
	}
}

func TestQRWhenStatisticsServiceFails(t *testing.T) {
	tests := []struct {
		name       string
		node       http.HandlerFunc
		wantStatus int
		wantCode   string
	}{
		{
			name:       "api-node responde error",
			node:       func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
			wantStatus: http.StatusBadGateway, wantCode: "STATISTICS_UNAVAILABLE",
		},
		{
			name:       "api-node no responde a tiempo",
			node:       func(_ http.ResponseWriter, _ *http.Request) { time.Sleep(500 * time.Millisecond) },
			wantStatus: http.StatusGatewayTimeout, wantCode: "STATISTICS_TIMEOUT",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t, tc.node)
			resp, body := env.do(t, http.MethodPost, "/api/v1/matrix/qr", env.login(t), map[string]any{"matrix": [][]float64{{1, 2}, {3, 4}}})
			if resp.StatusCode != tc.wantStatus || errorCode(body) != tc.wantCode {
				t.Fatalf("status = %d, body = %v; want %d %s", resp.StatusCode, body, tc.wantStatus, tc.wantCode)
			}
		})
	}
}
