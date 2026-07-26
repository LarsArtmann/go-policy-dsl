# TODO List — go-policy-dsl

Short- and mid-term actionable work, bounded and estimable. Each item cites
evidence (code path or source report). When an item ships, it is **removed**
from this file and recorded in `CHANGELOG.md` — completed work never lives
here. Vague or unbounded ideas belong in `ROADMAP.md`, not here.

> **Harvest log:** items below were pulled from the 2026-07-26 status reports
> (`07-09`, `07-47`, `09-40`, `09-41`, `10-12`, `10-19`) and verified against
> the current code. Items already shipped were dropped; their entries live in
> `CHANGELOG.md` `[Unreleased]`. Decisions deferred to the user are marked
> **(decision needed)**.
>
> **Numbering:** `[T1]`–`[T6]` shipped in prior sessions (see `CHANGELOG.md`
> `[Unreleased]` — opaque-pattern fuzz, typed `Mode`, branded `CVE`,
> `[]Replacement`, `SuggestExplicit`, `Require` rejected). `[T7]` is the
> external-blocked consumer-migration item (kept at the bottom). New items
> continue from `[T8]`.

---

## Quick correctness wins (low effort, high clarity)

### [T8] Delete or wire up the dead `errMustCommit bool` field

`version_test.go:20` declares `errMustCommit bool` inside a test struct, but
the field is **never read** anywhere. It looks like an assertion that is
checking something but isn't — a latent lie. Either delete it, or wire it up to
actually assert error-commit behaviour.
_Evidence: `version_test.go:20`. Source: 2026-07-26 10-19 report §d.1, §f.2._

### [T9] Replace hand-rolled `cmpSign` with stdlib `cmp.Compare`

`version.go:122-125` defines `cmpSign(n int) int` to collapse a difference to
`-1/0/+1`. The stdlib `cmp.Compare` (Go 1.21+, this module targets 1.26.5) does
exactly this. Replacing it drops ~10 lines of hand-rolled sign logic and removes
an overflow edge case (`cmpSign` subtracts then signums; `cmp.Compare` compares
without subtracting, so it is correct for `MaxInt` vs `MinInt`).
_Evidence: `version.go:103-107` (three call sites), `version.go:122-125`.
Source: 2026-07-26 10-12 report §f.13._

### [T10] Move `parseVersionOrFatal` into a `testhelpers_test.go`

`parseVersionOrFatal` is defined in `builder_behavior_test.go:203` but used by
**both** `policy_test.go` and `builder_behavior_test.go`. Its location is
accidental — it lives in whichever test file was being edited when it was
written. A dedicated `testhelpers_test.go` is the honest home.
_Evidence: `builder_behavior_test.go:203` (definition); `policy_test.go:35`
(first external caller). Source: 2026-07-26 10-12 report §e.3, §f.11._

### [T11] Pin `Ban(...)` sets `Mode == ModeBan` with a test

The `Mode` typed enum was the headline change of its session, but no test
asserts that `Ban(...)` actually sets `spec.Mode == ModeBan`.
`TestBan_DefaultsToCriticalSecurity` checks `Severity`/`Category` but not
`Mode`. The rename is cosmetically honest but behaviourally under-pinned.
_Evidence: `policy_test.go` `TestBan_DefaultsToCriticalSecurity`. Source:
2026-07-26 09-41 report §b.1, §f.3._

### [T12] Add `TestNoPanicsInNonTestSource` regression test

The panic-free contract is a documented guarantee in `AGENTS.md` and
`erraudit ./...` verifies it at the tool level, but no in-repo test fails if
someone re-introduces a `panic(`. A trivial test that scans the package source
(non-test) for `panic(` prevents regression at PR time, independent of whether
`erraudit` is wired into CI.
_Evidence: `AGENTS.md` "Panic-Free" section. Source: 2026-07-26 10-12 report
§e.5, §f.3._

---

## Decisions needed (deferred to the user — shapes the API)

### [T13] Decide the unknown-`Mode` contract **(decision needed)**

`PolicySpec.Mode` is a typed enum, but `Mode("garbage")` compiles and
`Validate()` accepts it. Two honest options:

- **(a)** `Validate()` returns an error for any `Mode` not in
  `{ModeBan, ModeCompanionOnly, ""}` (make-bad-state-unrepresentable at the
  validate layer), or
- **(b)** document "any non-`ModeCompanionOnly` value (including unknown and
  `""`) is ban-active" as the contract and pin it with a test.

Today neither holds — the enum claims type safety it does not fully deliver at
the field-assignment boundary.
_Evidence: `policy.go` `Mode`, `Validate()`. Source: 2026-07-26 09-41 report
§b.2, §g.2._

### [T14] Decide the struct-tag policy **(decision needed)**

`PolicySpec`, `Mode`, `CVE`, `Replacement`, and `Version` have **no**
`json`/`yaml` struct tags (pure Go values, by design). `library-policy` emits
YAML and will need to marshal/unmarshal this struct at the cutover. Adding tags
now (before the first consumer migrates) lets consumers serialise directly;
leaving them off forces a manual field-map at the boundary (which the DSL's
"values, not config" philosophy arguably wants). This is a purity-vs-adoption
tradeoff that shapes the cutover.
_Evidence: `policy.go` `PolicySpec`, `version.go` `Version`. Source:
2026-07-26 07-47 report §f.3; 09-41 report §f.13, §g.3._

---

## CI & tooling

### [T15] Wire fuzz targets into CI

Two fuzz targets exist and pass their seeds (`FuzzParseVersion`,
`FuzzBuilder_PatternsOpaque`), but `.github/workflows/ci.yml` does not run them
with `-fuzztime`. The corpus only grows locally today.
_Evidence: `builder_fuzz_test.go`, `version_test.go`; `.github/workflows/ci.yml`.
Source: 2026-07-26 09-41 report §f.14._

### [T16] Add `dependabot.yml` for GitHub Actions version bumps

The CI pins `golangci-lint` and `setup-go` versions. A `dependabot.yml` keeps
them fresh automatically (closes the version-rot vector).
_Evidence: `.github/workflows/ci.yml`; no `.github/dependabot.yml`. Source:
2026-07-26 07-47 report §f.11._

---

## Ecosystem (blocked-on-external)

### [T7] Update "Consumers" section once `library-policy` migrates

README "Consumers" lists `library-policy` as the primary consumer and notes the
module is **not yet imported**. Update with the actual migration commit/PR once
it lands.

**Blocked-on-external (verified 2026-07-26):** `library-policy` does NOT yet
import `github.com/larsartmann/go-policy-dsl` — `go.mod` has no such dependency
and no `.go` file references the `policydsl.` package (checked directly). The
cutover is planned but unshipped.

This item **cannot** be completed until that cutover lands in `library-policy`.
Unblock trigger: a `library-policy` PR that adds
`require github.com/larsartmann/go-policy-dsl` to its `go.mod` and imports
`policydsl.Ban(...)`. At that point, also revisit whether any BREAKING change
(`Mode`, `[]Replacement`, `[]CVE`, panic-free removal) needs a coordinated
update on the consumer side.
_Evidence: `README.md` "Consumers"; `library-policy` `go.mod` +
`rg "github.com/larsartmann/go-policy-dsl" --include=*.go` = 0 hits._
