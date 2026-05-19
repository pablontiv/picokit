---
estado: Specified
tipo: task
---
# T002: Remove RoadmapRoot domain logic from picokit/diag

**Outcome**: [O02 Public release of picokit v0.1.0](README.md)
**Contribuye a**: picokit/diag es una librería genérica sin acoplamiento a lógica de dominio de roadmapctl.

## Preserva

- INV1: La API pública de picokit/diag no incluye conceptos específicos de roadmapctl.
  - Verificar: `grep -r "RoadmapRoot\|roadmap_root" /home/shared/picokit/diag/` retorna vacío.
- INV2: Coverage ≥ 85% en el paquete diag.
  - Verificar: `go test -cover ./diag/...` desde `/home/shared/picokit`.

## Contexto

Durante el bootstrapping inicial de picokit, el paquete `diag` fue extraído de roadmapctl con demasiado contexto de dominio. En particular:

- `diag.Report` tiene un campo `RoadmapRoot string \`json:"roadmap_root"\`` que es específico de roadmapctl.
- `diag.NewReport` fija `RoadmapRoot: ""` en lugar de no tenerlo.

Este campo no pertenece a una librería genérica. roadmapctl es quien sabe de `roadmap_root`; picokit solo debe proveer primitivas.

Archivos afectados en picokit:
- `/home/shared/picokit/diag/report.go` — contiene `Report`, `NewReport`, `RenderJSON`, `RenderText`, `ExitCode`.

## Alcance

**In**:
1. Eliminar el campo `RoadmapRoot string` de la struct `Report` en `diag/report.go`.
2. Ajustar `NewReport` para no setear ese campo (ya no existe).
3. Actualizar tests en `diag/` que referencien `RoadmapRoot` o que usen el campo en assertions.

**Out**:
- No modificar los otros paquetes de picokit (fuzzy, diff, hashfile, output, pathsec, autoupdate).
- No cambiar la firma de `NewReport` más allá de eliminar la referencia interna al campo eliminado.
- No añadir el campo `RoadmapRoot` en ningún otro lugar de picokit.

## Estado inicial esperado

- `diag/report.go` contiene `RoadmapRoot string \`json:"roadmap_root"\`` en la struct `Report`.
- `NewReport` asigna `RoadmapRoot: ""`.

## Criterios de Aceptación

- `grep -r "RoadmapRoot" /home/shared/picokit/diag/` retorna vacío.
- `go test ./diag/...` pasa desde `/home/shared/picokit`.
- `go test -cover ./diag/...` muestra coverage ≥ 85%.
- `go vet ./diag/...` pasa sin warnings.

## Fuente de verdad

- `/home/shared/picokit/diag/report.go`
- `/home/shared/picokit/diag/report_test.go` (si existe)
