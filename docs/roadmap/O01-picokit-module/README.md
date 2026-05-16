---
tipo: outcome
---
# O01: Módulo picokit — librería compartida del ecosistema pablontiv

Cuando todas las tasks estén completadas, existirá `github.com/pablontiv/picokit` como módulo Go con 7 paquetes genéricos disponibles para roadmapctl, backscroll y rootline: autoupdate, diag, output, pathsec, diff, hashfile y fuzzy.

El ecosistema pablontiv tiene tres CLI tools que comparten lógica genérica actualmente dispersa. En lugar de duplicar utilidades proyecto a proyecto, picokit centraliza el código común — fixes y mejoras se propagan automáticamente a todos los consumidores.

Criterio de inclusión: suficientemente genérico para que cualquiera de los tres proyectos lo use, aunque hoy solo lo use uno. Lógica domain-specific (storage de backscroll, schema de rootline, roadmap model de roadmapctl) se queda donde está.
