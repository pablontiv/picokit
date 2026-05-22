---
estado: Specified
tipo: task
---
# T003: Release v0.4.0 — CHANGELOG + tag

**Outcome**: [O04 autoupdate envDisable opcional + fix build Windows](README.md)
**Contribuye a**: publicar los dos cambios (T001 + T002) como release consumible vía `go get github.com/pablontiv/picokit@v0.4.0`.

[[blocked_by:./T001-fix-windows-build-remove-fmt-import.md]]
[[blocked_by:./T002-make-envdisable-variadic.md]]

## Contexto

T001 y T002 cambian comportamiento del package `autoupdate/`. Para que roadmapctl, backscroll y rootline puedan bumpear, hace falta una versión etiquetada. Como T002 es feature non-breaking y T001 es bugfix, justifica un minor bump: v0.3.0 → v0.4.0 (semver).

El repo usa crossbeam para release (`go-release.yml@v1` en CI), por lo que el push del tag suele bastar para que el binario y los artifacts se publiquen automáticamente.

## Alcance

**In**:

1. Editar `/home/shared/picokit/CHANGELOG.md`: agregar sección `## [v0.4.0]` arriba de la última versión existente. Contenido:

   ```markdown
   ## [v0.4.0] - YYYY-MM-DD

   ### Fixed

   - `autoupdate`: build de Windows fallaba con `imported and not used: "fmt"` por import muerto en `exec_windows.go`. Corregido.

   ### Changed

   - `autoupdate.New` ahora acepta `envDisable` como argumento variadic opcional. Call-sites previos con tres argumentos siguen compilando sin cambios. Llamadas nuevas pueden omitir el tercer arg para deshabilitar el opt-out por env.
   ```

   Reemplazar `YYYY-MM-DD` con la fecha del release. Seguir el formato Keep a Changelog que ya usa el archivo.

2. Verificar local antes del tag:
   - `just check && just test && just coverage-check` exit 0.
   - `GOOS=windows go build ./...` exit 0.

3. Commit y push del CHANGELOG con mensaje convencional: `chore(release): v0.4.0`.

4. Crear tag local y push:
   ```bash
   git tag v0.4.0
   git push origin v0.4.0
   ```

5. Verificar que CI corre release (`crossbeam/.github/workflows/go-release.yml@v1`) y que aparece la release en GitHub con artifacts para Linux, macOS, Windows.

**Out**:
- No tocar código de `autoupdate/` ni otros packages — eso se hizo en T001/T002.
- No bumpear consumidores (otras tasks en otros repos).

## Estado inicial esperado

- T001 y T002 mergeadas en `master`.
- HEAD del repo compila en Linux y Windows, tests verdes, coverage ≥85%.
- `git tag --sort=-v:refname | head -1` muestra `v0.3.0`.

## Criterios de Aceptación

- `CHANGELOG.md` contiene sección `v0.4.0` con líneas Fixed + Changed.
- Tag `v0.4.0` publicado en GitHub.
- Release de GitHub creado por CI con artifacts.
- `go get github.com/pablontiv/picokit@v0.4.0` desde otro módulo resuelve sin error (verificable desde uno de los consumidores).

## Fuente de verdad

- `/home/shared/picokit/CHANGELOG.md` — archivo a editar
- `/home/shared/picokit/.github/workflows/ci.yml` — referencia del pipeline
- `https://github.com/pablontiv/picokit/releases` — verificar release publicado
