---
version: v1.1
---

# Coverage Policy Spec — v1.1

This document is the authoritative contract for coverage enforcement in `pablontiv` repos.
It describes **what** the policy enforces; the code (`pkcov`) describes **how**.
Consumers declare compliance by referencing the version they implement.

---

## 1. Threshold uniforme

A single floor applies uniformly: **85% statement coverage** on the total profile
and on every package in the module discovered in the coverage profile. No per-package
discount or override exists — every package is held to the same bar.

The gate applies automatically to all packages the coverage profile reports. A package
is only excluded from the gate when explicitly listed in `exclude` (see §3); packages
absent from `exclude` are always checked.

The threshold is configured per-repo in `.coverage-floors.toml` (see §3).
Changing the numeric value requires an explicit, conscious decision recorded in that file;
the default is 85 and should only move upward (see §2).

A run fails if **any** package in the profile falls below the threshold, even if the
aggregate total passes.

---

## 2. Política monotónica

Coverage thresholds only go up. Lowering the `default` value in `.coverage-floors.toml`
requires explicit justification in the commit message explaining why the regression is
acceptable (e.g., removing an unmaintainable test suite, retiring a package).

PR reviewers must reject threshold reductions that lack justification. Silence is not
consent — an unexplained reduction is a defect, not a style choice.

Raising the threshold is always welcome and requires no justification.

---

## 3. Schema de `.coverage-floors.toml`

The configuration file lives at the repo root as `.coverage-floors.toml`.

```toml
default = 85

# exclude = ["cmd/experimental"]  # optional; each entry documents deliberate debt
```

Fields:

- `default` (integer, required): statement coverage floor in percent, applied to every
  package discovered in the coverage profile and to the aggregate total.
- `packages` (array of strings, **deprecated in v1.1**): accepted by v1.1 implementations
  for backward compatibility with v1.0 configs, but its content is ignored — the gate
  applies to all discovered packages regardless. No error is emitted if this field is
  present. Consumers should delete it when migrating (see §9).
- `exclude` (array of strings, optional): relative import path suffixes of packages to
  exempt from the gate. Each exclusion represents deliberate debt; the reason must be
  recorded in the commit message or as a comment in the TOML file.

There is no per-package threshold override. Every non-excluded package shares the single
`default` floor. Per-package granularity is deferred to a future major version.

---

## 4. Visibilidad local

Every repo subject to this policy exposes two recipes in its `Justfile` (or `Makefile`):

- `coverage` — generates the coverage report and prints per-package results.
  Used during development to inspect current state without failing the build.
- `coverage-check` — runs the coverage gate. Exits non-zero if any package is below
  the threshold. This is the recipe invoked by CI and the pre-push hook.

Both recipes delegate to `pkcov`, the shared CLI that implements this spec.
Repos must not inline their own coverage logic; they invoke `pkcov` so that
policy changes propagate from a single source.

Example (Justfile):
```
coverage:
    pkcov report

coverage-check:
    pkcov check
```

---

## 5. Gate pre-push

Repos must install a pre-push hook that runs `pkcov check` whenever `*.go` files are
included in the pushed commits. The hook lives at `.githooks/pre-push` and is wired
via `git config core.hooksPath .githooks` (typically in a `bootstrap` or `setup` recipe).

The hook must:
1. Detect whether any `*.go` file changed relative to the remote ref.
2. If yes, run `pkcov check`; block the push on non-zero exit.
3. If no Go files changed, skip the check and exit 0.

`git push --no-verify` is prohibited in normal workflow. Bypassing the hook is allowed
only in documented emergencies — the bypass must be recorded in the commit message or
a follow-up commit explaining the exception. Routine use of `--no-verify` to silence
coverage failures is a policy violation.

---

## 6. Política de dead code

Dead code inflates the denominator and lowers coverage without providing test value.
The only legitimate way to improve coverage without writing tests is to delete dead code.

Code is considered dead when it meets one of these conditions:

- A stub that returns a zero value and is never called from production paths.
- A function or method marked `// Deprecated:` with no remaining callers.
- A render helper, formatter, or adapter invoked only from other dead code.

Dead code must be deleted, not commented out. Commented-out code is also dead code and
must be removed. The exception is code that is intentionally unreachable as a guard
(e.g., a `default` panic in an exhaustive switch) — that code documents intent and is
not subject to this rule.

Tooling such as `deadcode` (golang.org/x/tools/cmd/deadcode) can assist in identifying
candidates, but human review is required before deletion.

---

## 7. Paquetes test-only

A package with no source statements — one that contains only `_test.go` files and
declares no non-test symbols — is reported as `SKIP`, not `FAIL`.

The checker detects test-only packages automatically by examining the coverage profile:
a package with zero statements in the profile is skipped without any configuration.
No annotation in `.coverage-floors.toml` is required or expected for this case.

This automatic detection is the primary mechanism for test-only packages. The `exclude`
field (§3) is orthogonal: it is reserved for packages that have coverable statements
but should be exempted from the gate for deliberate reasons — for example, experimental
CLI scaffolding or generated code that cannot meaningfully be tested. Using `exclude` on
a test-only package is redundant but harmless.

This rule exists because test-only packages (e.g., end-to-end test suites, fixture
packages) have no coverable source statements by definition. Penalizing them for 0%
coverage would be meaningless.

---

## 8. Versioning del spec

This document declares its version in YAML frontmatter (`version: v1.1`).

Versioning follows `vMAJOR.MINOR`:

- **MINOR** bumps: additive, non-breaking changes (new clarifications, new optional
  fields, extended behavior that does not invalidate existing compliant implementations).
- **MAJOR** bumps: changes that require existing compliant implementations to update
  their behavior or configuration (e.g., introducing per-package overrides, changing
  the threshold semantics, removing a section).

`v1.0 → v1.1` is a **MINOR** bump. Implementations conforming to v1.0 remain compliant:
the gate semantics are strictly additive (auto-discovery replaces manual listing, `packages`
is deprecated but still accepted, `exclude` is a new opt-out mechanism). No existing
`.coverage-floors.toml` breaks on a v1.1 implementation.

Consumers declare compliance by referencing the version they implement, typically in
their `CLAUDE.md` or `README.md`:

```
Coverage policy: complies with github.com/pablontiv/picokit coverage-spec v1.1
```

A consumer referencing `v1.0` is not required to comply with `v2.0` changes until
they explicitly upgrade their declaration.

---

## 9. Migration v1.0 → v1.1

Steps to migrate a repo from coverage-spec v1.0 to v1.1:

1. **Delete the `packages` array** from `.coverage-floors.toml`. The gate now applies
   automatically to all packages discovered in the coverage profile — no list is needed.
2. **Add `exclude` entries** for any package that should be exempted from the gate,
   with a justification comment:
   ```toml
   default = 85
   exclude = [
     "cmd/experimental",  # scaffolding, not yet tested
   ]
   ```
3. **Update the compliance declaration** in `CLAUDE.md` or `README.md` from `v1.0` to `v1.1`.
4. Run `pkcov check` locally to confirm the gate applies to all expected packages.
