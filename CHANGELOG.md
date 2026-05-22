# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.4.0] - 2026-05-21

### Fixed

- `autoupdate`: build de Windows fallaba con `imported and not used: "fmt"` por import muerto en `exec_windows.go`. Corregido.

### Changed

- `autoupdate.New` ahora acepta `envDisable` como argumento variadic opcional. Call-sites previos con tres argumentos siguen compilando sin cambios. Llamadas nuevas pueden omitir el tercer arg para deshabilitar el opt-out por env.

## [v0.1.0] - 2026-05-19

### Added

- `autoupdate` — Parameterized staged async binary updater
- `diag` — Diagnostic utilities for system and environment inspection
- `diff` — Utilities for computing and comparing differences
- `fuzzy` — Fuzzy matching and searching algorithms
- `hashfile` — Utilities for computing and verifying file hashes
- `output` — Output formatting and presentation utilities
- `pathsec` — Path security and validation utilities

Initial public release of picokit v0.1.0 as a zero-dependency Go utility library.
