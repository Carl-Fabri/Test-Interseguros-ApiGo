# ADR-002: Decisiones técnicas de la implementación

- **Estado:** Aceptado
- **Fecha:** 2026-09-30

Este documento registra las decisiones tomadas ante ambigüedades del enunciado y su justificación.

## 1. "Rotación de la matriz" vs. "factorización QR" → QR por rotaciones de Givens
**Problema:** la arquitectura pide que api-go "realice la rotación de la matriz" y la funcionalidad requerida pide
"la factorización QR". Son operaciones distintas si se lee "rotación" como girar 90°.

**Decisión:** un único endpoint `POST /api/v1/matrix/qr` que calcula la QR **mediante rotaciones de Givens**.
Así ambas frases describen lo mismo: la QR se obtiene rotando pares de filas hasta anular la parte inferior.

**Por qué Givens y no Householder o Gram-Schmidt:**
| Método                  | Estabilidad           | Comentario                                                        |
|-------------------------|-----------------------|-------------------------------------------------------------------|
| Gram-Schmidt clásico    | Pobre (pierde ortogonalidad) | Descartado                                                 |
| Householder             | Excelente             | ~1.5× menos operaciones en matrices densas                        |
| **Givens**              | Excelente             | Coincide con la "rotación" del enunciado; anula elementos de forma selectiva (salta los que ya son 0, ideal para matrices dispersas) |

Costo: O(m·n·min(m,n)) para R y O(m²·min(m,n)) para Q; con el límite de 100×100 responde en milisegundos.

**Detalles de la implementación:**
- **QR completa** (Q m×m, R m×n): funciona igual para matrices altas (m > n) y anchas (m < n).
- `math.Hypot` para calcular `r = √(x² + y²)` sin overflow ni underflow.
- Los elementos anulados se fijan en **0 exacto** (sin residuo de redondeo), así la detección de diagonal es confiable.
- Se **normalizan los signos** para que diag(R) ≥ 0: la QR es única con esa condición si A tiene rango completo,
  lo que hace la salida determinista y comparable con otras herramientas.
- Verificación en pruebas: Q·R ≈ A, QᵀQ ≈ I, R triangular superior, entrada no mutada, valores del orden de 1e150.

## 2. JWT: api-go emite y propaga el mismo token
- `POST /api/v1/auth/login` (api-go) valida un **usuario de demostración** definido por variables de entorno
  (el reto no pide gestión de usuarios) y emite un **JWT HS256** con `sub`, `iss`, `aud`, `iat`, `nbf` y `exp`.
- El cliente usa el token contra api-go; api-go **reenvía el mismo token** a api-node, que lo valida con el mismo
  secreto, emisor y audiencia. El frontend también puede llamar a api-node directamente con ese token.
- Seguridad: algoritmo fijado a HS256 en ambos lados (evita `alg: none` y confusión de algoritmo), `exp`
  obligatorio, secreto de al menos 32 caracteres validado al arrancar, credenciales comparadas en **tiempo
  constante** (hashes SHA-256 con `subtle.ConstantTimeCompare`), token **redactado** en los logs.
- **Alternativa descartada por ahora:** un token servicio-a-servicio distinto (otra audiencia). Es más seguro,
  pero duplica la configuración. Es la evolución natural si api-node pasa a tener más clientes.

## 3. Contratos OpenAPI propiedad de cada servicio + Scalar
- Cada servicio es dueño de su contrato: `api-go/api/openapi.yaml` y `api-node/openapi/openapi.yaml`.
  El proveedor publica su contrato en `/openapi.yaml`; los consumidores lo leen por HTTP. Así ningún servicio
  necesita archivos de fuera de su carpeta para construirse (ni siquiera la imagen Docker).
- Documentación interactiva con **Scalar** en `/docs`, usando la integración *standalone* oficial (HTML +
  script de CDN con versión fijada). Se descartó `@scalar/express-api-reference` porque solo publica ESM y el
  proyecto es CommonJS con Jest. Así la solución es idéntica en Go y en Node y no agrega dependencias.
- Contrato escrito a mano (*contract-first*) en lugar de generado por anotaciones: es la fuente de verdad del SDD.

## 4. Resiliencia en la comunicación HTTP
- Timeout configurable (`NODE_API_TIMEOUT_MS`). Timeout → **504** `STATISTICS_TIMEOUT`; caída o respuesta de
  error → **502** `STATISTICS_UNAVAILABLE`. Los errores internos nunca se exponen al cliente.
- Decisión: si api-node falla, la petición falla. No se devuelve una QR "parcial", porque el contrato promete Q, R
  y estadísticas juntas. Si se quisiera degradar, sería un campo `statistics: null` más un aviso.

## 5. Estadísticas en una sola pasada; definición de "diagonal"
- Máximo, mínimo, suma, conteo y diagonalidad se calculan recorriendo cada valor **una vez** (O(N)); el promedio
  sale de suma / conteo.
- **Diagonal:** todos los elementos con i ≠ j son 0, con tolerancia absoluta configurable (`DIAGONAL_TOLERANCE`,
  1e-10) para absorber el error de coma flotante (p. ej. 1e-17 en lugar de 0). Se acepta la **definición
  rectangular** (m×n) porque R es m×n; una matriz de 1×1 o nula es diagonal por definición.

## 6. Patrones aplicados
| Patrón              | api-go                                            | api-node                                   |
|---------------------|---------------------------------------------------|--------------------------------------------|
| Service Layer       | `internal/service` separa los casos de uso de HTTP | `src/services`                            |
| Dominio puro        | `internal/domain` (QR, validación)                | `src/domain` (estadísticas)                |
| DTOs                | `internal/model`                                  | `src/dtos` (validación + forma de salida)  |
| Middleware          | JWT, logger + requestid, recover, CORS, helmet, errores | JWT, pino-http, CORS, helmet, errores |
| Puertos/adaptadores | `StatisticsClient`, `TokenIssuer` (interfaces)    | Fábricas con inyección de dependencias     |
| Error único         | `{"error":{"code","message"}}` en ambos servicios |                                            |

## Consecuencias
- (+) Los cálculos se prueban sin HTTP; los adaptadores se reemplazan sin tocar el dominio.
- (+) Contratos y errores coherentes entre servicios.
- (−) El secreto compartido debe mantenerse sincronizado entre servicios (en la nube, desde un gestor de secretos).
