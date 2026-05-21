---
estado: Specified
tipo: task
---
# T003: Implementar CLI pkcov

**Outcome**: [O03 Coverage tooling compartido](README.md)
**Contribuye a**: binario consumible desde Justfiles y hooks (INV1)

[[blocked_by:./T002-port-coverage-package.md]]

## Preserva

- INV1 del outcome: pkcov produce output equivalente al script bash sobre los mismos inputs
  - Verificar: subcomando `report` y `check` reproducen líneas `PASS/FAIL/SKIP/TOTAL` del bash

## Contexto

El CLI es un wrapper delgado sobre `picokit/coverage` (T002). Sigue el patrón de `cmd/<name>/main.go` (primero en picokit; sirve de molde para futuros binarios). Cobra es la opción consistente con rootline/roadmapctl, que ya facilitan `--output json`.

Defaults razonables para uso desde Justfile sin args:
- `--profile coverage.out`
- `--floors .coverage-floors.toml`
- `--module` auto-detectado vía `coverage.DetectModulePrefix("go.mod")`

Output legible en stdout (sigue formato del bash de rootline para minimizar diff visual en los consumers). Flag `--output json` produce versión machine-readable.

## Alcance

**In**:

1. Crear `picokit/cmd/pkcov/main.go` y archivos por subcomando:
   - `pkcov report [--profile coverage.out] [--module <prefix>]` — imprime tabla per-package + total (mismo formato `PASS: pkg = N%` / `SKIP: ...` / `TOTAL: N%`)
   - `pkcov check [--profile coverage.out] [--floors .coverage-floors.toml] [--module <prefix>]` — exit 0 si verde, exit 1 con lista de violaciones si no
   - `pkcov version` — imprime version del binario + spec version implementada (`coverage-spec v1.0`)
   - `--output json` global produce JSON parseable con campos `total`, `per_package`, `violations`, `skipped`

2. Cobra setup mínimo (siguiendo el patrón de `cmd/rootline/root.go` en rootline para consistencia).

3. `go build ./cmd/pkcov/` produce binario funcional.

4. Tests para CLI:
   - Smoke: ejecutar `report` y `check` contra `testdata/coverage.out` (fixture de T002) y verificar exit codes + output esperado.
   - `--output json` produce JSON parseable con campos correctos.

**Out**:
- No instalar el binario globalmente (eso lo decide el repo consumidor con `go install` o vendoring).
- No implementar subcomandos más allá de los 3.
- No tocar Justfile / pre-push (T004).

## Estado inicial esperado

- T002 completada: `picokit/coverage/` compila y pasa tests.
- `cmd/pkcov/` no existe aún.

## Criterios de Aceptación

- `go build -o /tmp/pkcov ./cmd/pkcov/` exit 0.
- `/tmp/pkcov report --profile picokit/coverage/testdata/coverage.out --module github.com/pablontiv/rootline` imprime tabla con `PASS:`/`SKIP:`/`TOTAL:`.
- `/tmp/pkcov check ... --floors picokit/coverage/testdata/floors.toml` exit 0 (fixture en verde).
- `/tmp/pkcov check ...` con un floors.toml modificado (default=99) exit 1 + violación nombrada.
- `/tmp/pkcov check --output json ...` produce JSON parseable con `{"total":..., "per_package":{...}, "violations":[...]}`.
- `/tmp/pkcov version` imprime versión + `coverage-spec v1.0`.
- `golangci-lint run ./cmd/pkcov/` sin issues.

## Fuente de verdad

- `/home/shared/picokit/coverage/` (T002)
- `/home/shared/picokit/docs/coverage-spec.md` (T001)
- `/home/shared/rootline/cmd/rootline/root.go` — patrón cobra mínimo
- `/home/shared/picokit/cmd/pkcov/` (nuevo)
