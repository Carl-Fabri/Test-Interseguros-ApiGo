<!--
Sync Impact Report
- Version change: plantilla → 1.0.0 (ratificación inicial)
- Principios definidos: I. Servicio autónomo · II. Contrato primero · III. Dominio puro y arquitectura limpia ·
  IV. Pruebas obligatorias · V. Seguridad por defecto · VI. Simplicidad y costo mínimo
- Secciones agregadas: Restricciones técnicas · Flujo de desarrollo (SDD) · Gobierno
- Plantillas revisadas: ✅ plan-template.md (Constitution Check) · ✅ spec-template.md · ✅ tasks-template.md
- Pendientes: ninguno
-->

# Constitución de api-go

## Core Principles

### I. Servicio autónomo
api-go es un servicio independiente con su propio repositorio, código, configuración, imagen Docker, pruebas, CI y
despliegue. **No comparte código** con otros servicios: api-node se consume solo por HTTP según el contrato que
publica (`NODE_API_URL`). Ninguna funcionalidad puede requerir archivos de otro repositorio para construirse o
desplegarse.

### II. Contrato primero (NON-NEGOTIABLE)
Toda API pública se define en `api/openapi.yaml` **antes** de implementarse, y el contrato se actualiza en el mismo
cambio que modifica un endpoint. Los errores usan un formato único `{"error": {"code", "message"}}` con códigos
estables. Un cambio incompatible exige una nueva versión de ruta (`/api/v2`).

### III. Dominio puro y arquitectura limpia
Las dependencias apuntan hacia el dominio: `domain` (matemática pura) ← `service` (casos de uso y puertos) ←
adaptadores (`client`, `auth`) y entrega (`handler`, `middleware`, `model`). `domain` y `service` **no importan
Fiber**. Los errores de dominio se traducen a HTTP en un único lugar (`middleware/error_handler.go`).

### IV. Pruebas obligatorias (NON-NEGOTIABLE)
Toda lógica nueva o modificada lleva pruebas unitarias table-driven; todo endpoint nuevo o modificado lleva pruebas
de integración con la app completa y dependencias externas simuladas (`httptest`). La matemática se verifica por
propiedades (p. ej. Q·R ≈ A, QᵀQ ≈ I) con tolerancia, nunca con igualdad exacta de flotantes. El CI bloquea el
cambio si fallan formato, `go vet`, pruebas o build de la imagen.

### V. Seguridad por defecto
Toda ruta de negocio exige JWT (HS256 con algoritmo fijo, `iss`, `aud` y `exp` obligatorios). Los secretos solo
entran por variables de entorno o Secrets Manager y el servicio **no arranca** con un secreto de menos de 32
caracteres. Las entradas tienen límites de tamaño, las credenciales se comparan en tiempo constante y los errores
internos nunca se exponen al cliente.

### VI. Simplicidad y costo mínimo
Se elige la solución más simple que cumpla el requisito (YAGNI). Nuevas dependencias o infraestructura se justifican
por escrito (ADR o plan). En la nube se prioriza el menor costo razonable: recursos compartidos, Fargate Spot y CI
sin despliegues automáticos.

## Restricciones técnicas

- **Lenguaje y framework:** Go 1.27+, Fiber v3; biblioteca estándar siempre que sea suficiente.
- **Contenedor:** Dockerfile multi-stage con imagen final distroless no root.
- **Configuración:** solo variables de entorno, validadas al arrancar (`internal/config`).
- **Despliegue:** AWS ECS Fargate con CloudFormation (`deploy/aws/`), detrás del ALB compartido.
- **Documentación:** GoDoc en todo lo exportado; README y `.claude/CLAUDE.md` al día; decisiones relevantes como ADR.

## Flujo de desarrollo (SDD)

1. `/speckit-specify` → `specs/NNN-funcionalidad/spec.md` (qué y por qué, sin detalles de implementación).
2. `/speckit-clarify` → se resuelven las ambigüedades del enunciado antes de planificar.
3. `/speckit-plan` → `plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md` (pasa el Constitution Check).
4. `/speckit-tasks` → `tasks.md` ordenado por historia de usuario, con las pruebas primero.
5. `/speckit-implement` → implementación tarea por tarea hasta que todas las pruebas pasan.
6. **Definición de terminado:** pruebas, documentación (GoDoc, README, contrato) y Docker/despliegue actualizados.

## Governance

Esta constitución prevalece sobre cualquier otra práctica del repositorio. Todo plan (`plan.md`) debe pasar el
Constitution Check; una violación solo se acepta documentada en "Complexity Tracking" con la alternativa descartada.
Las enmiendas se registran con versionado semántico (MAJOR: se elimina o redefine un principio; MINOR: se agrega un
principio o sección; PATCH: aclaraciones) y se reflejan en las plantillas afectadas. La guía operativa del día a día
está en `.claude/CLAUDE.md`.

**Version**: 1.0.0 | **Ratified**: 2026-09-29 | **Last Amended**: 2026-09-30
