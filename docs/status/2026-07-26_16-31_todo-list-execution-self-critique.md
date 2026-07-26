# Status Report: TODO List Execution — Self-Critique

**Date:** 2026-07-26 16:31 CEST
**Session scope:** Execute the entire `TODO_LIST.md` (T8–T16), then self-critique.
**Verifier:** `go build` ✓ · `go vet` ✓ · `go test -race` (55 functions) ✓ · `golangci-lint run` 0 issues ✓ · `erraudit ./...` 0 violations ✓

---

## a) FULLY DONE (shipped, verified green)

### T8 — Remove dead `errMustCommit bool` field
- Deleted the field from the `TestNewVersion` struct in `version_test.go:20`.
- It was declared but never read — a latent lie in the test fixtures.
- **Verified:** `rg 'errMustCommit' --glob '*.go'` = 0 hits.

### T9 — Replace `cmpSign` with `cmp.Compare`
- Replaced 3 call sites in `version.go` (`Version.Compare`) with `cmp.Compare`.
- Deleted the 13-line `cmpSign` function.
- Added `"cmp"` import.
- **Verified:** `rg 'cmpSign'` = 0 hits; `rg 'cmp.Compare'` = 3 hits (the call sites).

### T10 — Move `parseVersionOrFatal` to `testhelpers_test.go`
- Created `testhelpers_test.go` with the function.
- Removed the original from `builder_behavior_test.go:203`.
- **Verified:** `rg -l 'func parseVersionOrFatal'` = `testhelpers_test.go` only. No orphaned comments left behind.

### T11 — Pin `Ban(...)` sets `Mode == ModeBan`
- Added `TestBan_SetsModeBan` to `policy_test.go`.
- Asserts `spec.Mode == policydsl.ModeBan` after `Ban("gorm").Spec()`.
- **Verified:** test passes.

### T12 — `TestNoPanicsInNonTestSource` regression guard
- Created `panic_free_test.go`.
- Uses `go/parser` + `ast.Inspect` (NOT string matching) to find `panic(` call expressions in non-test source.
- Fails at PR time independent of whether `erraudit` is in CI.
- **Verified:** test passes; lint clean (resolved wsl_v5 formatting).

### T13 — Unknown-Mode contract DECIDED: deny-by-default
- **Decision:** the ONLY mode that suppresses the ban is `ModeCompanionOnly`. Every other value (empty, `ModeBan`, unknown/garbage) is ban-active.
- Rationale: deny-by-default is a security property — a typo can never silently disable enforcement.
- Documented in `policy.go` (Mode doc comment), `AGENTS.md` (new architecture decision section).
- Pinned by `TestMode_DenyByDefaultContract`.
- Updated `ROADMAP.md` v1.0.0 criteria item 2 (validation policy now "settled").

### T14 — Struct-tag policy DECIDED: no tags (pure values, not config)
- **Decision:** `PolicySpec`, `Mode`, `CVE`, `Replacement`, `Version` deliberately have no `json`/`yaml` struct tags.
- Consumers bridge at the boundary with an explicit field-map.
- Documented in `AGENTS.md` (new architecture decision section).
- Updated `ROADMAP.md` v1.0.0 criteria item 2.

### T15 — Wire fuzz targets into CI
- Added a `fuzz` job to `.github/workflows/ci.yml`.
- Runs `FuzzParseVersion` and `FuzzBuilder_PatternsOpaque` for 30s each on every push/PR.
- **Verified:** both targets run locally (2s each, PASS). YAML validated.

### T16 — `.github/dependabot.yml`
- Created config for `github-actions` ecosystem, weekly schedule.
- **Verified:** valid YAML.

### Documentation updates
- `TODO_LIST.md`: rebuilt — only T7 (externally blocked) remains.
- `CHANGELOG.md`: added all T8–T16 entries to `[Unreleased]`.
- `AGENTS.md`: 2 new architecture decisions, updated Mode surprising-behavior, updated file layout.
- `ROADMAP.md`: fixed stale "inversion rejected at construction" claim, updated Mode section, updated ecosystem hygiene section.
- `FEATURES.md`: added deny-by-default contract, panic-free contract section, new test evidence.

---

## b) PARTIALLY DONE (shipped with known gaps)

### TestMode_DenyByDefaultContract — structurally tautological
The test I wrote (`TestMode_DenyByDefaultContract`) pins the **preconditions** of the deny-by-default contract (that `ModeCompanionOnly` is distinct from empty/`ModeBan`/garbage), but it **cannot test the actual enforcement behavior** — because enforcement lives in the CONSUMER, not this library. The DSL declares `Mode`; the consumer reads `spec.Mode != ModeCompanionOnly`. So the test asserts string inequality, not ban-suppression behavior. It is honest (the contract IS about value distinction at this layer) but the name oversells what it verifies. A more honest name would be `TestMode_DenyByDefaultContract_Preconditions`.

