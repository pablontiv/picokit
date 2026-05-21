---
estado: Completed
tipo: task
---
# T001: Escribir docs/coverage-spec.md v1.0

**Outcome**: [O03 Coverage tooling compartido](README.md)
**Contribuye a**: capturar el contrato autoritativo que los consumidores referencian (INV2)

## Preserva

- INV2 del outcome: la spec describe el qué; el código describe el cómo
  - Verificar: el doc captura el comportamiento del script bash de rootline sin duplicar el código Go

## Contexto

`/home/shared/rootline/scripts/check-coverage-floors.sh` (115 líneas) implementa el contrato actualmente: per-package floor uniforme desde `.coverage-floors.toml`, skip de paquetes test-only, exit 1 con detalle. Funciona en producción (TOTAL 88.9%, 13/13 PASS). La task formaliza ese contrato como documento Markdown corto (~150 líneas) en picokit.

El doc no es una traducción del bash; es el contrato que el bash y el futuro pkcov ambos cumplen. Sirve para:
- Discutir cambios de política sin leer código (e.g. "¿agregamos per-package override?" → cambio al spec, no hack local)
- Los consumidores declaran "cumple v1.0" en su CLAUDE.md/README
- Onboarding sin tener que leer Go

## Alcance

**In**: crear `/home/shared/picokit/docs/coverage-spec.md` con encabezado `version: v1.0` y las siguientes 8 secciones:

1. **Threshold uniforme** — floor de 85% aplicado al total y a cada paquete sin descuento. Configurable por repo via TOML (default 85). Cambiar el número requiere decisión consciente en TOML.
2. **Política monotónica** — el threshold sólo sube. Bajarlo requiere justificación explícita en commit. La revisión del PR debe rechazar bajadas sin justificación.
3. **Schema de `.coverage-floors.toml`** — formato TOML con `default = N` + `packages = [...]`. No per-package override en v1.
4. **Visibilidad local** — el repo expone recipes `coverage` (reporte) y `coverage-check` (gate) en su Justfile/Makefile, invocando `pkcov`.
5. **Gate pre-push** — el repo ejecuta `pkcov check` desde `.githooks/pre-push` cuando cambian archivos `*.go`. `--no-verify` prohibido salvo emergencia documentada en commit.
6. **Política de dead code** — stubs (`func DetectX(...) { return nil }`), comandos `// Deprecated:` no usados, render helpers no invocados → se borran. Borrar dead code es la única forma legítima de subir coverage sin agregar tests.
7. **Paquetes test-only** — paquetes sin source statements (sólo `_test.go`) se reportan como `SKIP`, no como falla.
8. **Versioning del spec** — el doc declara `version: vMAJOR.MINOR`. Cambios breaking suben major. Consumidores referencian la versión que cumplen.

**Out**:
- No implementar código (eso es T002+).
- No documentar la API de la librería en este doc (vive como godoc en `coverage/`).
- No incluir ejemplos de migración por repo (eso vive en los outcomes de adopción).

## Estado inicial esperado

- `/home/shared/picokit/docs/` existe (verificado).
- No existe `coverage-spec.md` aún.
- El script de rootline (`/home/shared/rootline/scripts/check-coverage-floors.sh`) está disponible como referencia funcional.

## Criterios de Aceptación

- `/home/shared/picokit/docs/coverage-spec.md` existe.
- Frontmatter o encabezado declara `version: v1.0`.
- Las 8 secciones nombradas en "Alcance" están presentes.
- Cada sección es prosa accionable (no placeholders).
- Total ~150 líneas (rango aceptable 100–200).
- El doc no incluye ejemplos de código Go (esos van en godoc).

## Fuente de verdad

- `/home/shared/rootline/scripts/check-coverage-floors.sh` — comportamiento de referencia
- `/home/shared/rootline/.coverage-floors.toml` — schema de referencia
- `/home/shared/picokit/docs/` — directorio destino
