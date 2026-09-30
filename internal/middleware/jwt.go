// Package middleware contiene los middlewares HTTP transversales de api-go.
package middleware

import (
	"strings"

	"github.com/elfab/retotecnico/api-go/internal/auth"
	"github.com/gofiber/fiber/v3"
)

// LocalSubject es la clave de c.Locals donde se guarda el "sub" del token validado.
const LocalSubject = "subject"

// TokenVerifier valida un JWT crudo. Lo implementa auth.JWTManager.
type TokenVerifier interface {
	Verify(raw string) (*auth.Claims, error)
}

// RequireJWT exige un header "Authorization: Bearer <token>" válido.
// Guarda el subject en c.Locals y el token crudo en el contexto para reenviarlo a api-node.
func RequireJWT(verifier TokenVerifier) fiber.Handler {
	return func(c fiber.Ctx) error {
		scheme, token, found := strings.Cut(c.Get(fiber.HeaderAuthorization), " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "se requiere un token Bearer en el header Authorization")
		}

		claims, err := verifier.Verify(strings.TrimSpace(token))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, auth.ErrInvalidToken.Error())
		}

		// * El token crudo viaja en el context.Context para que el cliente de api-node lo reenvíe (propagación del JWT).
		c.Locals(LocalSubject, claims.Subject)
		c.SetContext(auth.WithToken(c.Context(), strings.TrimSpace(token)))
		return c.Next()
	}
}
