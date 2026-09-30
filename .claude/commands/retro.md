---
description: Cierra la sesión verificando la definición de terminado y consolidando lo aprendido
---

## 1. Verificar la definición de terminado
Revisa los cambios de esta sesión en este servicio y confirma cada punto (✅ o "no aplica: <motivo>"):
- **Pruebas**: ¿toda lógica nueva o modificada tiene pruebas unitarias (e integración si cambió un endpoint)? Ejecútalas.
- **Documentación**: ¿están al día la documentación en el código, `.claude/CLAUDE.md`, `README.md` y `api/openapi.yaml`?
- **Docker**: ¿`Dockerfile`, `docker-compose.yml`, `.env.example` y `.dockerignore` reflejan las dependencias, variables y puertos actuales?
Si falta algo, dilo y propón cómo completarlo antes de seguir.

## 2. Consolidar lo aprendido
Clasifica cada aprendizaje:
1. **Regla o convención que siempre debe cumplirse** → propón el cambio a `.claude/CLAUDE.md` o a `.claude/rules/` (muéstrame el diff antes de aplicarlo).
2. **Decisión de arquitectura o de contrato** → propón un ADR en `docs/architecture/` o un cambio en `api/openapi.yaml`.
3. **Contexto o preferencia que no se deduce del código** → guárdalo en la memoria.
4. **Algo que ya está en el código o en git** → no lo guardes.

Termina con el checklist del paso 1 y una lista breve de lo guardado y lo propuesto.
