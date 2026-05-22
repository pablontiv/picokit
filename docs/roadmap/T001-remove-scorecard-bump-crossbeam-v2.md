---
estado: Specified
tipo: task
---
# T001: Remover workflow Scorecard local y bumpear a `crossbeam@v2`

**Contribuye a**: eliminar startup_failures de Scorecard en picokit (3/3 recientes) y heredar los cambios saneados de crossbeam.

## Preserva

- INV1: CodeQL, gitleaks, golangci-lint y release automation siguen funcionando.

## Contexto

picokit usa `pablontiv/crossbeam@v1`. Workflow Scorecard ha tenido 3/3 startup_failure recientes (infra/timeout). Tras la publicación de `crossbeam@v2`:
1. Eliminar workflow scorecard local.
2. Bumpear referencias `@v1` → `@v2`.

picokit es el repo más chico del ecosistema y se sugiere **validar primero acá** antes de propagar el bump a backscroll/rootline/roadmapctl.

## Alcance

**In**:
1. Eliminar `.github/workflows/scorecard.yml`.
2. Bumpear `pablontiv/crossbeam/...@v1` → `@v2` en `.github/workflows/*.yml`.
3. Actualizar README/CLAUDE.md si listan scorecard.

**Out**:
- No tocar `pkcov` ni la documentación de coverage-spec.

## Estado inicial esperado

- `crossbeam@v2` publicado.
- `.github/workflows/scorecard.yml` existe.

## Criterios de Aceptación

- `ls .github/workflows/scorecard.yml` retorna "No such file or directory".
- `grep -rE 'pablontiv/crossbeam/.*@v1' .github/workflows/` retorna 0 matches.
- Próximo push a main: `gh run list --repo pablontiv/picokit --branch main --limit 5` no muestra `startup_failure`.

## Fuente de verdad

- `/home/shared/picokit/.github/workflows/`
- `/home/shared/picokit/README.md` (si lista workflows)
