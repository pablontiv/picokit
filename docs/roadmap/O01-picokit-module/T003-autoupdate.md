---
estado: Specified
tipo: task
---
# T003: Implementar paquete autoupdate (staged async binary updater)

**Outcome**: [O01 Módulo picokit](README.md)
**Contribuye a**: Self-update disponible para los 3 proyectos del ecosistema

[[blocked_by:./T002-zero-dep-packages.md]]

## Preserva

- INV1: Un fallo de red en FetchAndStage nunca interrumpe el comando en curso.
  - Verificar: correr con red desconectada; el proceso termina normalmente
- INV2: Un fallo de permisos al reemplazar el binario retorna nil silenciosamente.
  - Verificar: binario en path sin permisos de escritura; el comando continúa normal
- INV3: Los tests no hacen llamadas de red reales.
  - Verificar: `go test ./autoupdate/... -count=1` pasa con red desconectada

## Contexto

El paquete `autoupdate` es el componente principal de picokit. Toma el updater de roadmapctl y lo parametriza para que cualquier proyecto pueda instanciar su propio updater.

Patrón staged async de dos fases que se distribuyen en dos invocaciones del CLI:
1. Invocación N: `go FetchAndStage(currentVersion)` — goroutine en background, descarga el binario siguiente a staging sin bloquear
2. Invocación N+1: `ApplyStagedIfAvailable()` — sync al inicio, detecta el binario staged, lo aplica con atomic rename y re-exec

Wiring en el CLI consumidor (no implementado en esta task):
```go
updater.ApplyStagedIfAvailable()        // sync: si hay update, re-exec y no continúa
go updater.FetchAndStage(currentVersion) // goroutine: descarga para el próximo run
```

Staging dir: `~/.cache/{Binary}/staged/{tag}/{Binary}`
Skip: `currentVersion == "dev"` o env `{EnvDisable} == "1"`
Integridad: SHA-256 verificado contra `checksums.txt` del release
Apply: `os.Rename(staged, os.Executable())` — atómico en mismo filesystem
Re-exec Unix: `syscall.Exec(newBinary, os.Args, os.Environ())`
Re-exec Windows: `exec.Command(newBinary, os.Args[1:]...) + os.Exit(0)`

Fuentes:
- `/home/shared/roadmapctl/internal/updater/updater.go` — FetchAndStage implementado (T001)
- `/home/shared/roadmapctl/docs/roadmap/O21-auto-update/T002-updater-apply-reexec.md` — spec de ApplyStagedIfAvailable
- `/home/shared/roadmapctl/docs/roadmap/O21-auto-update/T004-updater-tests.md` — spec de tests

## Alcance

**In**:
1. `autoupdate/updater.go` — `type Updater struct { Repo, Binary, EnvDisable string }`, `func New(repo, binary, envDisable string) *Updater`, `func (u *Updater) FetchAndStage(currentVersion string) error`
2. `autoupdate/apply.go` — `func (u *Updater) ApplyStagedIfAvailable() error`
3. `autoupdate/exec_unix.go` (build tag `!windows`) — re-exec via `syscall.Exec`
4. `autoupdate/exec_windows.go` (build tag `windows`) — re-exec via `exec.Command` + `os.Exit`
5. `autoupdate/updater_test.go` — tests de FetchAndStage con `httptest.NewServer`
6. `autoupdate/apply_test.go` — tests de ApplyStagedIfAvailable con exec function inyectable

**Out**:
- No wiring en ningún proyecto consumidor (backscroll, roadmapctl, rootline)
- No descargar nada en `ApplyStagedIfAvailable` — solo detectar staged y re-exec

## Estado inicial esperado

- T002 completada: `hashfile/` existe y exporta `WriteAtomic` (usado internamente por autoupdate)
- `go.mod` existe con module path correcto

## Criterios de Aceptación

- `TestFetchAndStage_SkipsDevVersion`: `FetchAndStage("dev")` retorna nil sin llamadas HTTP
- `TestFetchAndStage_SkipsNoUpdateEnv`: con `EnvDisable` seteado, retorna nil sin HTTP
- `TestFetchAndStage_SkipsIfAlreadyStaged`: binario staged existente → no re-descarga
- `TestFetchAndStage_VerifiesSHA256`: SHA256 incorrecto retorna error, nada escrito a staging
- `TestApply_SkipsIfNothingStaged`: retorna nil si no hay staged
- `TestApply_SkipsIfNotNewer`: versión staged igual o menor → skip, retorna nil
- `go test ./autoupdate/... -race -count=1` pasa en Linux y macOS
- Cobertura `./autoupdate/...` ≥85%

## Fuente de verdad

- `/home/shared/picokit/autoupdate/` (crear)
- `/home/shared/roadmapctl/internal/updater/updater.go` (fuente FetchAndStage)
- `/home/shared/roadmapctl/docs/roadmap/O21-auto-update/T002-updater-apply-reexec.md` (spec apply)
- `/home/shared/roadmapctl/docs/roadmap/O21-auto-update/T004-updater-tests.md` (spec tests)
