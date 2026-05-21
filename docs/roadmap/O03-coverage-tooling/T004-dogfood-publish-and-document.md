---
estado: Completed
tipo: task
---
# T004: Dogfood + README + publicar tag

**Outcome**: [O03 Coverage tooling compartido](README.md)
**Contribuye a**: picokit cumple su propia spec; tag consumible por adoptantes (INV3)

[[blocked_by:./T003-implement-pkcov-cli.md]]

## Preserva

- INV3 del outcome: picokit cumple su propia spec
  - Verificar: `pkcov check` corriendo sobre picokit pasa con default=85; pre-push hook activo

## Contexto

picokit hoy tiene un recipe `coverage-summary` que solo imprime (sin gate). Esta task lo reemplaza por el contrato completo del coverage-spec v1.0: `coverage` + `coverage-check` recipes invocando el `pkcov` recién compilado, `.coverage-floors.toml`, gate pre-push, README actualizado, y tag publicado consumible por rootline/backscroll/roadmapctl.

picokit como dogfooder es crítico: si la herramienta no cumple su propia spec, el contrato pierde autoridad. Si algún paquete de picokit está por debajo de 85, esta task se bloquea hasta subirlo (es el mismo principio de pre-work que aplica a los repos consumidores).

## Alcance

**In**:

1. Medir cobertura actual de cada paquete picokit (`autoupdate`, `diag`, `diff`, `fuzzy`, `hashfile`, `output`, `pathsec`, `coverage`) con `go test ./... -coverprofile=coverage.out` + `go tool cover -func`. Si algún paquete está <85, subir antes (sub-paso interno de esta task — no se crea task aparte).

2. Crear `/home/shared/picokit/.coverage-floors.toml`:
   ```toml
   default = 85
   packages = [
     "autoupdate", "coverage", "diag", "diff",
     "fuzzy", "hashfile", "output", "pathsec",
   ]
   ```

3. Editar `/home/shared/picokit/Justfile`:
   - Borrar `coverage-summary` recipe existente
   - Agregar `coverage` recipe que corre `go test ... -coverprofile=coverage.out` y luego `pkcov report`
   - Agregar `coverage-check` recipe que corre tests + `pkcov check`

4. Editar `/home/shared/picokit/.githooks/pre-push` (o crear si no existe) con el bloque condicional sobre `*.go` que invoca `just coverage-check`, siguiendo el patrón de rootline (`/home/shared/rootline/.githooks/pre-push`).

5. Actualizar `/home/shared/picokit/README.md`:
   - Listar `coverage/` en la sección de paquetes
   - Sección breve sobre `pkcov` con ejemplos de invocación
   - Link a `docs/coverage-spec.md`

6. Verificación local: `just coverage-check` exit 0 sobre picokit completo.

7. Commit los archivos y publicar un tag versionado consumible vía `go get github.com/pablontiv/picokit@<tag>`. Convención del repo (semver). Recomendación: `v0.X.0` para esta minor (nueva funcionalidad).

**Out**:
- No instalar `pkcov` globalmente en el sistema.
- No modificar CI workflows de picokit (asumido que ya tienen su threshold via crossbeam u otro mecanismo; este task se enfoca en dogfooding local).
- No actualizar consumers (rootline/backscroll/roadmapctl); eso vive en sus propios outcomes.

## Estado inicial esperado

- T003 completada: `pkcov` binary construye y funciona.
- `picokit/Justfile` tiene recipe `coverage-summary` (medición sin gate).
- Cada paquete picokit tiene coverage medible.

## Criterios de Aceptación

- Cada paquete en `picokit/.coverage-floors.toml` reporta ≥85% en `pkcov check` local.
- `just coverage` imprime tabla y total.
- `just coverage-check` exit 0.
- `.githooks/pre-push` incluye bloque de coverage gate condicional sobre `*.go`.
- README lista el nuevo subpaquete + link a spec.
- Tag publicado y `go list -m github.com/pablontiv/picokit@<tag>` resuelve correctamente (verificable con `go get` desde otro repo).

## Fuente de verdad

- `/home/shared/picokit/Justfile`
- `/home/shared/picokit/.githooks/pre-push` (o nuevo)
- `/home/shared/picokit/README.md`
- `/home/shared/picokit/.coverage-floors.toml` (nuevo)
- `/home/shared/rootline/.githooks/pre-push` — patrón a imitar
