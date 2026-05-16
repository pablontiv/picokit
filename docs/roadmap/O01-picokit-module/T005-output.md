---
estado: Specified
tipo: task
---
# T005: Implementar paquete output (multi-format formatter)

**Outcome**: [O01 Módulo picokit](README.md)
**Contribuye a**: Formateo de output consistente en el ecosistema (text/json/robot)

[[blocked_by:./T001-repo-setup.md]]

## Preserva

- INV1: `WriteJSON` produce JSON válido y parseable para cualquier input serializable.
  - Verificar: `json.Valid(buf.Bytes())` para la salida de `WriteJSON`

## Contexto

Backscroll tiene un `Formatter` rico en `internal/output/output.go` (150 líneas) con tres modos: text (human-readable), JSON (indentado), y robot (key=value para LLM injection con token limiting). Los otros proyectos del ecosistema tienen soluciones ad-hoc (tabwriter en rootline, RenderJSON/Text en roadmapctl).

El paquete `output` de picokit generaliza el formatter de backscroll extrayendo los primitivos reusables. El cambio clave: reemplazar `[]models.SearchResult` (tipo domain-specific de backscroll) por primitivos genéricos.

El rendering específico de SearchResult (bordes ASCII, score, tags de sesiones) queda en backscroll; picokit expone solo los building blocks.

Fuente: `/home/shared/backscroll/internal/output/output.go`

## Alcance

**In**:
1. `output/output.go` — `type Format int` con constantes `FormatText`, `FormatJSON`, `FormatRobot`
2. `output/output.go` — `type Formatter struct { Format Format; MaxTokens int }`, `func NewFormatter(format Format, maxTokens int) *Formatter`
3. `output/output.go` — `func (f *Formatter) WriteJSON(w io.Writer, v any) error` — JSON encoder indentado
4. `output/output.go` — `func (f *Formatter) WriteLines(w io.Writer, lines []string) error` — escribe líneas según Format activo
5. `output/output.go` — `func TokenCount(text string) int` — estimación de tokens (palabras × 1.3)
6. `output/output_test.go` — tests de todas las funciones, cobertura ≥85%

**Out**:
- No implementar rendering específico de SearchResult, sessions o cualquier tipo de backscroll
- No depender de backscroll/models ni de ningún proyecto del ecosistema

## Estado inicial esperado

- T001 completada: `go.mod` existe con module path correcto

## Criterios de Aceptación

- `WriteJSON(w, struct{Name string}{"foo"})` produce `{"Name":"foo"}` indentado
- `json.Valid(output)` para cualquier v serializable
- `WriteLines` en FormatJSON produce array JSON de strings
- `WriteLines` en FormatRobot produce `result_0=...`, `result_1=...`
- `TokenCount("hello world")` retorna 2 (±1 por estimación)
- `go test ./output/... -race` pasa, cobertura ≥85%

## Fuente de verdad

- `/home/shared/picokit/output/` (crear)
- `/home/shared/backscroll/internal/output/output.go` (fuente)
