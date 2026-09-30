# Research: Autenticación JWT

**Feature**: `001-autenticacion-jwt` | **Date**: 2026-09-29

## R1. Biblioteca JWT

- **Decision**: `github.com/golang-jwt/jwt/v5`.
- **Rationale**: estándar de facto en Go, mantenida, con opciones explícitas `WithValidMethods`, `WithIssuer`,
  `WithAudience` y `WithExpirationRequired` que cubren FR-005 sin código propio.
- **Alternatives considered**: `gofiber/contrib/jwt` (acopla la validación a Fiber y dificulta reutilizarla fuera del
  handler); implementación propia (riesgo de errores criptográficos).

## R2. Algoritmo de firma

- **Decision**: HS256 con secreto compartido de ≥ 32 caracteres.
- **Rationale**: dos servicios bajo el mismo control; HMAC es simple y rápido. RFC 7518 recomienda una clave de al
  menos el tamaño del hash (256 bits).
- **Alternatives considered**: RS256/ES256 (mejor si hubiera muchos verificadores o terceros, pero exige gestionar
  pares de claves y JWKS; innecesario para dos servicios propios).

## R3. Propagación a api-node

- **Decision**: reenviar el mismo token del usuario (`Authorization: Bearer`), transportado en `context.Context`.
- **Rationale**: un solo mecanismo de autorización y trazabilidad del usuario de punta a punta; el servicio
  (`MatrixService`) no conoce el token.
- **Alternatives considered**: token servicio-a-servicio con otra audiencia (más seguro, pero duplica configuración;
  queda documentado como evolución); mTLS (excesivo para el alcance).

## R4. Comparación de credenciales

- **Decision**: SHA-256 de usuario y contraseña + `subtle.ConstantTimeCompare`, evaluando ambas siempre.
- **Rationale**: evita ataques de tiempo y no revela si el usuario existe; hashear iguala longitudes (requisito de
  `ConstantTimeCompare`).
- **Alternatives considered**: comparación directa `==` (filtra información por tiempo); bcrypt (útil con contraseñas
  almacenadas; aquí la contraseña viene de configuración).

## R5. Almacenamiento de credenciales

- **Decision**: variables de entorno (`AUTH_USERNAME`, `AUTH_PASSWORD`); en AWS la contraseña la genera Secrets Manager.
- **Rationale**: el reto no requiere gestión de usuarios; nada sensible en el repositorio.
- **Alternatives considered**: base de datos de usuarios / Cognito (fuera de alcance; el puerto `TokenIssuer` permite migrar).
