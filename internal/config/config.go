// Package config carga y valida la configuración de api-go desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// minJWTSecretLength es la longitud mínima del secreto HS256 (256 bits) recomendada por RFC 7518.
const minJWTSecretLength = 32

// Config agrupa la configuración del servicio.
type Config struct {
	Port string // Puerto HTTP en el que escucha el servicio.

	NodeAPIURL     string        // URL base de api-node.
	NodeAPITimeout time.Duration // Timeout de las llamadas HTTP a api-node.

	JWTSecret   string        // Secreto compartido para firmar y validar JWT (HS256).
	JWTIssuer   string        // Claim "iss" de los tokens emitidos.
	JWTAudience string        // Claim "aud" de los tokens emitidos.
	JWTTTL      time.Duration // Vigencia de los tokens emitidos.

	AuthUsername string // Usuario de demostración habilitado para el login.
	AuthPassword string // Contraseña del usuario de demostración.

	CORSAllowedOrigins []string // Orígenes permitidos (p. ej. el frontend).
	MatrixMaxDimension int      // Máximo de filas y de columnas aceptado por matriz.
}

// Load lee la configuración usando getenv (normalmente os.Getenv), aplica valores
// por defecto y valida los obligatorios. Devuelve todos los errores juntos.
func Load(getenv func(string) string) (Config, error) {
	var errs []error
	env := func(key, fallback string) string {
		if v := strings.TrimSpace(getenv(key)); v != "" {
			return v
		}
		return fallback
	}
	positiveInt := func(key string, fallback int) int {
		raw := env(key, "")
		if raw == "" {
			return fallback
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			errs = append(errs, fmt.Errorf("%s debe ser un entero positivo: %q", key, raw))
			return fallback
		}
		return n
	}

	cfg := Config{
		Port:               env("PORT", "8080"),
		NodeAPIURL:         strings.TrimRight(env("NODE_API_URL", "http://localhost:3000"), "/"),
		NodeAPITimeout:     time.Duration(positiveInt("NODE_API_TIMEOUT_MS", 5000)) * time.Millisecond,
		JWTSecret:          getenv("JWT_SECRET"),
		JWTIssuer:          env("JWT_ISSUER", "api-go"),
		JWTAudience:        env("JWT_AUDIENCE", "matrix-services"),
		JWTTTL:             time.Duration(positiveInt("JWT_TTL_MINUTES", 60)) * time.Minute,
		AuthUsername:       env("AUTH_USERNAME", ""),
		AuthPassword:       getenv("AUTH_PASSWORD"),
		CORSAllowedOrigins: splitList(env("CORS_ALLOWED_ORIGINS", "http://localhost:4200")),
		MatrixMaxDimension: positiveInt("MATRIX_MAX_DIMENSION", 100),
	}

	// ! Fallar al arrancar con un secreto débil es preferible a emitir tokens fáciles de forjar.
	if len(cfg.JWTSecret) < minJWTSecretLength {
		errs = append(errs, fmt.Errorf("JWT_SECRET es obligatorio y debe tener al menos %d caracteres", minJWTSecretLength))
	}
	if cfg.AuthUsername == "" || cfg.AuthPassword == "" {
		errs = append(errs, errors.New("AUTH_USERNAME y AUTH_PASSWORD son obligatorios"))
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}

// FromEnv carga la configuración desde las variables de entorno del proceso.
func FromEnv() (Config, error) {
	return Load(os.Getenv)
}

// splitList separa una lista por comas descartando elementos vacíos.
func splitList(raw string) []string {
	var out []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}