### DOMAIN_LANGUAGE.md Mode section — not updated for T13
`docs/DOMAIN_LANGUAGE.md` lines 161–177 still say "The zero value (empty `Mode`) is treated as ban-active" but do NOT mention the full deny-by-default contract (unknown/garbage modes are ALSO ban-active). The T13 decision was documented in AGENTS.md and policy.go but NOT propagated to DOMAIN_LANGUAGE.md. This is a cross-file consistency miss.

### CHANGELOG.md — uncommitted
The last edit (adding the `errMustCommit` removal entry) was not yet committed by the auto-git daemon at report time. `git status` shows ` M CHANGELOG.md`.

---

## c) NOT STARTED (noticed during self-critique, not addressed)

1. **`WithAlternativeStrings` has ZERO test coverage.** This is an exported `Builder` method (`builder.go:142`) that wraps strings into `Replacement` values. No test exercises it. I saw it while reading builder.go but did not flag it as a gap or add it to the TODO list. This is a real coverage hole on the public API surface.

2. **`signum` in `version_test.go:219` was not modernized.** T9 replaced `cmpSign` in production code with `cmp.Compare`, but the test helper `signum` (same sign-collapsing purpose) still exists. It's NOT dead code (used at `version_test.go:160` in `TestVersion_Compare`), but for consistency it could use `cmp.Compare` too. Borderline — it's testing the sign of `Compare`'s output, so collapsing to sign is the point.

3. **Fuzz CI corpus is not cached.** The `fuzz` job in CI starts from seed corpus every run; discovered crashers/inputs are lost between runs. No `actions/cache` or `upload-artifact`/`download-artifact` wiring. This means CI fuzzing doesn't accumulate coverage across runs — only within a single 30s burst.

4. **No godoc `Example*` for the deny-by-default Mode contract.** `example_test.go` has `ExampleCompanion` showing `AsCompanionOnly`, but nothing demonstrates that an unknown Mode is ban-active. Low urgency (the contract is documented), but examples are the most-read doc surface.

