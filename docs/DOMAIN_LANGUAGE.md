# Domain Language — go-policy-dsl

Ubiquitous-language glossary for the library-governance-policy domain. These
terms are the vocabulary every consumer (`library-policy`, `go-linter-sdk`,
future CLIs/LSP servers) shares when reading and writing policies. Keep code
names and prose honest against this glossary; if a term here disagrees with
code, **code wins** and this file gets corrected.

---

## Core nouns

### Policy

A declarative rule about how libraries may be used in a codebase. A policy
either **bans** a library, **requires a companion** alongside a library, or
both. In code: `PolicySpec` (`policy.go:105`).

> The DSL declares _what_ a policy is. It does not execute, match, or report —
> that is the consumer's job.

### PolicySpec

The finished, immutable value produced by a `Builder`. A plain Go struct with
no behaviour. Constructed only via the fluent API so every field gets a
sensible default. In code: `PolicySpec` (`policy.go:105`).

### Builder

The fluent entry point. Returned by `Ban(name)` (ban a library) or built up
via `RequiresCompanion` / `AsCompanionOnly` for companion-only policies. Every
chainable method returns the `*Builder`; `Spec()` terminates the chain and
returns the immutable `PolicySpec`. In code: `Builder` (`builder.go:5`).

---

## Why a policy exists

### Reason

