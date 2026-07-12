# Follow-up Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lock the TOML parser's rejection behavior with tests, fix the macOS-only pathsec test failure, and repair + report the wedged gentle-ai review store.

**Architecture:** Tasks 1–2 are test-only changes in existing `_test.go` files. Task 3 is repo maintenance (`.git/gentle-ai` state) plus an upstream GitHub issue; no picokit source changes.

**Tech Stack:** Go 1.26 stdlib. Spec: `docs/superpowers/specs/2026-07-12-followups-design.md`.

## Global Constraints

- Zero external dependencies (go.mod stays require-free).
- Parser behavior unchanged in Task 1: tests assert existing loud-error behavior; if a case parses silently, STOP and report (BLOCKED) — do not change the parser without escalating.
- Coverage floor 85% per package.
- Verify: `just check && go test ./coverage/ ./pathsec/` (full `just test` also passes once Task 2 lands — it removes the only known local failure).
- Conventional commits, no AI attribution.

---

### Task 1: Parser rejection tests (REL-1)

**Files:**
- Modify: `coverage/toml_test.go` (extend the `tests` table in `TestParseFloorsTOML`)

**Interfaces:**
- Consumes: `parseFloorsTOML(data []byte) (*Floors, error)` (unexported, same package).
- Produces: nothing new — regression lock only.

- [ ] **Step 1: Add the failing-input test cases**

Append to the `tests` table in `TestParseFloorsTOML` (after the existing error cases):

```go
		{
			name:    "multi-line array rejected",
			input:   "default = 85\nexclude = [\n\"pkg/a\",\n]\n",
			wantErr: "line 2",
		},
		{
			name:    "single-quoted string rejected",
			input:   "default = 85\nexclude = ['pkg/a']\n",
			wantErr: "line 2",
		},
		{
			name:    "unicode escape rejected",
			input:   "default = 85\nexclude = [\"\\u0041\"]\n",
			wantErr: "line 2",
		},
		{
			name:    "multi-line basic string rejected",
			input:   "default = 85\nexclude = [\"\"\"a\"\"\"]\n",
			wantErr: "line 2",
		},
```

