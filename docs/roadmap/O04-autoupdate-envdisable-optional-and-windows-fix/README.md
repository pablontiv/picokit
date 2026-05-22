---
tipo: outcome
---
# O04: autoupdate — envDisable opcional + fix build Windows

Dos cambios coexistentes en el package `autoupdate/`. Ambos requieren un único release por afectar al mismo módulo.

**Cambio 1 — Fix build Windows roto en main**. `autoupdate/exec_windows.go:6` importa `"fmt"` pero el archivo no usa el paquete en ninguna función. Bajo el build tag `//go:build windows`, Go falla la compilación con `imported and not used: "fmt"`. Cualquier consumidor que intente cross-compile a Windows (e.g., goreleaser en CI) ve el error. El fix es trivial: borrar la línea del import. Verificable con `GOOS=windows go build ./autoupdate/...` sin requerir host Windows.

**Cambio 2 — Hacer `envDisable` opcional en `autoupdate.New`**. Hoy la firma es `func New(repo, binary, envDisable string) *Updater` — exige los tres strings. Los dos consumidores actuales pasan nombres distintos (`ROADMAPCTL_NO_UPDATE`, `BACKSCROLL_AUTOUPDATE_DISABLE`) sin convención común. Un tercer consumidor (rootline) quiere wirear autoupdate sin definir env var de opt-out. La firma actual obliga a inventar un nombre ficticio o pasar `""` con un workaround documentado. La solución limpia es variadic: `envDisable ...string`. Backwards-compatible — los call-sites existentes compilan sin tocar; nuevos call-sites pueden omitir el tercer arg.

**Resultado observable cuando todas las tasks estén completadas**: tag `v0.4.0` publicado; `GOOS=windows go build ./autoupdate/...` verde; `autoupdate.New("foo/bar", "bar")` válido (dos args); roadmapctl y backscroll siguen compilando con sus llamadas actuales sin modificación.

Invariantes:
- INV1: la firma variadic no rompe llamadas existentes con tercer arg explícito (verificable: `go build` en roadmapctl y backscroll contra v0.4.0 sin tocar su código).
- INV2: el coverage gate del package se mantiene ≥85% tras los cambios.

Scope: este outcome publica v0.4.0. La adopción en consumidores (bump go.mod) son tasks separadas en sus respectivos repos.
