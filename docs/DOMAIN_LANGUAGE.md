# Domain Language — go-policy-dsl

Ubiquitous-language glossary for the **library-governance-policy** domain.
These terms are the vocabulary every consumer (`library-policy`,
`go-linter-sdk`, future CLIs/LSP servers) shares when reading and writing
policies. Code names and prose must stay honest against this glossary; if a
term here disagrees with code, **code wins** and this file gets corrected.

---

## Table of Contents

- [Core Concept](#core-concept)
- [Policy Attributes](#policy-attributes)
- [Detection](#detection)
- [Suppression & Content Gating](#suppression--content-gating)
- [Remediation](#remediation)
- [Companions](#companions)
- [Version Constraints](#version-constraints)
- [Validation & Errors](#validation--errors)
- [Domain Contracts](#domain-contracts)
- [Builder Method Convention](#builder-method-convention)
- [Constructor Helpers](#constructor-helpers)

---

## Core Concept

### Library Governance Policy

A declarative rule about how libraries may be used in a codebase. A policy
either **bans** a library, **requires a companion** alongside a library, or
both. The DSL declares _what_ a policy is; it does not execute, match, or
report — that is the consumer's job.

### PolicySpec

The finished, immutable Go value that represents a policy. A plain struct
with no behaviour — constructed only via the fluent `Builder` so every field
gets a sensible default. Fields:

| Field          | Type              | Meaning                                              |
| -------------- | ----------------- | ---------------------------------------------------- |
| `Name`         | `string`          | The library this policy targets                      |
| `Reason`       | `string`          | Human-readable justification (see [Reason](#reason)) |
| `Severity`     | `Severity`        | How serious a violation is                           |
| `Category`     | `Category`        | Why the policy exists (security, performance...)     |
| `Detection`    | `Detection`       | How a consumer discovers the policy applies          |
| `Description`  | `string`          | Detailed description for reporting                   |
| `Alternatives` | `[]Replacement`   | Recommended swap-in alternatives                     |
| `CVEs`         | `[]CVE`           | Related validated CVE identifiers                    |
| `VersionMin`   | `*Version`        | Inclusive lower version bound (`nil` = unbounded)    |
| `VersionMax`   | `*Version`        | Inclusive upper version bound (`nil` = unbounded)    |
| `Companions`   | `[]CompanionSpec` | Libraries that must be present alongside             |
| `Mode`         | `Mode`            | What the policy enforces (ban, companion-only)       |

In code: `PolicySpec` in `policy.go`. The zero-value `PolicySpec{}` is a valid
empty policy that [validates](#validate) clean — it is not a meaningful ban. Consumers
must check `Name`/`Reason`/`Detection` before reporting.

### Builder

The fluent entry point. Created via `Ban(name)`; every chainable method
returns the `*Builder`; `Spec()` terminates the chain and returns the
immutable `PolicySpec`. In code: `Builder` in `builder.go`.

### Reason

The human-readable justification for the policy, set via `Because(reason)`.
Required in spirit — a policy without a reason is unreviewable — but not
enforced by `Spec()` or `Validate()` (domain validation is the consumer's
job, by design).

---

## Policy Attributes

### Severity

Rates how serious a policy violation is. A `string` alias (NOT the
consumer's `finding.Severity`) so the DSL stays dependency-free. Consumers
bridge `Severity` to their own severity type at the boundary.

| Value                 | Meaning                                                            |
| --------------------- | ------------------------------------------------------------------ |
| `SeverityCritical`    | Blocks: production incident waiting to happen                      |
| `SeverityHigh`        | Strong warning: fix before merge unless explicitly justified       |
| `SeverityModerate`    | Recommendation worth acting on, not blocking                       |
| `SeverityLow`         | Informational (e.g. a newer version available, a style preference) |
| `SeverityInfo`        | Neutral observation (e.g. a detected companion that is present)    |
| `SeverityRecommended` | Non-blocking suggestion: a better-maintained alternative exists    |
| `SeverityDeprecated`  | Officially deprecated by maintainers; not immediately dangerous    |
| `SeverityObsolete`    | End-of-life or superseded; flag for removal planning               |

In code: `Severity` type and constants in `policy.go`. `Ban(name)` defaults
to `SeverityCritical`; `Companion(...)` defaults to `SeverityModerate`.

### Category

Classifies WHY a policy exists so consumers can filter and report by concern
("show me all security bans"). A `string` alias for the same dependency-free
reason as `Severity`.

| Value                   | Meaning                                          |
| ----------------------- | ------------------------------------------------ |
| `CategorySecurity`      | Security vulnerability or risk                   |
| `CategoryPerformance`   | Performance anti-pattern or regression           |
| `CategoryMaintenance`   | Maintainability concern (hard to maintain)       |
| `CategoryDeprecation`   | Library or pattern is deprecated                 |
| `CategoryArchitecture`  | Architectural concern (wrong layer, wrong shape) |
| `CategoryCorrectness`   | Correctness risk (silent data loss, race, etc.)  |
| `CategoryLicensing`     | Licensing incompatibility                        |
| `CategoryCompatibility` | Compatibility risk (breaking change, API drift)  |
| `CategoryConfiguration` | Configuration or operational concern             |

In code: `Category` type and constants in `policy.go`. `Ban(name)` defaults
to `CategorySecurity`; override with `WithCategory` for non-security concerns.

---

## Detection

### Detection

Declares how a consumer discovers that a policy applies. A policy may declare
source-level matches (import paths), manifest-level matches (go.mod module
paths), or both, plus suppression escape hatches and a content gate. In code:
`Detection` struct in `policy.go`.

> The DSL owns NO matching semantics: patterns are **opaque strings**, stored
> verbatim and never interpreted. Whether a pattern means a literal substring,
> a glob, or a regex is the consumer's decision. This contract is pinned by
> `FuzzBuilder_PatternsOpaque`. See [Opaque Patterns](#opaque-patterns).

### ImportPattern / ImportPatterns

**Source-level detection**: patterns matched against package paths in `.go`
source files.

- `ImportPattern(p)` — package-level convenience constructor returning a
  `Detection` with a single import pattern (the most common case for
  source-level bans).
- `ImportPatterns(p...)` — Builder method that **appends** patterns to the
  existing `Detection.ImportPatterns` slice.

### GoModPattern / GoModPatterns

**Manifest-level detection**: patterns matched against module paths in
`go.mod`.

- `GoModPattern(p)` — package-level convenience constructor returning a
  `Detection` with a single go.mod pattern.
- `GoModPatterns(p...)` — Builder method that **appends** patterns to the
  existing `Detection.GoModPatterns` slice.

### DetectVia

Builder method that **replaces** the entire `Detection` struct wholesale
(set semantics, unlike the append-style `ImportPatterns` / `GoModPatterns`).
Convenience over calling the individual append helpers when you want to
declare all detection fields at once.

---

## Suppression & Content Gating

### ExcludeIfContains

A **suppression list**: if any of these strings appears in the matched file
or `go.mod`, the violation is suppressed. This is the justified-use escape
hatch — "this library is present, but we have a documented reason that makes
it OK." The DSL stores the strings; the consumer owns the matching semantics.

Builder method `ExcludeIfContains(patterns...)` **appends** to
`Detection.ExcludeIfContains`.

### ExcludeIfTransitiveFrom

A **false-positive guard** for indirect dependencies: lists parent libraries
whose direct presence justifies this library appearing transitively. If a
listed parent pulls in the banned lib, no violation fires.

Use case: `ginkgo` pulls in `gomega` transitively. If `gomega` is banned but
`ginkgo` is an approved direct dependency, `ExcludeIfTransitiveFrom("ginkgo")`
prevents a false positive on `gomega`.

Builder method `ExcludeIfTransitiveFrom(libraries...)` **appends** to
`Detection.ExcludeIfTransitiveFrom`.

### RequireIfContains

A **content gate** — the inverse of `ExcludeIfContains`. When non-empty, the
policy only fires if at least one declared string appears in the matched file
content. `ExcludeIfContains` suppresses when a pattern IS present;
`RequireIfContains` activates ONLY when a pattern IS present.

Use case: `net/http` is both a client and server package.
`RequireIfContains("http.ServeMux", "http.ResponseWriter")` ensures a
policy only activates for server code, not client code. The DSL declares the
values; the consumer owns the matching semantics (case-insensitive substring
match in `library-policy`).

Builder method `RequireIfContains(patterns...)` **appends** to
`Detection.RequireIfContains`.

---

## Remediation

### Replacement

Recommends a single swap-in alternative for a banned library. Has `Library`
(module path) and `Reason` (why the replacement is better). Constructed via
`NewReplacement(library, reason)`. In code: `Replacement` struct in `policy.go`.

### Description

An optional detailed description for reporting, set via
`WithDescription(desc)`. Has a special interaction with `Suggest`: when
`Description` is empty at the time `Suggest` is called, it is auto-derived as
`"Replace with <library>: <reason>"`. An explicit `Description` set before or
after `Suggest` is never overwritten. Use `SuggestExplicit` to opt out of the
derivation entirely.

### Suggest

`Suggest(r Replacement)` appends the **full** `Replacement` (both `Library`
and `Reason`) to `Alternatives`, AND — when `Description` is empty — derives
`Description` from it (`"Replace with <library>: <reason>"`). This is
intentional and pinned by tests; an explicit `Description` set before or
after is never overwritten by a later `Suggest`.

### SuggestExplicit

`SuggestExplicit(r Replacement)` appends the replacement to `Alternatives`
WITHOUT deriving `Description`. It is the escape hatch for callers who find
the `Suggest` side-effect surprising or who want full control over
`Description`.

### Alternatives

The list of recommended replacements, typed `[]Replacement` so each entry
carries both `Library` and `Reason` (no information is discarded). Populated
by `Suggest` / `SuggestExplicit` (append) or replaced wholesale by
`WithAlternatives(...Replacement)` (set). A zero-value `PolicySpec` has `nil`
`Alternatives`.

### WithAlternativeStrings

`WithAlternativeStrings(libraries...)` is the string-convenience counterpart
to `WithAlternatives`: it wraps each bare module path in a `Replacement` with
an empty `Reason` and sets the slice wholesale. Use `Suggest` or
`SuggestExplicit` when you want to attach a reason to each alternative.

---

## Companions

### CompanionSpec

Declares a library that MUST be present alongside a chosen one (e.g.
`samber-do-auditlog` must accompany `samber/do`). Fields: `Library`,
`DetectionPattern`, `Reason`, and `Severity`. Constructed via:

- `Companion(lib, pattern, reason)` — defaults to `SeverityModerate`.
- `CompanionWithSeverity(lib, pattern, reason, sev)` — custom severity.

A zero-value `CompanionSpec{}` has an **empty** `Severity` (not the
`SeverityModerate` default that the `Companion` constructor sets). Consumers
bridging `Severity` must handle the empty case.

### RequiresCompanion

Builder method that **appends** a `CompanionSpec` to the policy's
`Companions` slice. Multiple companions can be required; each call appends.

### Mode

Declares what a policy enforces, as a typed enum on `PolicySpec`. It replaced
the former dishonest `CompanionOnly bool` field. Two values:

- **`ModeBan`** (default, set by `Ban(...)`) — the policy emits a ban finding
  for its target library and enforces any declared companions.
- **`ModeCompanionOnly`** (set by `AsCompanionOnly()`) — the policy suppresses
  the ban and enforces only that declared companions are present. Use this for
  "this library is fine, but if you use it you must also use X".

> **Deny-by-default:** the ONLY mode that suppresses the ban finding is
> `ModeCompanionOnly`. Every other value — the zero-value empty string,
> `ModeBan`, and any unknown/garbage string — is ban-active. A typo in `Mode`
> can never silently disable enforcement. See [Deny-by-Default](#deny-by-default).

> **History:** the field was `CompanionOnly bool`, a name that lied — it
> suppressed the _ban_, not the companion. The typed `Mode` enum is honest,
> reads as a question at call sites (`spec.Mode == ModeCompanionOnly`), and is
> extensible.

---

## Version Constraints

### Version

A parsed semantic version `{Major, Minor, Patch}` of the **library** targeted
by the policy (NOT the Go toolchain version). The zero `Version` is the valid
version `0.0.0`; an _unbounded_ range bound is represented by a `nil *Version`,
not by the zero value. Construct via:

- `NewVersion(major, minor, patch)` — rejects negatives, returns `(Version, error)`.
- `ParseVersion("1.2.3")` — strict `major.minor.patch` format; accepts a
  leading `v` (as in `v1.2.3`); rejects `1.2`, `1.2.3.4`, `latest`, `""`. Returns
  `(Version, error)`.

Both return errors; the library has no panic variants (see [Panic-Free](#panic-free)).
Ordered by `Compare` / `Before` / `After` / `Equal`. Uses stdlib `cmp.Compare`
(not hand-rolled sign logic, eliminating integer overflow edge cases).

In code: `Version` struct in `version.go`.

### VersionRange

`VersionRange(minVer, maxVer *Version)` is a Builder method that sets
inclusive `[min, max]` version constraints on the library targeted by the
policy. `nil` on either side means unbounded. An inverted range (`min > max`)
is **not** rejected at construction — the library never panics; detect it via
[Validate](#validate). Parse version strings with `ParseVersion` before the chain (there
is no string-convenience builder method, by panic-free design).

> **History:** `VersionRange` was renamed from `GoVersionRange` because the
> old name lied — it never constrained Go itself, only the library version.
> The fields were then retyped from `string` to `*Version` (v0.2.0).

### VersionMin / VersionMax

Fields on `PolicySpec` that store the inclusive version bounds. `nil` on
either side means unbounded. Set via `VersionRange`; readable directly on the
spec. An inverted range (`VersionMin > VersionMax`) is detectable via
`Validate()`.

---

## Validation & Errors

### Validate

`PolicySpec.Validate()` checks the structural invariants a spec must satisfy
regardless of domain. Currently the only such invariant is the version-range
ordering (`min <= max`). It does NOT enforce domain rules (a spec with no
`Reason`, no detection patterns, etc. still validates) — those are the
consumer's job, by design. `Spec()` itself remains validation-free and returns
exactly what was built.

Returns `nil` when the spec is structurally sound; returns
`*InvertedVersionRangeError` when the range is inverted. The return type is
the concrete error (not `error`), so callers using `:=` get type-safe access
to the offending bounds without `errors.As`.

### InvertedVersionRangeError

The structured error returned by `Validate()` when `VersionMin > VersionMax`.
Carries `Min` and `Max` pointers so callers can report the offending bounds
programmatically without parsing the error message. Implements `Is` so
`errors.Is(err, ErrInvertedVersionRange)` works. In code:
`InvertedVersionRangeError` in `policy.go`.

### ErrInvertedVersionRange

The sentinel error matched by `errors.Is` when `Validate` finds an inverted
range. `Validate` returns the concrete `*InvertedVersionRangeError` directly,
so callers get both the structured bounds AND sentinel matching.

### ErrInvalidCVE

Returned by `NewCVE(id)` when the string is not canonical `CVE-YYYY-NNNN`
format. Wraps the offending value in the error message.

### ErrInvalidVersion

Returned by `ParseVersion(s)` when the string is not exactly three
non-negative integer components separated by dots (e.g. rejects `1.2`,
`1.2.3.4`, `latest`, `""`). A leading `v` is accepted (`v1.2.3`).

### ErrNegativeVersion

Returned by `NewVersion(major, minor, patch)` when any component is
negative.

---

## Domain Contracts

These contracts are design invariants that shape the domain. They are
documented in `AGENTS.md`, pinned by tests, and must not be violated.

### Values, Not Config

`PolicySpec`, `Mode`, `CVE`, `Replacement`, and `Version` have **no**
`json`/`yaml` struct tags. The DSL declares _values_, not configuration. Adding
serialization tags would couple the public API to a specific wire format before
any consumer needs it. Consumers that need to marshal/unmarshal bridge at the
boundary with an explicit field-map. Revisit only if multiple consumers
independently request direct serialization AND agree on field names.

### No Execution Model

The DSL declares _what_ a policy is, not _how_ it is detected. Every consumer
reinvents matching semantics for `ImportPatterns` / `GoModPatterns` (literal?
glob? regex?). The DSL owns no matching semantics — patterns are opaque
strings, stored verbatim and never interpreted. See [Opaque Patterns](#opaque-patterns).

### Panic-Free

The library **never panics**. Every error condition — invalid CVE,
invalid/negative version, malformed version string, inverted version range —
is **returned** as a value, never panicked. There are deliberately no `Must*`
constructors (no `MustCVE`, `MustNewVersion`, `MustParseVersion`) and no
`VersionRangeStrings` convenience. Rationale: the fluent `Builder` chain
cannot propagate a returned error mid-chain, so any method that can fail lives
_outside_ the chain as a free function returning `(T, error)`. Pinned by
`TestNoPanicsInNonTestSource` (parses every non-test `.go` file via `go/parser`
and fails if any contains a `panic(` call) and `erraudit ./...` (0 violations).

### Deny-by-Default

The `Mode` typed enum has one design rule: **the ONLY value that suppresses
the ban finding is `ModeCompanionOnly`**. Every other value — the zero-value
empty string, `ModeBan`, and any unknown/garbage string — is ban-active. This
is a security property: a typo in `Mode` can never silently disable
enforcement. The worst case is an over-active ban, never a missed one.
`Validate()` does NOT reject unknown `Mode` values (unlike inverted version
ranges); it treats `Mode` validation as a domain rule left to the consumer.
Pinned by `TestMode_DenyByDefaultContract`.

### Opaque Patterns

The DSL owns NO matching semantics — detection patterns are opaque strings
stored verbatim and never interpreted (literal / glob / regex is the
consumer's decision). `FuzzBuilder_PatternsOpaque` pins this: any string
round-trips unchanged through every pattern entry point. See "Matching
semantics" in `ROADMAP.md` for raw ideas on optionally adding a matcher.

---

## Builder Method Convention

The `Builder` follows a naming convention enforced by tests and doc
comments, not by the type system. Contributors adding a new method should pick
the form that matches its semantics so the convention stays predictable at
call sites.

| Prefix / form                 | Semantics                                 | Examples                                                                                                                                                  |
| ----------------------------- | ----------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `With<X>(...)`                | **Set / replace** the field wholesale     | `WithSeverity`, `WithCategory`, `WithDescription`, `WithAlternatives(...Replacement)`, `WithAlternativeStrings(...string)`, `WithCVEs(...CVE)`            |
| Bare verb / noun (`<X>(...)`) | **Append** to a slice field               | `ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`, `ExcludeIfTransitiveFrom`, `RequireIfContains`, `RequiresCompanion`, `Suggest`, `SuggestExplicit` |
| `DetectVia(d)`                | **Replace** the whole `Detection` struct  | —                                                                                                                                                         |
| `As<X>()`                     | **Set** a mode flag                       | `AsCompanionOnly` (sets `ModeCompanionOnly`)                                                                                                              |
| `Spec()`                      | **Terminate** the chain, return the value | —                                                                                                                                                         |

> **Exceptions:** `Because(reason)` and `VersionRange(min, max)` are bare verbs
> that **set** (not append), breaking the "bare = append" pattern. They predate
> the convention and read more naturally as prose (`Ban("x").Because("...")`).
> New methods should follow the convention; these two are grandfathered.

---

## Constructor Helpers

Package-level functions (not methods) that construct typed values. They live
outside the fluent chain because they can return errors (see [Panic-Free](#panic-free)).

| Function                                                           | Returns            | Purpose                                                                                       |
| ------------------------------------------------------------------ | ------------------ | --------------------------------------------------------------------------------------------- |
| `Ban(name string) *Builder`                                        | `*Builder`         | Start a banned-library policy (defaults: `SeverityCritical` + `CategorySecurity` + `ModeBan`) |
| `Companion(lib, pattern, reason string) CompanionSpec`             | `CompanionSpec`    | Build a required-companion spec (defaults: `SeverityModerate`)                                |
| `CompanionWithSeverity(lib, pattern, reason string, sev Severity)` | `CompanionSpec`    | Companion with custom severity                                                                |
| `ImportPattern(pattern string) Detection`                          | `Detection`        | Convenience for source-import detection (single pattern)                                      |
| `GoModPattern(pattern string) Detection`                           | `Detection`        | Convenience for go.mod-path detection (single pattern)                                        |
| `NewReplacement(library, reason string) Replacement`               | `Replacement`      | Build a swap-in alternative value                                                             |
| `NewCVE(id string) (CVE, error)`                                   | `(CVE, error)`     | Validated CVE identifier (`CVE-YYYY-NNNN`); returns error on invalid                          |
| `NewVersion(major, minor, patch int) (Version, error)`             | `(Version, error)` | Parsed semver-lite value; rejects negatives                                                   |
| `ParseVersion(s string) (Version, error)`                          | `(Version, error)` | Parse a version string; returns error, never panics                                           |
