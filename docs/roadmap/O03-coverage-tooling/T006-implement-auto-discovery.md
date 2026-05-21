---
estado: Completed
tipo: task
---
# T006: Implementar v1.1 en coverage/ y cmd/pkcov/

**Outcome**: [O03 Coverage tooling compartido](README.md)
**Contribuye a**: el comportamiento de pkcov coincide con la spec v1.1 (INV2)

[[blocked_by:./T005-spec-v1-1-auto-discovery.md]]

## Preserva

- INV1 del outcome: pkcov sigue siendo funcionalmente equivalente al gate de coverage por paquete
  - Verificar: tests existentes en coverage_test.go siguen pasando; nuevos tests cubren auto-discovery
- INV2 del outcome: el código implementa lo que la spec describe
  - Verificar: para una config v1.1 (sin `packages`, con `exclude`), el comportamiento sigue la spec literal

## Contexto

La librería `coverage/` y el CLI `cmd/pkcov/` implementan v1.0: `coverage.Check` itera `f.Packages` explícito, `coverage.LoadFloors` exige que `packages` no esté vacío, y el output de `pkcov check` se ordena según ese array. Para v1.1 hay que invertir la dirección: iterar todos los paquetes descubiertos en el `Profile` y filtrar por `exclude`.

El parseo del profile (`ParseProfile`) ya descubre todos los paquetes — no hay que añadir descubrimiento, solo cambiar quien itera. `ParseProfile` produce `Profile.Packages` con todos los paquetes con coverage data (incluyendo Total==0 que se marca Skipped).

Edge case explícito: configs v1.0 existentes (con `packages = [...]`) deben seguir cargando sin error. El campo se mantiene en el struct pero se ignora — el TOML lo acepta, el código no lo lee. Esto cumple §8 (cambio MINOR no rompe v1.0).

## Alcance

**In**:

1. `coverage/floors.go`:
   - Añadir `Exclude []string \`toml:"exclude"\`` a `Floors`.
   - Mantener `Packages []string \`toml:"packages,omitempty"\`` como campo deprecated (TOML lo acepta, el código no lo usa).
   - `LoadFloors`: validar `Default > 0`. **Eliminar** el error "packages list is empty".
   - `Check`: reescribir el loop. Iterar `perPkg` (resultado de `p.PerPackage()`). Para cada paquete: si está en `f.Exclude`, saltar; si `b.Skipped`, añadir a `SkippedPackages`; else evaluar `Percent() < threshold` y reportar violación.
   - `Result.MissingPackages` queda como campo vacío (no se elimina para no romper consumidores JSON externos).

2. `cmd/pkcov/check.go`:
   - Reescribir output de texto: iterar `r.PerPackage` ordenado (reusar `sortedKeys` que ya vive en `report.go` — moverlo a package-level si está fuera de scope). Misma lógica PASS/SKIP/FAIL.
   - JSON output: estructura existente sigue válida; `Skipped` se llena igual.

3. `cmd/pkcov/root.go`:
   - `const specVersion = "v1.1"`.

4. Tests:
   - `coverage/coverage_test.go` (o nuevo `floors_test.go` si no existe): casos para
     - config sin `packages` ni `exclude` (auto-discovery puro).
     - config con `exclude = ["pkg-a"]` (auto-discovery menos uno).
     - config legacy con `packages = ["pkg-a", "pkg-b"]` — debe cargar sin error y comportarse idéntico al caso sin `packages` (i.e. `packages` ignorado).
     - test-only package (0 statements) sigue siendo `Skipped`.
   - `cmd/pkcov/pkcov_test.go`: golden de output text/JSON para los casos nuevos.

**Out**:

- No tocar `.coverage-floors.toml` del repo (T007).
- No tocar `Justfile`, `.github/workflows/ci.yml`, `go.mod` (T007).
- No subir cobertura de `cmd/pkcov` (T008).

## Estado inicial esperado

- T005 completada: la spec describe v1.1.
- `coverage/floors.go`, `cmd/pkcov/check.go`, `cmd/pkcov/root.go` existen con la implementación v1.0.

## Criterios de Aceptación

- `coverage.Floors` tiene campo `Exclude []string`. `Packages` sigue existiendo pero está marcado como deprecated en comentario.
- `coverage.LoadFloors` no rechaza configs sin `packages`.
- `coverage.Check` aplica el threshold a todos los paquetes del profile, excluyendo los listados en `Exclude`.
- `pkcov check` con una config v1.1 imprime una línea por paquete del profile (PASS/SKIP/FAIL) ordenadas alfabéticamente.
- `pkcov check --version`/`pkcov check` reporta soporte de spec v1.1 (vía `specVersion`).
- `go test ./coverage/... ./cmd/pkcov/...` verde, sin regresiones.
- Test explícito demuestra que una config legacy con `packages = [...]` carga sin error y se comporta como sin `packages`.

## Fuente de verdad

- `/home/shared/picokit/coverage/floors.go`
- `/home/shared/picokit/coverage/coverage.go` — reusar `ParseProfile`, `PerPackage`, `Bucket.Skipped`
- `/home/shared/picokit/cmd/pkcov/check.go`, `report.go` (para `sortedKeys`), `root.go`
- `/home/shared/picokit/docs/coverage-spec.md` v1.1 (post-T005)
