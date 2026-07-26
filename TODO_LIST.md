# TODO List — go-policy-dsl

Short- and mid-term actionable work, bounded and estimable. Each item cites
evidence (code path or source report). When an item ships, it is **removed**
from this file and recorded in `CHANGELOG.md` — completed work never lives
here. Vague or unbounded ideas belong in `ROADMAP.md`, not here.

> Source reports harvested: `docs/status/2026-07-26_07-09_full-review-self-critique.md`
> (P0–P4 lists), `docs/status/2026-07-19_01-38_sdk-release-and-buildflow-coverage-fix.md`
> (§f next-50 list). Items already shipped were dropped; their entries live in
> `CHANGELOG.md` `[Unreleased]`.

---

## Tests

### [T1] Fuzz tests for `Detection` pattern matching semantics

`FuzzParseVersion` covers the parser (the DSL's only pure-logic parsing
surface). Detection-pattern matching (`ImportPatterns` / `GoModPatterns`)
is deliberately left to consumers (see README "Design notes"), so a fuzz
target here depends on the matching-semantics decision in `ROADMAP.md`
("matcher subpackage?"). Add Go 1.18+ seed-corpora fuzz tests once a
consumer defines matching, so the DSL's contract on what a "pattern" is
stays pinned.
_Source: 2026-07-26 §c.17._

---

## API surface (decisions, not yet code)

### [T2] Rename `CompanionOnly` to an honest name or typed `Mode`

`PolicySpec.CompanionOnly bool` is slightly dishonest — it suppresses the
**ban**, not the companion. Either rename to `SuppressBan` / `BanSuppressed`,
or introduce a typed `type Mode = Ban | CompanionOnly | Both` enum and drop
the boolean. Booleans-as-modes smell. Run the `naming-review` skill first.
_Source: 2026-07-26 §f.32–33. Evidence: `policy.go:126`._

### [T3] Typed `CVE` with `CVE-YYYY-NNNN` validation?

`CVEs []string` accepts any string. Decide whether to add a typed `CVE`
constructor that validates the CVE ID format, or keep `[]string` and let the
consumer validate. Keep stdlib-only.
_Source: 2026-07-26 §f.35. Evidence: `policy.go:116`, `builder.go:108`._

### [T4] Consolidate `Alternatives []string` vs `Replacement{Library, Reason}`?

Two ways to express "swap-in alternative" exist: the `Suggest(Replacement)`
append path and the `WithAlternatives([]string)` set path. Decide whether to
consolidate (e.g. `[]Replacement`) or keep both, and document the decision in
`docs/DOMAIN_LANGUAGE.md`.
_Source: 2026-07-26 §f.36. Evidence: `policy.go:115`, `builder.go:90-105`._

### [T5] `SuggestExplicit` variant (no Description side-effect)?

The `Suggest` → `Description` auto-derivation is convenient but surprising.
Decide whether to keep it as-is (current: feature, pinned by tests), document
harder, or add a `SuggestExplicit` that does NOT derive `Description`.
_Source: 2026-07-26 §f.37. Evidence: `builder.go:90-98`._

### [T6] Consider a `Require(name)` builder

The doc lie I removed (a phantom `Require` builder referenced in package
docs) could become a real feature if the domain wants a distinct "required
library" policy type (not a ban, not a companion). Decide if in-scope; if
not, leave closed. This is a "consider" item — default is NOT to add it
unless a consumer needs it.
_Source: 2026-07-26 §f.28._

---

## Ecosystem (only after first consumer migrates)

### [T7] Add a "Consumers" section update once `library-policy` migrates

README "Consumers" lists `library-policy` as the primary consumer with
`domain/policy/spec.go` as the migration target. Update with the actual
migration commit/PR once it lands.
_Source: 2026-07-26 §f.40. Evidence: `README.md` "Consumers"._