The human-readable justification for the policy. Required in spirit — a policy
without a reason is unreviewable — but not enforced by `Spec()` (validation is
the consumer's job). Set via `Because(reason)`.

### Severity

Rates how serious a violation is. A `string` alias (NOT the consumer's
`finding.Severity`) so the DSL stays dependency-free. Consumers bridge
`Severity` → their own severity type at the boundary.

| Value              | Meaning                                                         |
| ------------------ | --------------------------------------------------------------- |
| `SeverityCritical` | Blocks: production incident waiting to happen                   |
| `SeverityHigh`     | Strong warning: fix before merge unless explicitly justified    |
| `SeverityModerate` | Recommendation worth acting on, not blocking                    |
| `SeverityLow`      | Informational                                                   |
| `SeverityInfo`     | Neutral observation (e.g. a detected companion that is present) |

In code: `Severity` (`policy.go:26`) and the constants (`policy.go:28-49`).

### Category

Classifies WHY a policy exists so consumers can filter and report by concern
("show me all security bans"). A `string` alias. Values: `CategorySecurity`,
`CategoryPerformance`, `CategoryMaintainability`, `CategoryCorrectness`,
`CategoryLicensing`, `CategoryCompatibility`, `CategoryConfiguration`.

In code: `Category` (`policy.go:53`) and the constants (`policy.go:55-63`).

---

## How a violation is found

### Detection

Declares how a consumer discovers that a policy applies. A policy may declare
source-level matches (import paths), manifest-level matches (go.mod module
paths), or both, plus suppression escape hatches. In code: `Detection`
(`policy.go:67`).

> The DSL owns NO matching semantics: patterns are **opaque strings**, stored
> verbatim and never interpreted. Whether a pattern means a literal substring,
> a glob, or a regex is the consumer's decision. This contract is pinned by
> `FuzzBuilder_PatternsOpaque` (any string round-trips unchanged through every
> pattern entry point). See "Matching semantics" in `ROADMAP.md`.

### ImportPattern / ImportPatterns

Source-level detection: patterns matched against package paths in `.go` source
files. `ImportPattern(p)` is the single-pattern convenience constructor;
`ImportPatterns(p...)` is the append-style builder method.

### GoModPattern / GoModPatterns

Manifest-level detection: patterns matched against module paths in `go.mod`.
`GoModPattern(p)` is the single-pattern convenience constructor;
`GoModPatterns(p...)` is the append-style builder method.

### ExcludeIfContains

A suppression list: if any of these strings appears in the matched file or
`go.mod`, the violation is suppressed (the justified-use escape hatch).

### ExcludeIfTransitiveFrom

A false-positive guard for indirect dependencies: lists parent libraries whose
direct presence justifies this library appearing transitively. If a listed
parent pulls in the banned lib, no violation fires.

---

## What to do instead

### Replacement

Recommends a single swap-in alternative for a banned library. Has a `Library`
(module path) and a `Reason`. Constructed via `NewReplacement(library, reason)`.
In code: `Replacement` (`policy.go`).

### Suggest (the side-effect verb)

`Suggest(r Replacement)` appends the **full** `Replacement` (both `Library`
and `Reason`) to `Alternatives`, AND — when `Description` is empty — derives
`Description` from it (`"Replace with <library>: <reason>"`). This is
intentional and pinned by tests; an explicit `Description` set before/after is
never overwritten by a later `Suggest`.

### SuggestExplicit (the no-magic variant)

`SuggestExplicit(r Replacement)` appends the replacement to `Alternatives`
WITHOUT deriving `Description`. It is the escape hatch for callers who find the
`Suggest` side-effect surprising or who want full control over `Description`.

### Alternatives

The list of recommended replacements, typed `[]Replacement` so each entry
carries both `Library` and `Reason` (no information is discarded). Populated by
`Suggest` / `SuggestExplicit` (append) or replaced wholesale by
`WithAlternatives(...Replacement)` (set).

### CVE

A validated CVE identifier in canonical MITRE form `CVE-YYYY-NNNN...` (e.g.
`CVE-2021-44228`). A branded `string` type so an unvalidated free-form string
cannot reach a `PolicySpec`. Construct via `NewCVE(id)`, which validates the
form and returns an error on rejection; anything that is not `CVE-` + 4 digits +
`-` + 4+ digits is refused. Tagged onto a policy via `WithCVEs(...CVE)`. In
code: `CVE` (`cve.go`).

---

## Companions

### CompanionSpec

Declares a library that MUST be present alongside a chosen one (e.g.
`samber-do-auditlog` must accompany `samber/do`). Has `Library`,
`DetectionPattern`, `Reason`, and `Severity`. Constructed via `Companion(...)`
(defaults to `SeverityModerate`) or `CompanionWithSeverity(...)`.

### Mode

Declares what a policy enforces, on `PolicySpec` as a typed enum (it replaced
the former dishonest `CompanionOnly bool` field). Two values:

- `ModeBan` (default, set by `Ban(...)`) — the policy emits a ban finding for
  its target library and enforces any declared companions.
- `ModeCompanionOnly` (set by `AsCompanionOnly()`) — the policy suppresses the
  ban and enforces only that declared companions are present. Use this for
  "this library is fine, but if you use it you must also use X".

The zero value (empty `Mode`) is treated as ban-active, consistent with a
zero-value `PolicySpec`. In code: `Mode` (`policy.go`).

> **History:** the field was `CompanionOnly bool`, a name that lied — it
> suppressed the _ban_, not the companion. The typed `Mode` enum is honest,
> reads as a question at call sites (`spec.Mode == ModeCompanionOnly`), and is
> extensible. See `CHANGELOG.md` `[Unreleased]`.

---

## Version constraints

### Version

A parsed semantic version `{Major, Minor, Patch}` of the **library** targeted
by the policy (NOT the Go toolchain version). The zero `Version` is the valid
version `0.0.0`; an _unbounded_ range bound is represented by a `nil *Version`,
not by the zero value. Construct via `NewVersion(major, minor, patch)` or
`ParseVersion("1.2.3")` (both return errors; the library has no panic
variants). Ordered by `Compare` / `Before` / `After` / `Equal`.
In code: `Version` (`version.go`).

### VersionRange (builder method)

`VersionRange(minVer, maxVer *Version)` sets inclusive `[min, max]` version
constraints; `nil` on either side means unbounded. Parse version strings with
`ParseVersion` before the chain (the library is panic-free, so a string
convenience that hides a parse error behind a panic was removed). An inverted
range (`min > max`) is **not** rejected at construction — detect it via
`Validate()`. In code: `builder.go`.

### Validate (the structural check)

`PolicySpec.Validate() error` checks the structural invariants a spec must
satisfy regardless of domain. Currently the only such invariant is the
version-range ordering (`min <= max`). It does NOT enforce domain rules (a
spec with no `Reason`, no detection patterns, etc. still validates) — those
are the consumer's job, by design. `Spec()` itself remains validation-free
and returns exactly what was built.

> **History:** `VersionRange` was renamed from `GoVersionRange` (and
> `VersionMin`/`Max` from `GoVersionMin`/`Max`) because the old name lied —
> it never constrained Go itself, only the library version. The fields were
> then retyped from `string` to `*Version` (2026-07-26 review). The builder's
> inversion `panic` (and the `Must*` / `VersionRangeStrings` conveniences that
> panicked on error) were subsequently removed so the library returns errors
> exclusively; `Validate()` is now the single source of truth for the inversion
> invariant. See `CHANGELOG.md` `[Unreleased]`.

---

## Builder method convention

| Prefix / form                 | Semantics                                 | Examples                                                                                                                             |
| ----------------------------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `With<X>(...)`                | **Set / replace** the field wholesale     | `WithSeverity`, `WithCategory`, `WithDescription`, `WithAlternatives(...Replacement)`, `WithCVEs(...CVE)`                            |
| Bare verb / noun (`<X>(...)`) | **Append** to a slice field               | `ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`, `ExcludeIfTransitiveFrom`, `RequiresCompanion`, `Suggest`, `SuggestExplicit` |
| `DetectVia(d)`                | **Replace** the whole `Detection` struct  | —                                                                                                                                    |
| `As<X>()`                     | **Set the `Mode`**                        | `AsCompanionOnly` (sets `ModeCompanionOnly`)                                                                                         |
| `Spec()`                      | **Terminate** the chain, return the value | —                                                                                                                                    |

This convention is enforced only by tests and doc comments, not by the type
system. Contributors adding a new method should pick the form that matches its
semantics.
