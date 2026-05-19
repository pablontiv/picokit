---
estado: Specified
tipo: task
---
# T001: Add community health and GitHub files

**Outcome**: [O02 Public release of picokit v0.1.0](README.md)
**Contribuye a**: El repo cumple los estándares de salud comunitaria y seguridad requeridos antes de publicar como módulo externo.

## Preserva

- INV1: Todos los workflows CI existentes pasan sin modificación.
  - Verificar: `git push` desencadena CI verde (o correr localmente si no hay push).

## Contexto

picokit es una librería Go zero-dependencias. Antes de publicarla como módulo importable, el repo necesita los mismos archivos de comunidad y seguridad que tienen rootline (`/home/shared/rootline`) y roadmapctl (`/home/shared/roadmapctl`).

Archivos de referencia en `/home/shared/roadmapctl/`:
- `LICENSE` — PolyForm Noncommercial License 1.0.0
- `CODE_OF_CONDUCT.md` — Contributor Covenant
- `SECURITY.md` — política de vulnerabilidades (adaptar: quitar referencias a roadmapctl, generalizar para una librería)
- `CONTRIBUTING.md` — guía de contribución (adaptar: dev setup para librería Go sin binario, sin rootline ni Cobra)
- `.github/CODEOWNERS` — `* @pablontiv`
- `.github/dependabot.yml` — gomod + github-actions weekly
- `.github/ISSUE_TEMPLATE/bug_report.md`
- `.github/ISSUE_TEMPLATE/feature_request.md`
- `.github/PULL_REQUEST_TEMPLATE.md`

Archivos a crear desde cero (no tienen equivalente en roadmapctl):
- `README.md` — documentar los 7 paquetes del módulo con una línea por paquete: autoupdate, diag, diff, fuzzy, hashfile, output, pathsec.
- `CHANGELOG.md` — entrada inicial para v0.1.0 con los 7 paquetes listados.

## Alcance

**In**:
1. Crear `LICENSE` (copia exacta de roadmapctl).
2. Crear `CODE_OF_CONDUCT.md` (copia exacta de roadmapctl).
3. Crear `SECURITY.md` adaptado: generalizar el scope (sin mencionar rootline/subprocess), conservar el proceso de reporte y las medidas de seguridad (CodeQL, Gitleaks, Scorecard, SHA-pinned actions, Dependabot).
4. Crear `CONTRIBUTING.md` adaptado: dev setup con `go build ./...` y `go test ./...` y `golangci-lint run ./...`; sin binario, sin rootline, sin Cobra.
5. Crear `.github/CODEOWNERS` con `* @pablontiv`.
6. Crear `.github/dependabot.yml` (copia de roadmapctl).
7. Crear `.github/ISSUE_TEMPLATE/bug_report.md` (adaptar: referencias a roadmapctl → picokit).
8. Crear `.github/ISSUE_TEMPLATE/feature_request.md` (adaptar: referencias a roadmapctl → picokit).
9. Crear `.github/PULL_REQUEST_TEMPLATE.md` (copia/adaptar de roadmapctl).
10. Crear `README.md` con tabla o lista de los 7 paquetes.
11. Crear `CHANGELOG.md` con entrada `## v0.1.0` y lista de paquetes iniciales.

**Out**:
- No modificar workflows CI existentes (`ci.yml`, `codeql.yml`, `scorecard.yml`).
- No modificar código Go ni tests.
- No modificar `go.mod`.

## Estado inicial esperado

- `/home/shared/picokit/` no tiene `LICENSE`, `README.md`, `CHANGELOG.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`.
- `/home/shared/picokit/.github/` solo contiene `workflows/`.

## Criterios de Aceptación

- `ls /home/shared/picokit/LICENSE` existe y contiene PolyForm Noncommercial.
- `ls /home/shared/picokit/README.md` existe y menciona los 7 paquetes.
- `ls /home/shared/picokit/CHANGELOG.md` existe con una entrada v0.1.0.
- `ls /home/shared/picokit/SECURITY.md` existe sin referencias a roadmapctl/rootline.
- `ls /home/shared/picokit/.github/CODEOWNERS` existe con `* @pablontiv`.
- `ls /home/shared/picokit/.github/dependabot.yml` existe.
- `ls /home/shared/picokit/.github/ISSUE_TEMPLATE/` contiene `bug_report.md` y `feature_request.md`.
- `ls /home/shared/picokit/.github/PULL_REQUEST_TEMPLATE.md` existe.
- `go build ./...` desde `/home/shared/picokit` pasa sin errores.

## Fuente de verdad

- `/home/shared/roadmapctl/LICENSE`
- `/home/shared/roadmapctl/CODE_OF_CONDUCT.md`
- `/home/shared/roadmapctl/SECURITY.md`
- `/home/shared/roadmapctl/CONTRIBUTING.md`
- `/home/shared/roadmapctl/.github/CODEOWNERS`
- `/home/shared/roadmapctl/.github/dependabot.yml`
- `/home/shared/roadmapctl/.github/ISSUE_TEMPLATE/`
- `/home/shared/roadmapctl/.github/PULL_REQUEST_TEMPLATE.md`
- `/home/shared/picokit/` (destino de todos los archivos)
