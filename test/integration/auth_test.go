package integration

import (
	"net/http"
	"testing"
)

func TestLogin(t *testing.T) {
	env := newTestEnv(t, fakeNodeOK)

	t.Run("credenciales válidas emiten un Bearer token", func(t *testing.T) {
		resp, body := env.do(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"username": testUser, "password": testPassword})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, body = %v", resp.StatusCode, body)
		}
		if body["tokenType"] != "Bearer" || body["accessToken"] == "" || body["expiresIn"].(float64) <= 0 {
			t.Errorf("respuesta inválida: %v", body)
		}
	})

	tests := []struct {
		name       string
		payload    any
		wantStatus int
		wantCode   string
	}{
		{"contraseña incorrecta", map[string]string{"username": testUser, "password": "mal"}, http.StatusUnauthorized, "INVALID_CREDENTIALS"},
		{"campos vacíos", map[string]string{"username": "", "password": ""}, http.StatusBadRequest, "INVALID_REQUEST"},
		{"JSON mal formado", "{no-json", http.StatusBadRequest, "INVALID_REQUEST"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := env.do(t, http.MethodPost, "/api/v1/auth/login", "", tc.payload)
			if resp.StatusCode != tc.wantStatus || errorCode(body) != tc.wantCode {
				t.Fatalf("status = %d, body = %v; want %d %s", resp.StatusCode, body, tc.wantStatus, tc.wantCode)
			}
		})
	}
}

func TestProtectedRouteRequiresValidToken(t *testing.T) {
	env := newTestEnv(t, fakeNodeOK)
	payload := map[string]any{"matrix": [][]float64{{1}}}

	for name, token := range map[string]string{"sin token": "", "token inválido": "abc.def.ghi"} {
		t.Run(name, func(t *testing.T) {
			resp, body := env.do(t, http.MethodPost, "/api/v1/matrix/qr", token, payload)
			if resp.StatusCode != http.StatusUnauthorized || errorCode(body) != "UNAUTHORIZED" {
				t.Fatalf("status = %d, body = %v", resp.StatusCode, body)
			}
		})
	}
	if env.node.callCount != 0 {
		t.Error("sin un token válido no se debe llamar a api-node")
	}
}
