# Feature Specification: Autenticación JWT

**Feature Branch**: `001-autenticacion-jwt`

**Created**: 2026-09-29

**Status**: Implemented

**Input**: User description: "Aplicar un nivel de seguridad utilizando JWT para proteger las consultas a las APIs. La API en Go emite el token y la API en Node.js debe aceptarlo."

## Clarifications

### Session 2026-09-29

- Q: ¿Quién emite el token y cómo se autentica el usuario? → A: api-go expone `POST /api/v1/auth/login` con un usuario de demostración definido por variables de entorno (el reto no pide gestión de usuarios).
- Q: ¿Cómo llega la autorización a api-node? → A: api-go reenvía el **mismo** token del usuario; api-node lo valida con el mismo secreto, emisor y audiencia.
- Q: ¿Algoritmo de firma? → A: HS256 con secreto compartido de al menos 32 caracteres (RFC 7518); algoritmo fijado al validar.
- Q: ¿Vigencia del token? → A: configurable con `JWT_TTL_MINUTES`, 60 minutos por defecto; `exp` obligatorio.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Obtener un token de acceso (Priority: P1)

Un cliente (el frontend o una persona con curl) envía usuario y contraseña y recibe un token de acceso con su tiempo
de vigencia, que usará en las siguientes consultas.

**Why this priority**: sin token no se puede usar ningún endpoint protegido; es la puerta de entrada del sistema.

**Independent Test**: `POST /api/v1/auth/login` con las credenciales de demo devuelve 200 con `accessToken`,
`tokenType: "Bearer"` y `expiresIn > 0`; con credenciales incorrectas devuelve 401.

**Acceptance Scenarios**:

1. **Given** credenciales válidas, **When** el cliente llama al login, **Then** recibe 200 con un JWT HS256 que contiene `sub`, `iss`, `aud`, `iat`, `nbf` y `exp`.
2. **Given** una contraseña incorrecta, **When** llama al login, **Then** recibe 401 `INVALID_CREDENTIALS` sin revelar si el usuario existe.
3. **Given** un cuerpo vacío o JSON mal formado, **When** llama al login, **Then** recibe 400 `INVALID_REQUEST`.

---

### User Story 2 - Proteger los endpoints de negocio (Priority: P1)

Los endpoints que procesan matrices solo responden a clientes con un token vigente emitido por api-go.

**Why this priority**: es el requisito de seguridad explícito del reto.

**Independent Test**: `POST /api/v1/matrix/qr` sin token, con token mal formado, expirado, de otra audiencia o firmado
con otro secreto devuelve 401 y no llama a api-node.

**Acceptance Scenarios**:

1. **Given** un token vigente, **When** el cliente llama a un endpoint protegido con `Authorization: Bearer <token>`, **Then** la petición se procesa.
2. **Given** un token ausente, inválido, expirado o con `alg: none`, **When** llama al endpoint, **Then** recibe 401 `UNAUTHORIZED`.

---

### User Story 3 - Propagar la identidad a api-node (Priority: P2)

Cuando api-go consulta a api-node en nombre del usuario, reenvía el mismo token para que api-node aplique su propia
autorización sin un segundo mecanismo.

**Why this priority**: mantiene ambos servicios protegidos sin acoplar código entre ellos.

**Independent Test**: en la prueba de integración, el api-node simulado recibe `Authorization: Bearer <token del usuario>`.

**Acceptance Scenarios**:

1. **Given** una petición autenticada a `/matrix/qr`, **When** api-go llama a api-node, **Then** envía el mismo header `Authorization`.

### Edge Cases

- Token sin `exp` → rechazado (expiración obligatoria).
- Token firmado con HS512 o `alg: none` → rechazado (algoritmo fijado a HS256).
- `JWT_SECRET` ausente o de menos de 32 caracteres → el servicio no arranca y lista el error.
- Esquema distinto de `Bearer` (p. ej. `Basic`) → 401.
- Reloj del token en el futuro (`nbf`) → rechazado hasta su vigencia.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST emitir un JWT HS256 al recibir credenciales válidas del usuario de demostración.
- **FR-002**: El token MUST incluir `sub`, `iss` (`JWT_ISSUER`), `aud` (`JWT_AUDIENCE`), `iat`, `nbf` y `exp` (`JWT_TTL_MINUTES`).
- **FR-003**: El sistema MUST comparar las credenciales en tiempo constante y responder igual ante usuario o contraseña incorrectos.
- **FR-004**: Todo endpoint de negocio MUST exigir `Authorization: Bearer <token>` válido.
- **FR-005**: La validación MUST fijar el algoritmo HS256 y exigir emisor, audiencia y expiración.
- **FR-006**: El sistema MUST reenviar el token del usuario en las llamadas a api-node.
- **FR-007**: El sistema MUST negarse a arrancar sin `JWT_SECRET` (≥ 32 caracteres), `AUTH_USERNAME` y `AUTH_PASSWORD`.
- **FR-008**: Los errores de autenticación MUST usar el formato único de error con códigos `UNAUTHORIZED` / `INVALID_CREDENTIALS`.

### Key Entities

- **Credenciales**: usuario y contraseña de demostración (configuración, no persistidas).
- **Token de acceso**: JWT firmado con su fecha de expiración; identifica al usuario (`sub`).
- **Claims**: `sub`, `iss`, `aud`, `iat`, `nbf`, `exp`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las peticiones a endpoints de negocio sin token válido se rechazan con 401 (verificado por pruebas de integración).
- **SC-002**: El login responde en menos de 50 ms en local.
- **SC-003**: Un token emitido por api-go es aceptado por api-node sin configuración adicional más allá del secreto compartido.
- **SC-004**: Ninguna prueba ni log contiene el token o la contraseña en claro.

## Assumptions

- Un único usuario de demostración es suficiente para el alcance del reto; un proveedor de identidad real queda fuera del alcance.
- api-go y api-node comparten `JWT_SECRET`, `JWT_ISSUER` y `JWT_AUDIENCE` (en AWS, desde Secrets Manager).
- No se implementa refresh token ni revocación: la vigencia corta (60 min) acota el riesgo.
