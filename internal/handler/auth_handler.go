// Package handler contiene los handlers HTTP: traducen DTOs a llamadas de casos de uso
// y resultados a DTOs de respuesta. No contienen lógica de negocio.
package handler

import (
	"strings"
	"time"

	"github.com/elfab/retotecnico/api-go/internal/model"
	"github.com/elfab/retotecnico/api-go/internal/service"
	"github.com/gofiber/fiber/v3"
)

// Authenticator es el caso de uso de login que necesita el handler.
type Authenticator interface {
	Login(username, password string) (service.AccessToken, error)
}

// AuthHandler expone la autenticación por HTTP.
type AuthHandler struct {
	auth Authenticator
	now  func() time.Time
}

// NewAuthHandler crea el handler de autenticación.
func NewAuthHandler(auth Authenticator) *AuthHandler {
	return &AuthHandler{auth: auth, now: time.Now}
}

// Login maneja POST /api/v1/auth/login.
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "el cuerpo debe ser un JSON con username y password")
	}
	if strings.TrimSpace(req.Username) == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "username y password son obligatorios")
	}

	token, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(model.LoginResponse{
		AccessToken: token.Token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(token.ExpiresAt.Sub(h.now()).Seconds()),
	})
}
