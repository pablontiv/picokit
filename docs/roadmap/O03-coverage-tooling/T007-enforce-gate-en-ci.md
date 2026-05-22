---
estado: Completed
tipo: task
---
# T007: Cerrar enforcement loop — go.mod tidy + Justfile go run + CI coverage-gate

**Outcome**: [O03 Coverage tooling compartido](README.md)
**Contribuye a**: el gate del coverage-spec se ejecuta de hecho en CI y localmente sin instalación previa (INV3)

[[blocked_by:./T006-implement-auto-discovery.md]]

## Preserva

- INV3 del outcome: picokit cumple su propia spec
  - Verificar: tras esta task, `pkcov check` corre en CI como step real y en el pre-push hook funciona sin necesitar `go install pkcov`. El job CI Tidy queda verde.

## Contexto

Tras T004 el dogfood quedó incompleto en dos formas que sólo se hicieron visibles después de mergeado:

1. **CI Tidy job rojo** desde commit `a36b023`: `cobra` y `go-toml/v2` aparecen como `// indirect` en `go.mod` pero el CLI `pkcov` las usa directamente. `go mod tidy` produce un diff que `git diff --exit-code go.mod go.sum` (paso del workflow `crossbeam/go-ci.yml@v1`) rechaza.

2. **El gate no corre en CI**. `crossbeam/go-ci.yml@v1` acepta `coverage-threshold: 85` como input genérico pero no consume `.coverage-floors.toml` ni invoca `pkcov`. Localmente, `just coverage-check` falla con `pkcov: command not found` porque el Justfile invoca `pkcov` esperándolo en PATH; el hook pre-push hereda el mismo problema (depende de `just coverage-check`).

Esta task cierra ambos. La solución para (2) es invocar `pkcov` vía `go run ./cmd/pkcov` desde el propio repo, que sigue siendo "invocar pkcov" (spec §4) pero elimina la dependencia de instalación. Para consumidores externos de picokit, ellos seguirán `go install github.com/pablontiv/picokit/cmd/pkcov@latest`.

Decisión consciente sobre `cmd/pkcov` (actualmente 70.5%): se añade a `exclude = ["cmd/pkcov"]` con comentario justificando la exclusión como deuda temporal. La task T008 cierra ese hueco. Excluir explícitamente es preferible a invisibilidad silenciosa: el archivo TOML hace visible la deuda y T008 la rastrea.

## Alcance

**In**:

1. `go mod tidy` en la raíz del repo. Verificar que `cobra` y `go-toml/v2` se promueven a un bloque `require ( ... )` directo. Commitear `go.mod` y `go.sum` resultantes.

2. `Justfile`:
   - `coverage` recipe: cambiar `pkcov report ...` → `go run ./cmd/pkcov report ...`
   - `coverage-check` recipe: cambiar `pkcov check ...` → `go run ./cmd/pkcov check ...`

3. `.coverage-floors.toml`:
   ```toml
   default = 85

   # v1.1: el gate aplica a todos los paquetes del módulo.
   # cmd/pkcov está temporalmente excluido (70.5% < 85). T008 cierra esta deuda.
   exclude = [
     "cmd/pkcov",
   ]
   ```
   Quitar el campo `packages = [...]`.

4. `.github/workflows/ci.yml`:
   - Añadir job `coverage-gate` (en el mismo workflow, después o en paralelo a `ci`):
     ```yaml
     coverage-gate:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version-file: go.mod
         - run: go test ./... -coverprofile=coverage.out -count=1
         - run: go run ./cmd/pkcov check --profile coverage.out --floors .coverage-floors.toml
     ```
   - Actualizar el llamado a `crossbeam/go-release.yml@v1`: añadir `"coverage-gate"` a `quality-gate-jobs` y a `needs: [ci, gitleaks, coverage-gate]`.

5. Verificación local antes de push:
   - `git diff go.mod go.sum` → vacío tras `go mod tidy`.
   - `just coverage-check` → exit 0; output muestra PASS para los 8 paquetes (cmd/pkcov no aparece — excluido).
   - **Test negativo**: subir temporalmente `default` a 99 en `.coverage-floors.toml`, correr `just coverage-check` → exit 1 con violations. Revertir.

**Out**:

- No subir cobertura de `cmd/pkcov` (T008).
- No tocar `crossbeam` (warning de `actions/setup-go` con ambos `go-version` y `go-version-file` está en crossbeam, no aquí).
- No cambiar la versión del action `crossbeam/go-ci.yml@v1` (el coverage-threshold genérico que pasa queda como redundancia barata).

## Estado inicial esperado

- T006 completada: `pkcov` ya soporta auto-discovery + exclude.
- `go.mod` tiene `cobra`/`go-toml/v2` como `// indirect`.
- `Justfile` invoca `pkcov` desnudo.
- `.coverage-floors.toml` lista los 8 paquetes vía `packages = [...]`.
- `.github/workflows/ci.yml` solo delega a crossbeam reusable workflows.
- CI rama master en rojo por el Tidy step desde `a36b023`.

## Criterios de Aceptación

- `git diff go.mod go.sum` sale vacío después de `go mod tidy` (CI Tidy job pasa).
- `just coverage-check` exit 0 local (con `go run ./cmd/pkcov`).
- `.coverage-floors.toml` no contiene `packages = [...]`; contiene `exclude = ["cmd/pkcov"]` con comentario justificando.
- `.github/workflows/ci.yml` define job `coverage-gate` que invoca `go run ./cmd/pkcov check`.
- `coverage-gate` aparece en `quality-gate-jobs` del release job.
- Tras push, CI run completa en verde para todos los jobs.

## Fuente de verdad

- `/home/shared/picokit/go.mod`, `go.sum`
- `/home/shared/picokit/Justfile`
- `/home/shared/picokit/.coverage-floors.toml`
- `/home/shared/picokit/.github/workflows/ci.yml`
- `/home/shared/picokit/docs/coverage-spec.md` v1.1 (post-T005)
