// Package integration contiene las pruebas de integración de api-go: construyen la app
// completa (server.New) con el cliente HTTP real, apuntando a un api-node simulado con
// httptest. No abren puertos propios ni dependen de servicios externos.
package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/elfab/retotecnico/api-go/internal/client"
	"github.com/elfab/retotecnico/api-go/internal/config"
	"github.com/elfab/retotecnico/api-go/internal/server"
	"github.com/gofiber/fiber/v3"
)

const (
	testUser     = "admin"
	testPassword = "admin123"
)

// testEnv agrupa la app bajo prueba y lo que recibió el api-node simulado.
type testEnv struct {
	app  *fiber.App
	node *nodeRecorder
}

// nodeRecorder registra las peticiones que llegan al api-node simulado.
type nodeRecorder struct {
	mu        sync.Mutex
	authz     string
	body      map[string]any
	callCount int
}

func (n *nodeRecorder) record(r *http.Request) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.callCount++
	n.authz = r.Header.Get("Authorization")
	n.body = map[string]any{}
	_ = json.NewDecoder(r.Body).Decode(&n.body)
}

// fakeNodeOK simula un api-node que responde estadísticas válidas.
func fakeNodeOK(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"max":35,"min":-0.94,"average":4.2,"sum":75.6,"count":18,"anyDiagonal":false,
		"matrices":[{"name":"Q","isDiagonal":false},{"name":"R","isDiagonal":false}]}`))
}

func newTestEnv(t *testing.T, nodeHandler http.HandlerFunc) *testEnv {
	t.Helper()
	rec := &nodeRecorder{}
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.record(r)
		nodeHandler(w, r)
	}))
	t.Cleanup(node.Close)

	cfg := config.Config{
		JWTSecret: strings.Repeat("t", 32), JWTIssuer: "api-go", JWTAudience: "matrix-services", JWTTTL: time.Hour,
		AuthUsername: testUser, AuthPassword: testPassword,
		CORSAllowedOrigins: []string{"http://localhost:4200"}, MatrixMaxDimension: 10,
	}
	stats := client.NewStatisticsHTTPClient(node.URL, 200*time.Millisecond)

	return &testEnv{app: server.New(cfg, stats, server.Options{LogOutput: io.Discard}), node: rec}
}

// do ejecuta una petición contra la app y decodifica el JSON de respuesta (si lo hay).
func (e *testEnv) do(t *testing.T, method, path, token string, payload any) (*http.Response, map[string]any) {
	t.Helper()
	var body io.Reader
	if payload != nil {
		raw, ok := payload.(string)
		if !ok {
			b, _ := json.Marshal(payload)
			raw = string(b)
		}
		body = bytes.NewBufferString(raw)
	}
	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := e.app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })

	out := map[string]any{}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(resp.Body).Decode(&out)
	}
	return resp, out
}

// login obtiene un token válido.
func (e *testEnv) login(t *testing.T) string {
	t.Helper()
	resp, body := e.do(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"username": testUser, "password": testPassword})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: status %d, body %v", resp.StatusCode, body)
	}
	return body["accessToken"].(string)
}

func errorCode(body map[string]any) string {
	errObj, _ := body["error"].(map[string]any)
	code, _ := errObj["code"].(string)
	return code
}
