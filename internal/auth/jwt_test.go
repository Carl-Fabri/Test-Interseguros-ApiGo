package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	testSecret = strings.Repeat("k", 32)
	fixedNow   = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
)

func newManager(now time.Time) *JWTManager {
	return NewJWTManager(testSecret, "api-go", "matrix-services", time.Hour, func() time.Time { return now })
}

func TestIssueAndVerify(t *testing.T) {
	m := newManager(fixedNow)

	token, expiresAt, err := m.Issue("admin")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if !expiresAt.Equal(fixedNow.Add(time.Hour)) {
		t.Errorf("expiresAt = %v, want %v", expiresAt, fixedNow.Add(time.Hour))
	}

	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.Subject != "admin" {
		t.Errorf("subject = %q, want admin", claims.Subject)
	}
}

func TestVerifyRejects(t *testing.T) {
	valid, _, _ := newManager(fixedNow).Issue("admin")
	otherSecret, _, _ := NewJWTManager(strings.Repeat("x", 32), "api-go", "matrix-services", time.Hour, func() time.Time { return fixedNow }).Issue("admin")
	otherAudience, _, _ := NewJWTManager(testSecret, "api-go", "otra", time.Hour, func() time.Time { return fixedNow }).Issue("admin")
	noneAlg, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "admin"}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	tests := []struct {
		name  string
		token string
		now   time.Time
	}{
		{"token expirado", valid, fixedNow.Add(2 * time.Hour)},
		{"firmado con otro secreto", otherSecret, fixedNow},
		{"audiencia distinta", otherAudience, fixedNow},
		{"algoritmo none", noneAlg, fixedNow},
		{"token mal formado", "no-es-un-jwt", fixedNow},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newManager(tc.now).Verify(tc.token); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("err = %v, want ErrInvalidToken", err)
			}
		})
	}
}

func TestTokenContext(t *testing.T) {
	if _, ok := TokenFromContext(context.Background()); ok {
		t.Error("un contexto vacío no debería tener token")
	}
	token, ok := TokenFromContext(WithToken(context.Background(), "abc"))
	if !ok || token != "abc" {
		t.Errorf("TokenFromContext = %q, %v; want abc, true", token, ok)
	}
}
