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

| Value               | Meaning                                                        |
| ------------------- | -------------------------------------------------------------- |
| `SeverityCritical`  | Blocks: production incident waiting to happen                 |
| `SeverityHigh`      | Strong warning: fix before merge unless explicitly justified   |
| `SeverityModerate`  | Recommendation worth acting on, not blocking                  |
| `SeverityLow`       | Informational                                                 |
| `SeverityInfo`      | Neutral observation (e.g. a detected companion that is present) |

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

> The DSL defines the patterns; the consumer defines the matching semantics
> (literal substring, glob, regex). See "Open question: matching semantics"
> in `ROADMAP.md`.

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
In code: `Replacement` (`policy.go:85`).

### Suggest (the side-effect verb)

`Suggest(r Replacement)` does TWO things: appends `r.Library` to
`Alternatives` AND, if `Description` is empty, derives `Description` from the
replacement (`"Replace with <library>: <reason>"`). This is intentional and
pinned by tests; an explicit `Description` set before/after is never
overwritten by a later `Suggest`.

### Alternatives

The list of recommended replacement library names. Populated by `Suggest`
(append) or replaced wholesale by `WithAlternatives` (set).

---

## Companions

### CompanionSpec

Declares a library that MUST be present alongside a chosen one (e.g.
`samber-do-auditlog` must accompany `samber/do`). Has `Library`,
`DetectionPattern`, `Reason`, and `Severity`. Constructed via `Companion(...)`
(defaults to `SeverityModerate`) or `CompanionWithSeverity(...)`.

### CompanionOnly

A flag on `PolicySpec`: when true, the policy never emits a ban finding — it
only enforces that declared companions are present. Use this for "this library
is fine, but if you use it you must also use X". Set via `AsCompanionOnly()`.

> **Naming note (open):** `CompanionOnly` is slightly dishonest — it
> _suppresses the ban_, not the companion. A future rename to `SuppressBan`
> or a typed `Mode` enum is tracked in `TODO_LIST.md`.

---

## Version constraints

### VersionRange

Inclusive `[min, max]` version constraints on the **library** targeted by the
policy (NOT the Go toolchain version). Empty string on either side means
unbounded on that side. Set via `VersionRange(min, max)`.

> **History:** this was renamed from `GoVersionRange` (and `GoVersionMin` /
> `GoVersionMax`) because the old name lied — it never constrained Go itself,
> only the library version. See `CHANGELOG.md` `[Unreleased]`.

> **Open question:** `VersionMin` / `VersionMax` are `string`-typed, so the
> nonsensical inverted range `("2.0.0", "1.0.0")` is representable. A typed
> `Version` domain that rejects inversion at construction is sketched in
> `ROADMAP.md`.

---

## Builder method convention

| Prefix / form                | Semantics                                | Examples                                   |
| ---------------------------- | ---------------------------------------- | ------------------------------------------ |
| `With<X>(...)`               | **Set / replace** the field wholesale    | `WithSeverity`, `WithCategory`, `WithDescription`, `WithAlternatives`, `WithCVEs` |
| Bare verb / noun (`<X>(...)`) | **Append** to a slice field              | `ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`, `ExcludeIfTransitiveFrom`, `RequiresCompanion`, `Suggest` |
| `DetectVia(d)`               | **Replace** the whole `Detection` struct | —                                          |
| `As<X>()`                    | **Set a mode flag**                       | `AsCompanionOnly`                          |
| `Spec()`                     | **Terminate** the chain, return the value | —                                          |

This convention is enforced only by tests and doc comments, not by the type
system. Contributors adding a new method should pick the form that matches its
semantics.
