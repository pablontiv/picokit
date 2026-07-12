# Follow-up fixes: parser rejection tests, pathsec macOS test, gentle-ai store repair — Design

Date: 2026-07-12
Status: Approved (autonomous session, user-delegated)

## Problems

1. **REL-1 (from 4R review):** `coverage/toml.go` rejects unsupported TOML forms
   (multi-line arrays, single-quoted strings, unicode escapes, multi-line strings)
   with loud errors, but no test locks that behavior. A future parser change could
   silently start accepting or misparsing them.
2. **pathsec macOS failure:** `TestEvalExistingPrefixRealPath` compares
   `evalExistingPrefix(realDir)` against the raw `t.TempDir()` path. On macOS,
   TMPDIR lives under the `/var → /private/var` symlink, so the function correctly
   returns the resolved path and the assertion fails. The test is wrong, not the code:
   symlink resolution is the function's documented purpose (see
   `TestEvalExistingPrefixWithSymlink`).
3. **gentle-ai review store wedged:** `.git/gentle-ai/review-transactions/v2/`
   contains a stale `LOCK` (owner PID 74679 dead, acquired 2026-07-12T18:29Z) and two
   terminal lineages (`review-f9ef23bf20861290`, `review-6bc0797a22af4ce1`). Ambiguous
   discovery makes `validate` return `scope-changed` with empty context and
   `start` fail with "transaction changed concurrently". Tool: gentle-ai 2.0.2
   (github.com/Gentleman-Programming/gentle-ai).

## Design

### 1. Parser rejection tests (`coverage/toml_test.go`)

Extend the existing `TestParseFloorsTOML` table with error cases (each expects an
error containing the offending line number):

- multi-line array opener: `exclude = [\n  "a",\n]` → error at line 1 (Atoi/array error)
- single-quoted string: `exclude = ['pkg/a']` → "expected quoted string"
- unicode escape: `exclude = ["A"]` → "unsupported escape sequence"
- multi-line basic string: `name = """x"""` on a known key is N/A (no string keys);
  instead: `exclude = ["""a"""]` → error (empty string parsed, then `"a"""` breaks
  array grammar — assert error, exact message unpinned)

Tests only; parser behavior unchanged. If any case unexpectedly passes parsing,
that is a parser bug to fix within the loud-subset contract (error, never silent).

### 2. pathsec test canonicalization (`pathsec/pathsec_test.go`)

In `TestEvalExistingPrefixRealPath`, canonicalize the expectation:

```go
want, err := filepath.EvalSymlinks(realDir)
if err != nil { t.Fatal(err) }
```

and compare `result != want`. Portable across macOS/Linux; keeps full assertion
strength (no skip).

### 3. gentle-ai store repair + upstream bug report

Repair (maintainer action, user-authorized):
- Verify LOCK owner PID is dead; delete `LOCK`.
- Move both lineage directories to `.git/gentle-ai/review-transactions/archive-2026-07-12/`
  (preserved for forensics, out of discovery scope).
- Validation: run a full clean cycle over the follow-up commits
  (`review start` → lens work if selected → `finalize --evidence` → `validate --gate pre-commit`)
  and confirm the gate allows.

Upstream report (via issue-creation skill) on Gentleman-Programming/gentle-ai:
- Bug 1: with two terminal lineages, `review validate` returns
  `scope-changed`/`create-new-lineage` with an empty context block instead of the
  "multiple lineages; specify --lineage" error it returns for pre-commit — and the
  suggested action (`create-new-lineage`) is impossible because `review start` fails
  with "transaction changed concurrently: expected compact revision ''".
- Bug 2: `review start` does not detect/steal stale locks from dead PIDs.
- Include reproduction narrative and version (2.0.2, macOS).

## Testing

Items 1–2: `go test ./coverage/ ./pathsec/` green on macOS (the previously failing
pathsec test now passes locally) and coverage floors hold (≥85%). Item 3: gate cycle
allows pre-commit on the new work.

## Non-goals

- Extending the TOML parser grammar.
- Changing pathsec behavior.
- Patching gentle-ai itself (upstream's job; we file the issue).
