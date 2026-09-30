---
paths:
  - "**/*_test.go"
  - "test/**/*"
---

# Pruebas en api-go
- Table-driven con `t.Run(tc.name, ...)`; nombres de caso descriptivos en español.
- Comparar flotantes con tolerancia (p. ej. `math.Abs(a-b) < 1e-9`), nunca con `==`.
- Unitarias junto al código (`internal/**/x_test.go`); integración de endpoints en `test/integration/` usando `app.Test`.
- El cliente hacia api-node se simula con una implementación falsa de su interfaz; las pruebas nunca llaman a la red.
