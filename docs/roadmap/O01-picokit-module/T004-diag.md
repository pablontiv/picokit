---
estado: Completed
tipo: task
---
# T004: Implementar paquete diag (structured errors + report)

**Outcome**: [O01 Módulo picokit](README.md)
**Contribuye a**: Error handling y diagnósticos consistentes en el ecosistema

[[blocked_by:./T001-repo-setup.md]]

## Preserva

- INV1: El tipo `Error` implementa la interfaz `error` de Go.
  - Verificar: `var _ error = (*Error)(nil)` compila sin errores

## Contexto

Dos implementaciones divergentes del mismo concepto existen en el ecosistema:

- backscroll `internal/diagnostics/diagnostics.go` (42 líneas): `Error{Code, Message, Cause}` simple con `New()` y `Wrap()`
- roadmapctl `internal/diagnostics/`: `Report` estructurado, `Diagnostic`, `Severity`, JSON/text render, exit codes con `ExitOK`…`ExitInternal`

El paquete `diag` unifica ambas en dos capas coexistentes dentro del mismo paquete:
- **Capa simple**: para errores de CLI cotidianos — de backscroll, sin cambios
- **Capa report**: para diagnósticos estructurados — de roadmapctl, sin los IDs domain-specific

Los diagnostic IDs específicos de roadmapctl (`RMC_CONFIG_MISSING`, etc.) se quedan en roadmapctl. Solo las constantes de exit code son universales.

Fuentes:
- `/home/shared/backscroll/internal/diagnostics/diagnostics.go`
- `/home/shared/roadmapctl/internal/diagnostics/`

## Alcance

**In**:
1. `diag/error.go` — `type Error struct`, `func New(code, msg string) *Error`, `func Wrap(code, msg string, cause error) *Error`, métodos `Error() string` y `Unwrap() error`
2. `diag/report.go` — `type Severity`, `type Diagnostic`, `type Summary`, `type Report`, `func NewReport(kind, root string, diagnostics []Diagnostic) Report`, `func RenderJSON(w io.Writer, report Report) error`, `func RenderText(w io.Writer, report Report) error`, `func ExitCode(report Report, strict bool) int`
3. `diag/exit.go` — constantes `ExitOK=0`, `ExitValidation=1`, `ExitUsage=2`, `ExitEnvironment=3`, `ExitInternal=4`
4. `diag/diag_test.go` — tests de ambas capas, cobertura ≥85%

**Out**:
- No incluir diagnostic IDs domain-specific de ningún proyecto
- No depender de roadmapctl, backscroll ni ningún otro proyecto del ecosistema

## Estado inicial esperado

- T001 completada: `go.mod` existe con module path correcto

## Criterios de Aceptación

- `var _ error = (*Error)(nil)` compila
- `New("ERR_FOO", "algo falló").Error()` retorna string con code y message
- `Wrap("ERR_FOO", "algo falló", err).Unwrap()` retorna el error original
- `RenderJSON(w, report)` produce JSON válido con `json.Valid()`
- `RenderText(w, report)` produce output con severidades legibles
- `ExitCode(report, false)` retorna 0 para report sin errores
- `ExitCode(report, true)` retorna non-zero si hay warnings con strict=true
- `go test ./diag/... -race` pasa, cobertura ≥85%

## Fuente de verdad

- `/home/shared/picokit/diag/` (crear)
- `/home/shared/backscroll/internal/diagnostics/diagnostics.go` (fuente capa simple)
- `/home/shared/roadmapctl/internal/diagnostics/` (fuente capa report)