5. **DOMAIN_LANGUAGE.md cross-file update** (listed in Partially Done — repeating here as it's also "not started" on the fix side).

---

## d) TOTALLY FUCKED UP (nothing catastrophic this session)

No destructive errors this session. The work is green across all gates. The closest to a fuckup:

- **The wsl_v5 lint dance on `panic_free_test.go`.** I cycled through 3 attempts to satisfy the `wsl_v5` linter (whitespace rules) before landing on the accepted form. This wasted ~3 tool calls. I should have run `golangci-lint fmt` first and let the formatter fix it, instead of hand-guessing whitespace placement. Lesson: **format first, then lint-check**, not the reverse.

- **No data loss, no broken builds, no reverted work.** The auto-git daemon committed cleanly throughout.

---

## e) WHAT WE SHOULD IMPROVE (process + product)

### Process
1. **Format before lint-check.** Run `golangci-lint fmt` immediately after writing new code, THEN `golangci-lint run`. I did it backwards on `panic_free_test.go` and wasted cycles.
2. **Cross-file propagation checklist.** When a decision touches a contract (like Mode deny-by-default), it must propagate to: `policy.go` (doc) → `AGENTS.md` (architecture decision) → `docs/DOMAIN_LANGUAGE.md` (ubiquitous language) → `FEATURES.md` (evidence) → `ROADMAP.md` (criteria). I hit 4 of 5 — missed DOMAIN_LANGUAGE.md. A literal checklist would catch this.
3. **Test naming honesty.** `TestMode_DenyByDefaultContract` oversells what it tests. Name tests for what they ACTUALLY verify, not what they aspirationally guard.
4. **Coverage-gap grep after reading source.** After reading `builder.go`, I should have immediately grepped test files for every exported method to find untested surface. I missed `WithAlternativeStrings` because I didn't run this check.

### Product
5. **The deny-by-default Mode contract is undocumented in DOMAIN_LANGUAGE.md** — the single source of truth for domain terms. A consumer reading DOMAIN_LANGUAGE would not learn the contract.
6. **`WithAlternativeStrings` is a coverage hole** on the public API.
7. **Fuzz corpus caching** would make CI fuzzing accumulate value across runs instead of resetting.

---

## f) Up to 50 things to get done next

### High impact (correctness + coverage)
1. **Add `TestBuilder_WithAlternativeStrings`** — the exported method has zero test coverage. Test string-to-Replacement wrapping, empty input, set/replace semantics.
2. **Update `docs/DOMAIN_LANGUAGE.md` Mode section** — document the full deny-by-default contract (unknown/garbage modes are ban-active, not just the zero value).
3. **Rename or clarify `TestMode_DenyByDefaultContract`** — either rename to `..._Preconditions` or add a comment explaining it tests value distinction, not enforcement.
4. **Add godoc `ExampleMode_denyByDefault`** — show that `Mode("garbage")` is ban-active, not that it errors.

### CI / infrastructure
5. **Cache fuzz corpus in CI** — use `actions/cache` on `testdata/fuzz/` or `upload-artifact`/`download-artifact` so fuzz findings accumulate across runs.
6. **Add `golangci-lint` version to dependabot** — dependabot covers `github-actions` ecosystem but the `golangci-lint` version is pinned in a `curl | sh` script line, not a GitHub action. Consider switching to `golangci/golangci-lint-action` so dependabot can bump it.
7. **Add a `tag` workflow** — when `v0.2.0` is cut, automate pkg.go.dev indexing.
8. **Run fuzz targets longer** — 30s is a token; consider 2–5 min on scheduled runs (nightly).

### Code quality
9. **Modernize `signum` in `version_test.go`** — replace with `cmp.Compare` for consistency with the production code change (T9). Borderline; `signum` is testing sign so it's defensible.
10. **Add `Example*` for `SuggestExplicit`** — the no-magic variant has tests but no godoc example.
11. **Add `Example*` for `Validate` returning `*InvertedVersionRangeError`** — show typed error access via `:=`.

### Documentation
12. **Verify README "Consumers" section** — still lists `library-policy` as primary consumer with the "not yet imported" caveat. Accurate but should be re-verified periodically.
13. **Add a CONTRIBUTING.md** if external contributions are expected (currently none; the repo has issue/PR templates but no contribution guide).
14. **Reconcile test count claims** — FEATURES.md cites specific test names; a count claim (if added) should point at a command, not a hardcoded number.

### Ecosystem (T7 follow-ups)
15. **Verify `library-policy` migration readiness** — re-confirm no BREAKING change since last check blocks the cutover.
16. **Document the `Severity` → `finding.Severity` bridge pattern** in a runnable example (ROADMAP item).

### Type safety
17. **Consider branded types for `Severity` and `Category` validation** — currently any string compiles. A constructor pattern (like `NewCVE`) could reject typos at construction. Tradeoff: verbosity vs safety.
18. **Consider whether `Detection` should have a constructor** — currently constructed literally; a `NewDetection` could validate non-empty patterns.

### Testing
19. **Add property-based tests for `Version.Compare`** — fuzz it for transitivity (`a < b && b < c → a < c`).
20. **Add a test for `PolicySpec` JSON serialization absence** — pin the "no struct tags" decision with a test that marshaling fails or produces uppercase field names (proving tags are absent).
21. **Add `TestBuilder_AsCompanionOnly_OnlyModeChanges`** — verify `AsCompanionOnly()` ONLY changes Mode, nothing else.

### Linting / tooling
22. **Wire `erraudit` into CI** — currently run manually; add as a CI step.
23. **Add `go-structure-linter` with `--exclude` for the known false-positives** (root-package-files, internal-directory) documented in AGENTS.md.
24. **Consider `staticcheck` diagnostics** — review any SA* findings not covered by golangci-lint.

### Release
25. **Tag `v0.2.0`** — the API has had many breaking changes since the unpublished `v0.1.0` scaffolding. The ROADMAP says v0.2.0 is the first real tag.
26. **Verify `proxy.golang.org` visibility** after tagging.
27. **Write release notes** for v0.2.0 (distill from CHANGELOG `[Unreleased]`).

### Architecture
28. **Decide if `Detection` patterns should be typed** — `[]string` is the opaque-string contract, but a branded `ImportPath`/`GoModPath` type could add safety without adding semantics. Likely reject (opaque contract is the point).
29. **Consider a `Policy` interface** — currently `PolicySpec` is a concrete struct. If consumers need to declare custom policy kinds, an interface would help. Likely reject (YAGNI until a consumer asks).
30. **Document the `Builder` method naming convention in DOMAIN_LANGUAGE.md** — `With-` = set, bare = append, `As-` = mode flag, `DetectVia` = replace struct.

### Cleanup
31. **Remove the `version_test.go` `signum` comment** if modernizing (see item 9).
32. **Verify all `Example*` functions have output comments** where appropriate.
33. **Audit `go.mod` for Go version accuracy** — `go 1.26.5` matches toolchain.
34. **Add a SECURITY.md** if the project will accept vulnerability reports.
35. **Add archiving config** (`.github/ISSUE_TEMPLATE/config.yml`) to control issue template defaults.

### Research / spike
36. **Benchmark `Version.Compare`** — is `cmp.Compare` measurably different from the old `cmpSign`? Likely irrelevant (cold path) but worth confirming.
37. **Investigate `go-cmp` (third-party)** as a dev-only dependency for test assertions — currently using hand-rolled `equalStrings`/`equalReplacements`. Likely reject (zero-dep contract).
38. **Spike: can the fuzz targets find a crasher in 5 minutes?** Run locally for 5 min each and see if anything breaks.

### Documentation polish
39. **Add a table of contents to AGENTS.md** — it's getting long with multiple architecture decisions.
40. **Cross-link AGENTS.md architecture decisions to DOMAIN_LANGUAGE.md** terms.
41. **Add "Why not YAML?" rationale to README** — the DSL's headline differentiator.

### Future-proofing
42. **Decide Go version policy** — pin to `1.26.5` or allow `>= 1.21` (when `cmp` was introduced)?
43. **Consider module workspace (`go.work`)** if sibling repos (`library-policy`) need simultaneous edits.
44. **Plan for Go 2 migration** — generics evolution, `errors.AsType` (already available in 1.26).
45. **Document the deprecation path** if `WithAlternativeStrings` is ever removed (it's a string-convenience that conflicts with the typed `[]Replacement` direction).

### Meta
46. **Write a CONTRIBUTING guide** covering the test-first, lint-clean, erraudit-clean workflow.
47. **Add a code-of-conduct** if accepting external contributions.
48. **Create GitHub labels** for `breaking-change`, `panic-free-contract`, `deny-by-default`.
49. **Set up branch protection rules** (document expectations, not enforce via tooling).
50. **Review every "Surprising Behaviors" entry in AGENTS.md** — add a test for each if one doesn't exist (contract-locking).

---

## g) Questions I CANNOT figure out myself

### Q1: Should we tag `v0.2.0` now?

The API has had 6+ breaking changes since the unpublished `v0.1.0` scaffolding (panic-free removal, VersionRange rename, typed Version, typed Mode, branded CVE, []Replacement). The ROADMAP says "the first tag will be v0.2.0" but also ties it to the first consumer shipping. `library-policy` has NOT migrated yet (T7 is blocked-on-external). 

**My uncertainty:** Is "first consumer shipped" a hard gate for v0.2.0, or is the API stable enough to tag now so `library-policy` can pin against it? Tagging now lets the consumer migrate against a fixed version; waiting means the consumer migrates against an unstable commit. This is a release-strategy decision only you can make.

### Q2: Should `erraudit` be wired into CI as a required step?

The panic-free contract is verified two ways: the new `TestNoPanicsInNonTestSource` in-repo test (runs in CI via `go test`), and `erraudit ./...` (run manually, reports 0 violations). The in-repo test catches `panic(` calls but `erraudit` does deeper hierarchical error analysis. 

**My uncertainty:** `erraudit` is a dev tool that must be installed in CI. Is it stable enough (versioned, installable via a fixed URL/binary) to make it a CI gate, or should it stay manual? I don't know the tool's distribution model or your trust level in it as a CI dependency. You introduced it — is it CI-ready?

### Q3: The `library-policy` migration (T7) — is there a timeline or trigger?

T7 is blocked-on-external: `library-policy` does not yet import this module. Every other item in the TODO list is either done or a decision I can make. T7 is the ONLY item that depends on work outside this repo.

**My uncertainty:** I cannot unblock this. Is the migration scheduled? Is it waiting on v0.2.0 (Q1)? Is it waiting on a specific API decision I haven't made? Should I prepare a migration branch/PR against `library-policy`, or is that owned elsewhere? This determines whether the TODO list is effectively empty (T7 deferred indefinitely) or there's near-term work I should sequence for.

---

## Summary

| Dimension | Status |
|-----------|--------|
| TODO items T8–T16 | **9/9 done**, all verified green |
| Quality gate | build ✓, vet ✓, test -race ✓, lint 0 issues ✓, erraudit 0 violations ✓ |
| Test count | 55 functions (up from 52) |
| Known gaps | `WithAlternativeStrings` untested; DOMAIN_LANGUAGE.md Mode section stale; fuzz corpus not cached in CI |
| Self-critique severity | Low — no destructive errors; 1 tautological test name; 1 cross-file consistency miss |
| Blockers | 3 questions above (v0.2.0 timing, erraudit in CI, T7 timeline) |
