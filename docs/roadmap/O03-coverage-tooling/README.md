---
tipo: outcome
---
# O03: Coverage tooling compartido

Hoy tres repos del usuario (rootline, backscroll, roadmapctl) tienen implementaciones divergentes del mismo gate de cobertura: rootline tiene un script bash de ~115 líneas con per-package floors validado en producción (TOTAL 88.9%, 13/13 paquetes PASS); backscroll tiene un script de 17 líneas solo-total con threshold por env; roadmapctl tiene un script de 30+ líneas que mezcla bash y python3 y lee threshold del `.roadmapctl.toml`. Cada uno reinventó la rueda con menor o mayor fidelidad al estándar.

Este outcome consolida el comportamiento en picokit como librería Go + CLI `pkcov`, tomando rootline como referencia funcional (su script ya validó el contrato en producción). La única generalización funcional es leer el module prefix de `go.mod` en lugar de hardcodearlo. Se acompaña con `docs/coverage-spec.md` v1.0 como contrato autoritativo que los consumidores referencian.

Resultado observable cuando todas las tasks estén completadas: existe `picokit/coverage/` (librería) + `picokit/cmd/pkcov/` (CLI) + `picokit/docs/coverage-spec.md` (spec). Picokit mismo dogfooded el gate. Tag publicado consumible vía `go get github.com/pablontiv/picokit@<tag>`. Los repos consumidores pueden adoptar pkcov reemplazando su bash local.

Invariantes preservadas por las tasks:
- INV1: el comportamiento de pkcov es funcionalmente equivalente al script bash de rootline sobre los mismos inputs (validado con fixtures + diff)
- INV2: la spec v1.0 captura el contrato sin duplicar la implementación; describe el qué, el código describe el cómo
- INV3: picokit cumple su propia spec (dogfooding)

Scope: este outcome crea el tooling. Las adopciones en rootline/backscroll/roadmapctl son outcomes/tasks separadas en sus respectivos repos, cada una con su propio pre-work si aplica.
