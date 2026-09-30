# Research: Factorización QR de matrices

**Feature**: `002-factorizacion-qr` | **Date**: 2026-09-29

## R1. Algoritmo de factorización

- **Decision**: rotaciones de Givens sobre pares de filas adyacentes, de abajo hacia arriba en cada columna.
- **Rationale**: numéricamente estable (transformaciones ortogonales); resuelve la ambigüedad del enunciado
  ("rotación" y "QR" describen la misma operación); anula elementos de forma selectiva y omite los que ya son cero.
- **Alternatives considered**:
  | Método | Motivo del descarte |
  |---|---|
  | Gram-Schmidt clásico | Pierde ortogonalidad con matrices mal condicionadas |
  | Gram-Schmidt modificado | Mejor, pero sigue siendo menos estable que las transformaciones ortogonales |
  | Householder | Estable y ~1.5× menos operaciones en densas; no refleja la "rotación" del enunciado |
  | Biblioteca externa (gonum) | Dependencia pesada para un algoritmo de ~40 líneas; menos defendible en la entrevista |

## R2. QR completa vs. reducida

- **Decision**: completa (Q m×m, R m×n).
- **Rationale**: definida para cualquier forma (m ≥ n y m < n) sin casos especiales; Givens la produce de forma natural.
- **Alternatives considered**: reducida (Q m×n, R n×n): menos memoria, pero no está definida igual para matrices anchas.

## R3. Estabilidad numérica

- **Decision**: `math.Hypot(x, y)` para r; ceros exactos en los elementos anulados; normalización de signos (diag(R) ≥ 0); `-0 → 0`.
- **Rationale**: `sqrt(x²+y²)` desborda con valores ~1e200; los ceros exactos hacen que R sea estrictamente triangular
  (clave para la detección de diagonal en api-node); los signos normalizados hacen la salida única y comparable.

## R4. Comunicación con api-node

- **Decision**: `net/http` con `http.Client{Timeout}`, detrás del puerto `StatisticsClient`; errores envueltos en
  `ErrStatisticsUnavailable` (502) y `ErrStatisticsTimeout` (504).
- **Rationale**: la biblioteca estándar es suficiente; el puerto permite probar el servicio con dobles y el
  cliente con `httptest`.
- **Alternatives considered**: devolver la QR aunque api-node falle (`statistics: null`): descartado porque el
  contrato promete los tres resultados juntos; queda documentado como posible degradación futura.

## R5. Límites de entrada

- **Decision**: `MATRIX_MAX_DIMENSION` = 100 y `BodyLimit` = 1 MiB.
- **Rationale**: el costo es cúbico; 100×100 responde en milisegundos y acota el uso de CPU ante abuso.
