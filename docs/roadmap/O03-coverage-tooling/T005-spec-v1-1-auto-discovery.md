---
estado: Specified
tipo: task
---
# T005: Bump coverage-spec a v1.1 — auto-discovery + exclude

**Outcome**: [O03 Coverage tooling compartido](README.md)
**Contribuye a**: cerrar el hueco de v1.0 donde `packages = [...]` permite que un paquete del módulo quede invisible al gate (INV2, INV3)

## Preserva

- INV2 del outcome: la spec describe el qué, el código describe el cómo
  - Verificar: la spec describe el comportamiento de auto-discovery sin asumir detalle de implementación Go
- INV3 del outcome: picokit cumple su propia spec
  - Verificar: tras v1.1, dogfooding pasa a aplicar el floor a todo el módulo (visible vía la diferencia en `.coverage-floors.toml`)

## Contexto

El commit `a36b023` dogfooded v1.0 en picokit con `.coverage-floors.toml` listando 8 paquetes. `cmd/pkcov` (70.5%) quedó fuera de la lista — el spec v1.0 §3 lo permite ("Packages absent from the list are not checked"), pero contradice la intención del usuario de "85% uniforme para todo el módulo". La ambigüedad entre §1 ("on each listed package") y §3 (lista explícita) hizo que el dogfood pareciera completo cuando estructuralmente tenía un hueco.

v1.1 elimina la ambigüedad: el gate aplica a TODO paquete del módulo descubierto en el coverage profile, con opt-out explícito mediante `exclude = [...]`. Es un cambio MINOR (aditivo) per spec §8 — implementaciones v1.0 siguen siendo compliant, pero v1.1 cierra la puerta a omisiones silenciosas.

Esta task es solo el documento. El código que lo implementa vive en T006.

## Alcance

**In**: editar `/home/shared/picokit/docs/coverage-spec.md`:

1. Frontmatter: `version: v1.1`.
2. **§1 Threshold uniforme** — reescribir: el floor aplica a "cada paquete del módulo descubierto en el coverage profile" (no "each listed package").
3. **§3 Schema de `.coverage-floors.toml`** — rewrite:
   - `default` (integer, required): sin cambios.
   - `packages` (deprecated en v1.1): aceptado por implementaciones v1.1 por retrocompatibilidad con configs v1.0, pero su contenido se ignora. No emitir error si está presente.
   - `exclude` (array of strings, optional): suffixes de import path a excluir del gate. Cada exclusión documenta deuda: el motivo debe quedar en el commit o como comentario en el TOML.
4. **§7 Paquetes test-only** — clarificar: la detección automática vía 0 statements en el profile sigue siendo el mecanismo principal; `exclude` es ortogonal y se reserva para casos excepcionales (scaffolding, CLIs experimentales).
5. **§8 Versioning** — actualizar el ejemplo a `v1.1`. Añadir nota explícita: `v1.0 → v1.1` es MINOR (aditivo, no rompe consumidores v1.0).
6. **§9 Migration v1.0 → v1.1** (sección nueva): "Borrar el array `packages`. El gate aplicará automáticamente a todos los paquetes del módulo. Si algún paquete debe excluirse temporalmente, añadirlo a `exclude = [...]` con justificación."
7. Actualizar `/home/shared/picokit/README.md` sección "Coverage tooling": referencia a `coverage-spec v1.1`, ejemplo de `.coverage-floors.toml` sin `packages`.

**Out**:

- No cambios en `coverage/` ni `cmd/pkcov/` (eso es T006).
- No modificar `.coverage-floors.toml` (eso es T007).
- No `go install pkcov`, no Justfile, no CI (T007).

## Estado inicial esperado

- `docs/coverage-spec.md` existe con `version: v1.0` y las 8 secciones de T001.
- README sección "Coverage tooling" referencia "coverage-spec v1.0".

## Criterios de Aceptación

- `docs/coverage-spec.md` frontmatter declara `version: v1.1`.
- §1 ya no contiene la frase "each listed package"; describe aplicación uniforme a todos los paquetes del módulo.
- §3 documenta `exclude` como opcional y `packages` como deprecated (compat con v1.0).
- §7 sigue describiendo el SKIP automático de test-only.
- §8 ejemplifica `v1.1` y nota cambio MINOR.
- §9 (nueva) contiene la guía de migración.
- README sección "Coverage tooling" referencia v1.1 con un ejemplo de `.coverage-floors.toml` que NO incluye `packages = [...]`.

## Fuente de verdad

- `/home/shared/picokit/docs/coverage-spec.md` — destino
- `/home/shared/picokit/README.md` — referencia secundaria
- T001 task — historial del spec v1.0
