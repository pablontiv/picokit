---
estado: Completed
tipo: task
---
# T002: Portar el cálculo de coverage de bash a Go en picokit/coverage/

**Outcome**: [O03 Coverage tooling compartido](README.md)
**Contribuye a**: librería Go reutilizable que implementa la spec (INV1)

[[blocked_by:./T001-write-coverage-spec.md]]

## Preserva

- INV1 del outcome: pkcov es funcionalmente equivalente al bash de rootline sobre los mismos inputs
  - Verificar: tests con fixtures copiados de rootline producen mismo PASS/FAIL/SKIP

## Contexto

`/home/shared/rootline/scripts/check-coverage-floors.sh` (líneas 51-76) implementa el cálculo en awk:

```awk
/^mode:/ { next }
{
    split($1, parts, ":")
    filepath = parts[1]
    if (index(filepath, prefix) != 1) next
    stmts = $2
    count = $3
    total += stmts
    if (count > 0) covered += stmts
}
END {
    if (total == 0) print "SKIP"
    else printf "%.1f\n", 100 * covered / total
}
```

Se porta literalmente a Go. El único cambio funcional vs el script bash: línea 53 hardcodea `github.com/pablontiv/rootline/` como prefijo; en Go se lee de `go.mod` (campo `module`) o se acepta por parámetro.

Patrón de subpaquetes en picokit: ver `fuzzy/fuzzy.go` y `pathsec/pathsec.go` — pure Go libs sin dependencias externas (excepto stdlib). `coverage/` agrega `github.com/pelletier/go-toml/v2` como dependencia para parsear `.coverage-floors.toml`.

## Alcance

**In**:

1. Crear `picokit/coverage/coverage.go` con:
   - `type Profile struct { Packages map[string]Bucket }`
   - `type Bucket struct { Covered, Total int; Skipped bool }`
   - `func (Bucket) Percent() float64`
   - `func ParseProfile(profilePath, modulePrefix string) (*Profile, error)` — lee `coverage.out`, agrupa por prefijo
   - `func (*Profile) Total() float64` — coverage agregado, lo mismo que `go tool cover -func | grep total`
   - `func (*Profile) PerPackage() map[string]Bucket`

2. Crear `picokit/coverage/floors.go` con:
   - `type Floors struct { Default int; Packages []string }`
   - `func LoadFloors(tomlPath string) (*Floors, error)` — parsea TOML, valida campos requeridos
   - `type Violation struct { Package string; Got, Need float64 }`
   - `type Result struct { Total float64; PerPackage map[string]Bucket; Violations []Violation; SkippedPackages []string; MissingPackages []string }`
   - `func Check(p *Profile, f *Floors) Result`

3. Crear `picokit/coverage/module.go` con:
   - `func DetectModulePrefix(goModPath string) (string, error)` — lee `go.mod`, extrae línea `module github.com/foo/bar`, retorna `"github.com/foo/bar/"` (con slash final para uso directo como prefix)

4. Tests en `picokit/coverage/coverage_test.go`:
   - Fixture `testdata/coverage.out` copiada de un run reciente de rootline (`go test ./... -coverprofile=coverage.out` en rootline, copiar archivo)
   - Fixture `testdata/floors.toml` copiada de `/home/shared/rootline/.coverage-floors.toml`
   - Tests: parse profile produce conteo per-package correcto; check pasa con rootline floors; check falla cuando se inyecta una violación artificial (poner default=99 en floors fixture); DetectModulePrefix lee go.mod válido y rechaza inválido.

5. Cobertura propia del paquete ≥85%.

**Out**:
- No crear el CLI todavía (eso es T003).
- No tocar Justfile / pre-push aún (eso es T004).
- No implementar `--output json` (eso vive en el CLI).

## Estado inicial esperado

- T001 completada: `docs/coverage-spec.md` existe.
- `/home/shared/rootline/scripts/check-coverage-floors.sh` y `/home/shared/rootline/.coverage-floors.toml` accesibles como referencia.
- `picokit/go.mod` no tiene `pelletier/go-toml/v2` aún.

## Criterios de Aceptación

- `picokit/coverage/coverage.go`, `floors.go`, `module.go` existen y compilan.
- `picokit/coverage/coverage_test.go` con fixtures en `testdata/`.
- `go test ./coverage/... -cover` ≥85%.
- Equivalencia: `diff <(bash /home/shared/rootline/scripts/check-coverage-floors.sh testdata/coverage.out testdata/floors.toml | grep -E "^(PASS|FAIL|SKIP|TOTAL):") <(go test ./coverage/... -run TestEquivalence -v)` retorna las mismas líneas PASS/FAIL/SKIP/TOTAL (orden tolerante).
- `go.mod` incluye `github.com/pelletier/go-toml/v2`.
- `golangci-lint run` sin issues nuevos.

## Fuente de verdad

- `/home/shared/picokit/docs/coverage-spec.md` (T001)
- `/home/shared/rootline/scripts/check-coverage-floors.sh` — comportamiento de referencia
- `/home/shared/rootline/.coverage-floors.toml` — schema
- `/home/shared/picokit/fuzzy/fuzzy.go`, `pathsec/pathsec.go` — patrones de subpaquete
- `/home/shared/picokit/go.mod` — módulo destino
