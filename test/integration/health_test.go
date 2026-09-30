package integration

import (
	"net/http"
	"testing"
)

func TestHealthIsPublic(t *testing.T) {
	env := newTestEnv(t, fakeNodeOK)

	// "/" también responde 200: es el health check por defecto de ECS Express Mode.
	for _, path := range []string{"/health", "/"} {
		resp, body := env.do(t, http.MethodGet, path, "", nil)
		if resp.StatusCode != http.StatusOK || body["status"] != "ok" {
			t.Fatalf("%s: status = %d, body = %v", path, resp.StatusCode, body)
		}
	}
}

func TestDocsAreServed(t *testing.T) {
	env := newTestEnv(t, fakeNodeOK)

	for path, contentType := range map[string]string{"/openapi.yaml": "application/yaml", "/docs": "text/html; charset=utf-8"} {
		resp, _ := env.do(t, http.MethodGet, path, "", nil)
		if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != contentType {
			t.Errorf("%s: status = %d, content-type = %q", path, resp.StatusCode, resp.Header.Get("Content-Type"))
		}
	}
}

func TestUnknownRouteReturnsStandardError(t *testing.T) {
	env := newTestEnv(t, fakeNodeOK)

	resp, body := env.do(t, http.MethodGet, "/no-existe", "", nil)
	if resp.StatusCode != http.StatusNotFound || errorCode(body) != "NOT_FOUND" {
		t.Fatalf("status = %d, body = %v", resp.StatusCode, body)
	}
}
