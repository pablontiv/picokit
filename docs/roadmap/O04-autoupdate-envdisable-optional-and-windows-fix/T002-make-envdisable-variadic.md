---
estado: Specified
tipo: task
---
# T002: Hacer `envDisable` opcional en `autoupdate.New` vía variadic

**Outcome**: [O04 autoupdate envDisable opcional + fix build Windows](README.md)
**Contribuye a**: feature principal del outcome — permitir llamadas a `autoupdate.New` sin tercer arg, sin romper call-sites históricos.

[[blocked_by:./T001-fix-windows-build-remove-fmt-import.md]]

## Preserva

- INV1 del outcome: las llamadas existentes en `roadmapctl/internal/cli/cli.go:5431` (con `"ROADMAPCTL_NO_UPDATE"`) y `backscroll/cmd/backscroll/main.go:9991` (con `"BACKSCROLL_AUTOUPDATE_DISABLE"`) deben seguir compilando sin modificación. Verificable: cross-build de esos repos contra el HEAD de picokit tras este cambio.
- INV2 del outcome: coverage del package autoupdate ≥85% tras el cambio.

## Contexto

Hoy `autoupdate.New(repo, binary, envDisable string) *Updater` exige los tres strings. Los dos consumidores actuales eligieron nombres divergentes (`ROADMAPCTL_NO_UPDATE` vs `BACKSCROLL_AUTOUPDATE_DISABLE`). Un tercer consumidor (rootline) quiere wirear autoupdate sin convención de env var. La firma actual obliga a inventar un nombre placeholder o pasar `""` (que funciona por `os.Getenv("") == ""` nunca igualando `"1"`, pero queda como código confuso).

La firma variadic es backwards-compatible: Go permite invocar funciones variadic con args explícitos. Las dos llamadas existentes funcionan sin tocar nada. Nuevos call-sites usan dos args.

## Alcance

**In**:

1. Editar `/home/shared/picokit/autoupdate/updater.go`. Cambiar:

   ```go
   func New(repo, binary, envDisable string) *Updater {
       return &Updater{
           Repo:         repo,
           Binary:       binary,
           EnvDisable:   envDisable,
           ...
       }
   }
   ```

   por:

   ```go
   // New creates a new Updater. envDisable is optional; if omitted or empty,
   // no environment variable can disable the updater (only version=="dev" does).
   func New(repo, binary string, envDisable ...string) *Updater {
       env := ""
       if len(envDisable) > 0 {
           env = envDisable[0]
       }
       return &Updater{
           Repo:         repo,
           Binary:       binary,
           EnvDisable:   env,
           ...
       }
   }
   ```

2. Actualizar el comentario doc del campo `EnvDisable` en el struct `Updater` para reflejar que `""` significa "sin opt-out por env".

3. Revisar `autoupdate/*_test.go`: si algún test construye `New(repo, binary, "")` con la intención de "sin env disable", reescribir como `New(repo, binary)` (sin el tercer arg) para ejercitar el nuevo path. Si un test ya pasa un env disable explícito (e.g., `"TEST_NO_UPDATE"`), dejarlo igual — sigue compilando.

4. Añadir al menos un test nuevo: `TestNew_OmitsEnvDisable_NoOptOut` que verifica que `New(repo, binary)` produce un `Updater` con `EnvDisable == ""` y que `FetchAndStage` con `currentVersion != "dev"` no se cortocircuita por env.

5. Verificar local:
   - `go test ./autoupdate/... -cover` ≥85%.
   - `go build ./...`.
   - `GOOS=windows go build ./...` (combinado con T001).

**Out**:
- No tocar `FetchAndStage`, `ApplyStagedIfAvailable`, ni otros métodos.
- No cambiar el struct `Updater` fuera del comentario del campo `EnvDisable`.
- No tocar consumidores (eso es responsabilidad de las tasks de bump en cada repo).

## Estado inicial esperado

- `autoupdate/updater.go` define `New(repo, binary, envDisable string)`.
- T001 completada (build Windows verde).

## Criterios de Aceptación

- `autoupdate.New` admite `New(repo, binary)` y `New(repo, binary, envDisable)` — ambas formas compilan y funcionan.
- Tests existentes pasan sin modificación funcional.
- Test nuevo cubre la rama sin env disable.
- Coverage del package autoupdate ≥85%.
- `go build ./...` y `GOOS=windows go build ./...` exit 0.

## Fuente de verdad

- `/home/shared/picokit/autoupdate/updater.go` — firma + struct
- `/home/shared/picokit/autoupdate/*_test.go` — tests a revisar/extender
