# Feature Specification: Factorización QR de matrices

**Feature Branch**: `002-factorizacion-qr`

**Created**: 2026-09-29

**Status**: Implemented

**Input**: User description: "Una API en Go que reciba como entrada un array de arrays de números que represente una matriz rectangular y devuelva la factorización QR de dicha matriz. La API en Go realizará la rotación de la matriz y luego enviará los datos resultantes a la segunda API en Node.js."

## Clarifications

### Session 2026-09-29

- Q: El enunciado pide "rotación de la matriz" en la arquitectura y "factorización QR" en la funcionalidad; ¿son dos operaciones? → A: Una sola: la QR se calcula **mediante rotaciones de Givens**, que es literalmente rotar pares de filas. Un único endpoint `POST /api/v1/matrix/qr`.
- Q: ¿QR completa o reducida? → A: Completa: Q es m×m ortogonal y R es m×n triangular superior; funciona igual para m ≥ n y m < n.
- Q: ¿Qué se envía a api-node y qué se devuelve al cliente? → A: Se envían Q y R; el cliente recibe Q, R y las estadísticas que devuelve api-node.
- Q: ¿Qué pasa si api-node no está disponible? → A: La petición falla con 502 (caído/error) o 504 (timeout); no se devuelve una QR parcial porque el contrato promete los tres resultados juntos.
- Q: ¿Límite de tamaño? → A: Configurable con `MATRIX_MAX_DIMENSION`, 100 filas y columnas por defecto.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Factorizar una matriz rectangular (Priority: P1)

Un cliente autenticado envía una matriz (array de arrays de números) y recibe sus matrices Q y R tales que A = Q·R.

**Why this priority**: es la funcionalidad requerida central del reto.

**Independent Test**: `POST /api/v1/matrix/qr` con `[[12,-51,4],[6,167,-68],[-4,24,-41]]` devuelve R = `[[14,21,-14],[0,175,-70],[0,0,35]]` y una Q ortogonal.

**Acceptance Scenarios**:

1. **Given** una matriz cuadrada, alta (m > n) o ancha (m < n) válida, **When** el cliente la envía, **Then** recibe Q (m×m) y R (m×n) con Q·R ≈ A, QᵀQ ≈ I, R triangular superior y diagonal de R ≥ 0.
2. **Given** una matriz singular, **When** la envía, **Then** recibe igualmente una factorización válida (rango incompleto reflejado en la diagonal de R).
3. **Given** valores muy grandes (≈ 1e150), **When** la envía, **Then** no hay desbordamientos y Q·R ≈ A en términos relativos.

---

### User Story 2 - Validar la entrada (Priority: P1)

El servicio rechaza de forma clara las entradas que no representan una matriz rectangular válida.

**Why this priority**: protege la correctitud del cálculo y la estabilidad del servicio.

**Independent Test**: matrices vacías, no rectangulares, con valores no numéricos o mayores al límite devuelven 400 sin llamar a api-node.

**Acceptance Scenarios**:

1. **Given** filas de distinto largo, **When** se envía, **Then** 400 `INVALID_MATRIX` indicando la fila problemática.
2. **Given** JSON mal formado o valores no numéricos, **When** se envía, **Then** 400 `INVALID_REQUEST`.
3. **Given** más de `MATRIX_MAX_DIMENSION` filas o columnas, **When** se envía, **Then** 400 `INVALID_MATRIX`.

---

### User Story 3 - Obtener estadísticas de Q y R desde api-node (Priority: P2)

Tras factorizar, api-go envía Q y R a api-node y devuelve al cliente sus estadísticas junto con la factorización.

**Why this priority**: materializa la comunicación entre servicios pedida por la arquitectura.

**Independent Test**: con un api-node simulado, la respuesta incluye `statistics` y api-node recibe `{matrices:[{name:"Q"},{name:"R"}]}` con el token del usuario.

**Acceptance Scenarios**:

1. **Given** api-node disponible, **When** se factoriza, **Then** la respuesta incluye `statistics` (max, min, average, sum, count, anyDiagonal, matrices).
2. **Given** api-node responde con error o está caído, **When** se factoriza, **Then** 502 `STATISTICS_UNAVAILABLE`.
3. **Given** api-node excede `NODE_API_TIMEOUT_MS`, **When** se factoriza, **Then** 504 `STATISTICS_TIMEOUT`.

### Edge Cases

- Matriz de 1×1 (incluido un valor negativo): Q = [[1]] y R = [[|a|]] con el signo en Q.
- Fila o columna única: la QR sigue siendo válida (m×1 o 1×n).
- Elementos ya nulos bajo la diagonal: se omiten las rotaciones (cero exacto en R).
- Matriz diagonal positiva: Q = I y R = A (api-node detecta ambas como diagonales).
- `-0` en la salida: se normaliza a `0` para un JSON limpio.
- La matriz de entrada nunca se modifica.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST aceptar una matriz rectangular de números como `{"matrix": number[][]}`.
- **FR-002**: El sistema MUST devolver la factorización QR completa: Q m×m ortogonal y R m×n triangular superior con diagonal no negativa.
- **FR-003**: El cálculo MUST ser numéricamente estable para valores de gran magnitud (sin overflow en la norma).
- **FR-004**: El sistema MUST validar que la matriz sea no vacía, rectangular, finita y dentro del límite configurado.
- **FR-005**: El sistema MUST enviar Q y R a api-node por HTTP y devolver sus estadísticas en la respuesta.
- **FR-006**: El sistema MUST aplicar un timeout configurable a api-node y distinguir caída (502) de timeout (504).
- **FR-007**: El endpoint MUST estar protegido con JWT (ver `001-autenticacion-jwt`).
- **FR-008**: El contrato MUST publicarse en OpenAPI y verse en una referencia interactiva.

### Key Entities

- **Matriz A**: entrada m×n de números reales finitos.
- **Resultado QR**: matrices Q (m×m) y R (m×n) con A = Q·R.
- **Estadísticas**: resultado de api-node sobre Q y R (máximo, mínimo, promedio, suma, conteo, diagonalidad).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Para todas las matrices de prueba, ‖Q·R − A‖ ≤ 1e-9 (relativo para magnitudes grandes) y ‖QᵀQ − I‖ ≤ 1e-9.
- **SC-002**: Una matriz de 100×100 se factoriza en menos de 50 ms en el servidor.
- **SC-003**: El 100 % de las entradas inválidas se rechazan con 400 antes de llamar a api-node.
- **SC-004**: La cobertura de pruebas del dominio es ≥ 95 %.

## Assumptions

- Los números llegan como float64 IEEE-754 (JSON estándar).
- api-node expone `POST /api/v1/statistics` según su contrato OpenAPI y acepta el mismo JWT.
- No se persisten matrices ni resultados.
