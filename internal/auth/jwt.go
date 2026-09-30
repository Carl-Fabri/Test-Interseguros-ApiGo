// Package auth emite y valida JWT (HS256) y transporta el token por el contexto de la petición.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken indica un token ausente, mal formado, expirado o con claims inválidos.
var ErrInvalidToken = errors.New("token inválido o expirado")

// Claims son los claims que api-go incluye en sus tokens.
type Claims struct {
	jwt.RegisteredClaims
}

// JWTManager firma y valida tokens con un secreto compartido.
// El mismo secreto, emisor y audiencia se configuran en api-node para validar el token reenviado.
type JWTManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

// NewJWTManager crea un JWTManager. now permite inyectar el reloj en pruebas (nil usa time.Now).
func NewJWTManager(secret, issuer, audience string, ttl time.Duration, now func() time.Time) *JWTManager {
	if now == nil {
		now = time.Now
	}
	return &JWTManager{secret: []byte(secret), issuer: issuer, audience: audience, ttl: ttl, now: now}
}

// Issue firma un token para subject y devuelve el token y su fecha de expiración.
func (m *JWTManager) Issue(subject string) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(m.ttl)
	claims := Claims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   subject,
		Issuer:    m.issuer,
		Audience:  jwt.ClaimStrings{m.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("firmar token: %w", err)
	}
	return token, expiresAt, nil
}

// Verify valida firma, algoritmo (solo HS256), expiración, emisor y audiencia.
func (m *JWTManager) Verify(raw string) (*Claims, error) {
	claims := &Claims{}
	// ! WithValidMethods fija HS256: rechaza "alg: none" y la confusión de algoritmos (p. ej. RS256 con la clave como secreto).
	// * Emisor, audiencia y expiración obligatorios: un token de otro sistema o vencido nunca es válido.
	_, err := jwt.ParseWithClaims(raw, claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return claims, nil
}

type tokenKey struct{}

// WithToken devuelve un contexto que transporta el token crudo de la petición,
// para que los adaptadores salientes (cliente de api-node) puedan reenviarlo.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}

// TokenFromContext recupera el token guardado con WithToken.
func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey{}).(string)
	return token, ok && token != ""
}
