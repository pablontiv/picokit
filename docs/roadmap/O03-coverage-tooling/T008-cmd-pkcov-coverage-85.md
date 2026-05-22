---
estado: Completed
tipo: task
---
# T008: Subir cobertura de cmd/pkcov a ≥85% y eliminar de exclude

**Outcome**: [O03 Coverage tooling compartido](README.md)
**Contribuye a**: eliminar la última excepción del dogfood — todo paquete del módulo cumple el floor (INV3)

[[blocked_by:./T007-enforce-gate-en-ci.md]]

## Preserva

- INV3 del outcome: picokit cumple su propia spec
  - Verificar: tras esta task, `.coverage-floors.toml` no contiene `cmd/pkcov` en `exclude`, y `pkcov check` sigue saliendo 0 — todos los paquetes del módulo a ≥85%, sin excepciones.

## Contexto

T007 estableció el gate v1.1 en CI con una exclusión documentada: `cmd/pkcov` quedó al 70.5% tras T003 y excluirlo temporalmente fue la forma honesta de desbloquear CI sin bajar el listón (spec §2 monotonicidad).

`cmd/pkcov` es el binario que implementa el gate — que justamente el binario de la spec no cumpla su propia spec es la deuda más explícita posible en el repo. Esta task la cierra.

Camino esperado: combinación de (a) eliminar dead code per spec §6 (stubs, paths inalcanzables, helpers no usados), (b) añadir tests para los paths actualmente no cubiertos (típicamente: branches de error, output JSON, validación de flags). El target conservador es 85%; subir más es bienvenido per §2.

## Alcance

**In**:

1. Medir cobertura actual de `cmd/pkcov` con `go test ./cmd/pkcov/... -coverprofile=cov.out -count=1 && go tool cover -func=cov.out`. Identificar funciones <85% y los lines no cubiertos.

2. Para cada hueco:
   - Si la línea es dead code (stub que retorna zero value sin callers, branch inalcanzable, comentario placeholder) — eliminar per spec §6.
   - Si la línea es lógica viva — añadir tests en `cmd/pkcov/pkcov_test.go` (o nuevo `_test.go` por subcomando si crece) cubriéndola. Casos típicos:
     - Output JSON con violations.
     - Flag combinations (`--profile` inexistente, `--floors` inválido, `--module` con/sin trailing slash).
     - `resolveModule` con flag vacío + go.mod ausente.
     - Exit codes según hay/no hay violations.

3. Editar `.coverage-floors.toml`: borrar la línea `"cmd/pkcov",` del array `exclude` (o vaciar el array si era el único). Mantener el comentario explicando que `exclude` se reserva para casos excepcionales.

4. Verificación local:
   - `go test ./cmd/pkcov/... -cover` reporta ≥85%.
   - `just coverage-check` exit 0 con `cmd/pkcov` apareciendo como PASS.

**Out**:

- No tocar la implementación de `cmd/pkcov` para añadir features. Solo tests + dead code removal.
- No tocar `coverage/`.
- No tocar el spec ni el workflow CI (ya están).

## Estado inicial esperado

- T007 completada: gate v1.1 corriendo en CI, `cmd/pkcov` excluido temporalmente.
- `cmd/pkcov` al 70.5% (rango ~70% — re-medir antes de empezar).
- `.coverage-floors.toml` contiene `exclude = ["cmd/pkcov"]`.

## Criterios de Aceptación

- `go test ./cmd/pkcov/... -cover` reporta ≥85% statement coverage.
- Cualquier código eliminado como dead code se borra (no comentado) per spec §6.
- `.coverage-floors.toml` no contiene `cmd/pkcov` en `exclude`.
- `just coverage-check` exit 0; `cmd/pkcov` aparece como PASS.
- CI completo en verde tras push.

## Fuente de verdad

- `/home/shared/picokit/cmd/pkcov/` — paquete a cubrir
- `/home/shared/picokit/.coverage-floors.toml` — quitar la exclusión
- `/home/shared/picokit/docs/coverage-spec.md` §6 — política de dead code
