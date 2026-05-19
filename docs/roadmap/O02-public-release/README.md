---
tipo: outcome
---
# O02: Public release of picokit v0.1.0

picokit se publica como módulo Go con tag semántico para que otros repos del ecosistema pablontiv puedan importarlo como dependencia explícita versionada.

Antes de publicar, el repo debe cumplir los mismos estándares de salud comunitaria y seguridad que rootline y roadmapctl: archivos de comunidad, política de seguridad, dependabot, CODEOWNERS y templates de issues/PR. Además, el paquete `diag` debe limpiarse de lógica de dominio específica de roadmapctl (campo `RoadmapRoot` en `Report`) que fue incluido por error durante el bootstrapping inicial del módulo.
