# TODO List — go-policy-dsl

Short- and mid-term actionable work, bounded and estimable. Each item cites
evidence (code path or source report). When an item ships, it is **removed**
from this file and recorded in `CHANGELOG.md` — completed work never lives
here. Vague or unbounded ideas belong in `ROADMAP.md`, not here.

> Source reports harvested: `docs/status/2026-07-26_07-09_full-review-self-critique.md`
> (P0–P4 lists), `docs/status/2026-07-19_01-38_sdk-release-and-buildflow-coverage-fix.md`
> (§f next-50 list). Items already shipped were dropped; items completed this
> session move to `CHANGELOG.md` `[Unreleased]`.

---

## Tests

### [T1] Composition test: `Ban(...).DetectVia(GoModPattern(...))`

Add a test exercising the common composition `Ban("x").DetectVia(GoModPattern("mod")).Spec()`
to pin that the convenience constructor composes with the replace-semantics
setter. No direct test exists today.
_Source: 2026-07-19 §f.11. Evidence: `policy_test.go` has no such case._

### [T2] Table-driven `Severity` / `Category` constants test

Pin the string values of every `Severity*` and `Category*` constant so an
accidental rename or value change is caught. Guards the public wire format
consumers bridge against.
_Source: 2026-07-19 §f.12. Evidence: no such table test exists._

### [T3] Fuzz tests for `Detection` pattern matching semantics

Once a consumer defines matching (literal / glob / regex), add Go 1.18+
`func Fuzz...` seed-corpora fuzz tests so the DSL's contract on what a
"pattern" is stays pinned. Depends on the matching-semantics decision in
`ROADMAP.md` ("matcher subpackage?").
_Source: 2026-07-26 §c.17._

---

## API surface (decisions, not yet code)

### [T4] Rename `CompanionOnly` to an honest name or typed `Mode`

`PolicySpec.CompanionOnly bool` is slightly dishonest — it suppresses the
**ban**, not the companion. Either rename to `SuppressBan` / `BanSuppressed`,
or introduce a typed `type Mode = Ban | CompanionOnly | Both` enum and drop
the boolean. Booleans-as-modes smell. Run the `naming-review` skill first.
_Source: 2026-07-26 §f.32–33. Evidence: `policy.go:126`._

### [T5] Typed `CVE` with `CVE-YYYY-NNNN` validation?

`CVEs []string` accepts any string. Decide whether to add a typed `CVE`
constructor that validates the CVE ID format, or keep `[]string` and let the
consumer validate. Keep stdlib-only.
_Source: 2026-07-26 §f.35. Evidence: `policy.go:116`, `builder.go:108`._

### [T6] Consolidate `Alternatives []string` vs `Replacement{Library, Reason}`?

Two ways to express "swap-in alternative" exist: the `Suggest(Replacement)`
append path and the `WithAlternatives([]string)` set path. Decide whether to
consolidate (e.g. `[]Replacement`) or keep both, and document the decision in
`docs/DOMAIN_LANGUAGE.md`.
_Source: 2026-07-26 §f.36. Evidence: `policy.go:115`, `builder.go:90-105`._

### [T7] `SuggestExplicit` variant (no Description side-effect)?

The `Suggest` → `Description` auto-derivation is convenient but surprising.
Decide whether to keep it as-is (current: feature, pinned by tests), document
harder, or add a `SuggestExplicit` that does NOT derive `Description`.
_Source: 2026-07-26 §f.37. Evidence: `builder.go:90-98`._

---

## Tooling / CI

### [T8] Wire `golangci-lint fmt --check` as a required CI gate

`golangci-lint fmt` is configured but not enforced as a check. Wire it as a
required gate (buildflow job or GitHub Action) so formatter drift never lands.
_Source: 2026-07-26 §f.21._

### [T9] Document the `Severity` ↔ `finding.Severity` bridge pattern

Add a README code sample showing how a consumer bridges `policydsl.Severity`
to their own severity type at the boundary. The design note exists but no
runnable example does.
_Source: 2026-07-26 §f.39. Evidence: `README.md` "Design notes"._

---

## Ecosystem (only after first consumer migrates)

### [T10] Add a "Consumers" section update once `library-policy` migrates

README "Consumers" lists `library-policy` as the primary consumer with
`domain/policy/spec.go` as the migration target. Update with the actual
migration commit/PR once it lands.
_Source: 2026-07-26 §f.40. Evidence: `README.md` "Consumers"._
