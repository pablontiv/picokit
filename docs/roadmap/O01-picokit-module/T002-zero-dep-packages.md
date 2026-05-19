---
estado: Completed
tipo: task
---
# T002: Portar paquetes sin dependencias externas (pathsec, diff, fuzzy, hashfile)

**Outcome**: [O01 Módulo picokit](README.md)
**Contribuye a**: 4 de los 7 paquetes disponibles, portados desde sus proyectos fuente

[[blocked_by:./T001-repo-setup.md]]

## Preserva

- INV1: Ningún paquete introduce dependencias externas nuevas en go.mod.
  - Verificar: `go mod tidy && git diff go.sum` no agrega módulos externos
- INV2: Las APIs exportadas son compatibles con los proyectos fuente.
  - Verificar: los tests cubren los mismos casos que los originales

## Contexto

Cuatro paquetes genéricos existen en el ecosistema y pueden portarse sin cambios de lógica:

- `pathsec/` ← roadmapctl `internal/fsx/`: path seguro anti-escape con `ResolveInside`
- `diff/` ← roadmapctl `internal/diff/`: generación de unified diff con `NewFile`, `UpdateFile`
- `fuzzy/` ← rootline `internal/fuzzy/`: fuzzy string matching con `Match`, `Distance`
- `hashfile/` ← backscroll `sync.HashFile` + roadmapctl `updater.writeAtomic`: SHA-256 y atomic write

Fuentes exactas:
- `/home/shared/roadmapctl/internal/fsx/` — pathsec
- `/home/shared/roadmapctl/internal/diff/` — diff
- `/home/shared/rootline/internal/fuzzy/` — fuzzy
- `/home/shared/backscroll/internal/sync/sync.go` función `HashFile`
- `/home/shared/roadmapctl/internal/updater/updater.go` función `writeAtomic`

## Alcance

**In**:
1. `pathsec/pathsec.go` — exportar `ResolveInside(root, candidate string) (abs, rel string, err error)`, `ErrPathEscape`, `ErrAbsolutePath`
2. `pathsec/pathsec_test.go` — cobertura ≥85%
3. `diff/diff.go` — exportar `NewFile(path, content string) string`, `UpdateFile(path, previous, content string) string`
4. `diff/diff_test.go` — cobertura ≥85%
5. `fuzzy/fuzzy.go` — exportar `Match(query string, candidates []string) []string`, `Distance(a, b string) int`
6. `fuzzy/fuzzy_test.go` — cobertura ≥85%
7. `hashfile/hashfile.go` — exportar `HashFile(path string) (string, error)` (SHA-256 hex) y `WriteAtomic(dest string, r io.Reader, mode os.FileMode) error`
8. `hashfile/hashfile_test.go` — cobertura ≥85%

**Out**:
- No modificar los proyectos fuente (backscroll, roadmapctl, rootline)
- No implementar autoupdate, diag ni output (eso es T003–T005)

## Estado inicial esperado

- T001 completada: `go.mod` existe con module path correcto, `just check` pasa

## Criterios de Aceptación

- `go test ./pathsec/... ./diff/... ./fuzzy/... ./hashfile/... -race` pasa
- Cobertura ≥85% por paquete: `go test ./... -cover`
- `go build ./...` pasa sin errores
- `go mod tidy` no agrega dependencias externas nuevas

## Fuente de verdad

- `/home/shared/picokit/pathsec/` (crear)
- `/home/shared/picokit/diff/` (crear)
- `/home/shared/picokit/fuzzy/` (crear)
- `/home/shared/picokit/hashfile/` (crear)
- `/home/shared/roadmapctl/internal/fsx/` (fuente pathsec)
- `/home/shared/roadmapctl/internal/diff/` (fuente diff)
- `/home/shared/rootline/internal/fuzzy/` (fuente fuzzy)
- `/home/shared/backscroll/internal/sync/sync.go` (fuente HashFile)
- `/home/shared/roadmapctl/internal/updater/updater.go` (fuente writeAtomic)
