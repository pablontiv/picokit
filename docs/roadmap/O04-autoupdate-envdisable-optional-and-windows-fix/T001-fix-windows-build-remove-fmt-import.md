---
estado: Specified
tipo: task
---
# T001: Fix build Windows — quitar import muerto `"fmt"` de `autoupdate/exec_windows.go`

**Outcome**: [O04 autoupdate envDisable opcional + fix build Windows](README.md)
**Contribuye a**: cumplir INV1 del outcome restaurando el build de Windows en main, precondición para publicar v0.4.0.

## Contexto

`/home/shared/picokit/autoupdate/exec_windows.go:6` declara `import "fmt"` pero ninguna función del archivo (`atomicReplace`, `platformExec`) llama a `fmt.*`. Bajo el build tag `//go:build windows`, Go falla con `imported and not used: "fmt"`. El error pasa desapercibido en desarrollo local sobre Linux porque el archivo solo entra al build cuando `GOOS=windows`. Sin embargo cualquier cross-compile a Windows lo dispara, y el pipeline de release (goreleaser) lo dispararía al construir el binario Windows.

## Alcance

**In**:

1. Editar `/home/shared/picokit/autoupdate/exec_windows.go`: borrar la línea `"fmt"` del bloque `import (...)`. Quedan `"os"` y `"os/exec"`, ambos usados.
2. Verificar con `GOOS=windows go build ./autoupdate/...` desde la raíz del módulo. Debe compilar sin error.
3. Verificar que el build Linux sigue verde: `go build ./autoupdate/...`.
4. Verificar que los tests pasan: `go test ./autoupdate/...`.

**Out**:
- No tocar la lógica de `atomicReplace` ni `platformExec`.
- No tocar `autoupdate/exec_unix.go` ni otros archivos del package.
- No releaser todavía — eso lo hace T003.

## Estado inicial esperado

- `autoupdate/exec_windows.go` tiene `"fmt"` en imports sin referencias en el cuerpo.
- `GOOS=windows go build ./autoupdate/...` falla con `imported and not used: "fmt"`.

## Criterios de Aceptación

- `autoupdate/exec_windows.go` no importa `"fmt"`.
- `GOOS=windows go build ./autoupdate/...` exit 0.
- `go build ./autoupdate/...` exit 0.
- `go test ./autoupdate/...` exit 0.
- `pkcov check` sigue verde (no se altera coverage del package).

## Fuente de verdad

- `/home/shared/picokit/autoupdate/exec_windows.go` — único archivo a modificar
