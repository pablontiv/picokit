---
estado: Specified
tipo: task
---
# T001: Inicializar repo picokit con CI/CD y tooling

**Outcome**: [O01 Módulo picokit](README.md)
**Contribuye a**: Base de infraestructura sobre la que se crean todos los paquetes

## Preserva

- INV1: El módulo compila sin errores desde un repo limpio.
  - Verificar: `cd /home/shared/picokit && go build ./...`

## Contexto

Picokit es un nuevo módulo Go en `/home/shared/picokit` con module path `github.com/pablontiv/picokit`. Necesita la misma infraestructura que backscroll y roadmapctl: CI delegado a crossbeam, goreleaser para builds multi-plataforma, y Justfile para comandos locales.

Referencias modelo:
- `/home/shared/backscroll/.github/workflows/ci.yml` — patrón de workflows con crossbeam
- `/home/shared/backscroll/.goreleaser.yml` — config goreleaser
- `/home/shared/backscroll/Justfile` — comandos locales
- `/home/shared/backscroll/install.sh` — script de instalación

Nota: picokit es una librería, no un binario de usuario. El `.goreleaser.yml` solo necesita generar releases con checksums y release notes, no binarios.

## Alcance

**In**:
1. `go.mod` con `module github.com/pablontiv/picokit` y versión go 1.26
2. `.github/workflows/ci.yml` delegando a `pablontiv/crossbeam@v1`: go-ci.yml + gitleaks.yml + go-release.yml
3. `.github/workflows/codeql.yml` delegando a `pablontiv/crossbeam@v1`: codeql.yml
4. `.github/workflows/scorecard.yml` delegando a `pablontiv/crossbeam@v1`: scorecard.yml
5. `.goreleaser.yml` configurado para librería (sin builds de binarios, con changelog)
6. `Justfile` con targets: check (gofmt+vet), test, fmt, coverage-summary, audit
7. `CLAUDE.md` con overview del módulo y convenciones

**Out**:
- No implementar ningún paquete Go (eso es T002–T005)
- No configurar GitHub Actions secrets (se configuran en GitHub)
- No agregar install.sh (picokit es librería, no binario instalable)

## Estado inicial esperado

- `/home/shared/picokit` existe como git repo (inicializado)
- `roadmapctl bootstrap` pasa sin errores en este repo

## Criterios de Aceptación

- `go build ./...` pasa desde `/home/shared/picokit`
- `.github/workflows/ci.yml` usa crossbeam igual que backscroll
- `goreleaser check --config .goreleaser.yml` pasa
- `Justfile` tiene targets check/test/fmt/coverage-summary/audit
- `just check` pasa en repo con solo la infraestructura

## Fuente de verdad

- `/home/shared/picokit/go.mod` (crear)
- `/home/shared/picokit/.github/workflows/` (crear)
- `/home/shared/picokit/.goreleaser.yml` (crear)
- `/home/shared/picokit/Justfile` (crear)
- `/home/shared/picokit/CLAUDE.md` (crear)
- `/home/shared/backscroll/.github/workflows/ci.yml` (modelo)
- `/home/shared/backscroll/.goreleaser.yml` (modelo)
- `/home/shared/backscroll/Justfile` (modelo)
