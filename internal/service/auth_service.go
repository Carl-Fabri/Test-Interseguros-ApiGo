package service

import (
	"crypto/sha256"
	"crypto/subtle"
	"time"
)

// AccessToken es un token emitido junto con su vigencia.
type AccessToken struct {
	Token     string
	ExpiresAt time.Time
}

// AuthService autentica al usuario de demostración configurado y emite su JWT.
//
// Decisión: el reto no pide gestión de usuarios, así que se usa un único usuario
// definido por variables de entorno. Reemplazarlo por un proveedor de identidad
// solo requiere otra implementación de este caso de uso.
type AuthService struct {
	usernameHash [32]byte
	passwordHash [32]byte
	issuer       TokenIssuer
}

// NewAuthService crea el servicio con las credenciales válidas y el emisor de tokens.
func NewAuthService(username, password string, issuer TokenIssuer) *AuthService {
	return &AuthService{
		usernameHash: sha256.Sum256([]byte(username)),
		passwordHash: sha256.Sum256([]byte(password)),
		issuer:       issuer,
	}
}

// Login valida las credenciales en tiempo constante (compara hashes de igual largo
// para no filtrar información por tiempo de respuesta) y emite un token.
func (s *AuthService) Login(username, password string) (AccessToken, error) {
	userHash := sha256.Sum256([]byte(username))
	passHash := sha256.Sum256([]byte(password))

	// ! Comparación en tiempo constante sobre hashes de igual largo: no filtra por tiempo cuántos caracteres coinciden.
	// ! Se evalúan AMBAS comparaciones siempre (sin cortocircuito) para no revelar si el usuario existe.
	userOK := subtle.ConstantTimeCompare(userHash[:], s.usernameHash[:])
	passOK := subtle.ConstantTimeCompare(passHash[:], s.passwordHash[:])
	if userOK&passOK != 1 {
		return AccessToken{}, ErrInvalidCredentials
	}

	token, expiresAt, err := s.issuer.Issue(username)
	if err != nil {
		return AccessToken{}, err
	}
	return AccessToken{Token: token, ExpiresAt: expiresAt}, nil
}
