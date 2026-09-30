# Data Model: Factorización QR de matrices

**Feature**: `002-factorizacion-qr`

## Matrix (dominio)

`[][]float64` en orden de filas. Reglas de `Validate(maxDimension)`:

| Regla | Error |
|---|---|
| Al menos una fila | `la matriz no puede estar vacía` |
| Filas no vacías | `las filas de la matriz no pueden estar vacías` |
| Filas y columnas ≤ `maxDimension` | `la matriz excede el tamaño máximo de NxN` |
| Todas las filas con el mismo largo | `la matriz debe ser rectangular: la fila i ...` |
| Valores finitos (sin NaN ni ±Inf) | `el valor en [i][j] no es un número finito` |

## QRResult (dominio)

| Campo | Forma | Propiedades |
|---|---|---|
| `Q` | m×m | Ortogonal (QᵀQ = I) |
| `R` | m×n | Triangular superior, `R[i][j] = 0` exacto para i > j, `R[k][k] ≥ 0` |

## Statistics (puerto hacia api-node)

| Campo | Tipo |
|---|---|
| `Max`, `Min`, `Average`, `Sum` | float64 |
| `Count` | int |
| `AnyDiagonal` | bool |
| `Matrices` | `[]{Name, IsDiagonal}` |

## DTOs HTTP

- **QRRequest**: `{ "matrix": number[][] }`
- **QRResponse**: `{ "q": number[][], "r": number[][], "statistics": StatisticsDTO }`
- **Petición a api-node**: `{ "matrices": [ { "name": "Q", "values": [[...]] }, { "name": "R", "values": [[...]] } ] }`

## Flujo de estados de una petición

```text
recibida ─JWT inválido─▶ 401
   │
   ├─JSON inválido─▶ 400 INVALID_REQUEST
   ├─matriz inválida─▶ 400 INVALID_MATRIX
   ▼
factorizada (QRGivens) ─api-node caído/error─▶ 502 · ─timeout─▶ 504
   ▼
200 { q, r, statistics }
```