Note on expectations: these inputs must produce an error mentioning the offending
line; the exact message is intentionally unpinned. Reasoning per case: line 2 of the
multi-line array is `"pkg/a",` — no `=` → "expected key = value"; `['pkg/a']` →
"expected quoted string"; `"A"` → `\u` is not `\"` or `\\` → "unsupported
escape sequence"; `"""a"""` → first `""` parses as an empty string, then `a` breaks
the array grammar → "expected ',' between array items".

- [ ] **Step 2: Run the new cases**

Run: `go test ./coverage/ -run TestParseFloorsTOML -v 2>&1 | tail -20`
Expected: all four new subtests PASS (the parser already rejects these inputs —
this is a regression lock, so "red" here would mean a parser bug). If any subtest
FAILS because `parseFloorsTOML` returned nil error, STOP: report BLOCKED with the
case name — the parser is silently accepting an unsupported form and the fix needs
design escalation, not an ad-hoc parser edit.

- [ ] **Step 3: Full package verification**

Run: `just check && go test ./coverage/`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add coverage/toml_test.go
git commit -m "test(coverage): lock parser rejection of unsupported TOML forms"
```

---

### Task 2: Fix pathsec test on macOS

**Files:**
- Modify: `pathsec/pathsec_test.go:273-280` (inside `TestEvalExistingPrefixRealPath`)

**Interfaces:**
- Consumes: `evalExistingPrefix(path string) (string, error)` (unexported, same package); `filepath.EvalSymlinks` (already imported via `path/filepath`).
- Produces: nothing new — test fix only.

- [ ] **Step 1: Confirm the current failure (red)**

Run: `go test ./pathsec/ -run TestEvalExistingPrefixRealPath`
Expected on macOS: FAIL with `result = "/private/var/..." want "/var/..."`.

- [ ] **Step 2: Canonicalize the expected value**

In `TestEvalExistingPrefixRealPath`, replace:

```go
	// Test with path that exists
	result, err := evalExistingPrefix(realDir)
	if err != nil {
		t.Fatalf("evalExistingPrefix error = %v", err)
	}
	if result != realDir {
		t.Fatalf("result = %q, want %q", result, realDir)
	}
```

with:

```go
	// evalExistingPrefix resolves symlinks, so the expectation must be
	// canonicalized too (macOS TMPDIR lives under the /var -> /private/var symlink).
	want, err := filepath.EvalSymlinks(realDir)
	if err != nil {
		t.Fatal(err)
	}

	// Test with path that exists
	result, err := evalExistingPrefix(realDir)
	if err != nil {
		t.Fatalf("evalExistingPrefix error = %v", err)
	}
	if result != want {
		t.Fatalf("result = %q, want %q", result, want)
	}
```

- [ ] **Step 3: Verify green**

Run: `go test ./pathsec/`
Expected: PASS (all pathsec tests, including the previously failing one).

- [ ] **Step 4: Full suite now green locally**

Run: `just check && just test`
Expected: PASS — this was the only known local failure.

- [ ] **Step 5: Commit**

```bash
git add pathsec/pathsec_test.go
git commit -m "test(pathsec): canonicalize expected path for macOS symlinked TMPDIR"
```

---

### Task 3: gentle-ai store repair + upstream issue

Orchestrator-executed (repo maintenance + outward-facing issue via the
issue-creation skill) — NOT dispatched to an implementer subagent.

**Files:**
- Delete: `.git/gentle-ai/review-transactions/v2/LOCK` (stale; owner PID 74679 verified dead)
- Move: `.git/gentle-ai/review-transactions/v2/review-f9ef23bf20861290/` and `.../review-6bc0797a22af4ce1/` → `.git/gentle-ai/review-transactions/archive-2026-07-12/`

**Interfaces:**
- Consumes: nothing from Tasks 1–2 except their commits existing (used to exercise the clean review cycle).
- Produces: working `gentle-ai review` lifecycle in this repo.

- [ ] **Step 1: Re-verify the lock is stale, then repair**

```bash
ps -p 74679 >/dev/null 2>&1 && echo "STOP: pid alive" || {
  mkdir -p .git/gentle-ai/review-transactions/archive-2026-07-12
  rm .git/gentle-ai/review-transactions/v2/LOCK
  mv .git/gentle-ai/review-transactions/v2/review-f9ef23bf20861290 \
     .git/gentle-ai/review-transactions/v2/review-6bc0797a22af4ce1 \
     .git/gentle-ai/review-transactions/archive-2026-07-12/
}
```

Expected: no "STOP" output; `v2/` left empty (or with only non-lineage metadata).

- [ ] **Step 2: Prove the lifecycle works end-to-end on the follow-up commits**

```bash
gentle-ai review start --cwd /Users/Shared/harness/picokit
```

Expected: fresh lineage, risk low or medium (test-only diff vs pushed master).
Then (running whatever lenses it selects; low tier selects none):

```bash
gentle-ai review finalize --evidence <evidence-file> --cwd /Users/Shared/harness/picokit
gentle-ai review validate --gate pre-commit --cwd /Users/Shared/harness/picokit
```

Expected: finalize → `"state": "approved"`; validate → `"allowed": true`.
The evidence file records the Task 1–2 test runs (commands + PASS results).

- [ ] **Step 3: File the upstream issue (issue-creation skill)**

Repo: `Gentleman-Programming/gentle-ai`. Title:
`review store wedges after scope-change: ambiguous lineage discovery returns empty-context scope-changed and start refuses new lineage`

Body must include: version 2.0.2 (brew, macOS); reproduction narrative (start →
finalize approved → untracked file hash changed → validate denies scope-changed →
start errors `transaction changed concurrently: expected compact revision ""` while
validate's suggested action is `create-new-lineage`; with two terminal lineages,
pre-push/pre-pr gates return `scope-changed` with an all-empty context block instead
of the `multiple compact facade review lineages found; specify --lineage` error that
pre-commit returns — and `validate` accepts `--lineage` to disambiguate even though
`--help` doesn't document it); stale-LOCK observation (dead PID not stolen; no
lock-expiry); what we expected (deterministic recovery path at gates). Note the
workaround used: delete stale LOCK + archive terminal lineages.

- [ ] **Step 4: Push picokit commits and confirm CI**

```bash
git push origin master
gh run list --branch master --limit 1
```

Expected: CI green (coverage-gate re-runs the parser end-to-end).

---

## Final verification

1. `just check && just test` — fully green locally (no pathsec failure).
2. CI green on master.
3. `gentle-ai review validate --gate pre-commit` allows.
4. Upstream issue URL captured in the session ledger.
